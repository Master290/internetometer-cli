package yandex

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type model struct {
	client *Client
	ctx    context.Context
	cancel context.CancelFunc

	ipv4   string
	ipv6   string
	region string
	isp    string

	downloadPrg progress.Model
	uploadPrg   progress.Model
	spinner     spinner.Model

	downloadMbps float64
	uploadMbps   float64
	curMbps      float64
	latency      string

	phase          string
	phaseStartTime time.Time
	phasePercent   float64
	err            error
}

const phaseTotalTime = 8.0

type progressMsg ProgressReport
type resultMsg struct {
	res *SpeedResult
	err error
}

var program *tea.Program

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Cmd {
			return func() tea.Msg {
				ipv4, _ := m.client.GetIPv4()
				ipv6, _ := m.client.GetIPv6()
				region, _ := m.client.GetRegion()
				isp, _ := m.client.GetISP()
				return initialInfoMsg{ipv4, ipv6, region, isp}
			}
		}(),
	)
}

type initialInfoMsg struct {
	ipv4, ipv6, region string
	isp                *ISPInfo
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			m.cancel()
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case initialInfoMsg:
		m.ipv4 = msg.ipv4
		m.ipv6 = msg.ipv6
		m.region = msg.region
		if msg.isp != nil {
			m.isp = msg.isp.Name
		}
		m.phase = "download"
		m.phaseStartTime = time.Now()
		m.phasePercent = 0
		return m, m.runSpeedTestCmd(program)
	case progressMsg:
		if msg.IsDownload {
		}
		return m, nil
	case currentMbpsMsg:
		m.curMbps = float64(msg)
		if !m.phaseStartTime.IsZero() {
			elapsed := time.Since(m.phaseStartTime).Seconds()
			m.phasePercent = elapsed / phaseTotalTime
			if m.phasePercent > 1.0 {
				m.phasePercent = 1.0
			}
		}
		return m, nil
	case phaseMsg:
		m.phase = string(msg)
		m.curMbps = 0
		m.phaseStartTime = time.Now()
		m.phasePercent = 0
		return m, nil
	case resultMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.downloadMbps = msg.res.DownloadMbps
			m.uploadMbps = msg.res.UploadMbps
			m.latency = fmt.Sprintf("%d ms", msg.res.Latency.Milliseconds())
		}
		m.phase = "done"
		return m, tea.Quit
	}
	return m, nil
}

type currentMbpsMsg float64
type phaseMsg string

func (m model) runSpeedTestCmd(p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		if p == nil {
			return nil
		}

		var mu sync.Mutex
		currentIsDownload := true
		var lastUpdate time.Time
		var phaseStartTime time.Time

		progress := func(pReport ProgressReport) {
			mu.Lock()
			defer mu.Unlock()

			if phaseStartTime.IsZero() && pReport.Bytes > 0 {
				phaseStartTime = time.Now()
			}

			if pReport.IsDownload != currentIsDownload {
				currentIsDownload = pReport.IsDownload
				phaseStartTime = time.Now()
				phase := "download"
				if !currentIsDownload {
					phase = "upload"
				}
				p.Send(phaseMsg(phase))
			}

			now := time.Now()
			if !phaseStartTime.IsZero() && now.Sub(lastUpdate) > 100*time.Millisecond {
				duration := now.Sub(phaseStartTime).Seconds()
				if duration > 0.01 {
					mbps := (float64(pReport.Bytes) * 8) / (duration * 1000000.0)
					p.Send(currentMbpsMsg(mbps))
					lastUpdate = now
				}
			}
		}
		res, err := m.client.RunSpeedTest(m.ctx, progress)
		return resultMsg{res, err}
	}
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var s strings.Builder
	s.WriteString(titleStyle.Render("Yandex Internetometer CLI"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("IPv4:   %s\n", m.ipv4))
	if m.ipv6 != "" {
		s.WriteString(fmt.Sprintf("IPv6:   %s\n", m.ipv6))
	}
	s.WriteString(fmt.Sprintf("Region: %s\n", m.region))
	s.WriteString(fmt.Sprintf("ISP:    %s\n", m.isp))
	s.WriteString("\n")

	switch m.phase {
	case "init":
		s.WriteString(m.spinner.View() + " Gathering information...")
	case "download":
		s.WriteString(m.spinner.View() + fmt.Sprintf(" Measuring Download: %.2f Mbps\n", m.curMbps))
		s.WriteString(m.renderBar())
	case "upload":
		s.WriteString(m.spinner.View() + fmt.Sprintf(" Measuring Upload:   %.2f Mbps\n", m.curMbps))
		s.WriteString(m.renderBar())
	case "done":
		s.WriteString(keywordStyle.Render("Results:"))
		s.WriteString(fmt.Sprintf("\nDownload: %.2f Mbps", m.downloadMbps))
		s.WriteString(fmt.Sprintf("\nUpload:   %.2f Mbps", m.uploadMbps))
		s.WriteString(fmt.Sprintf("\nLatency:  %s", m.latency))
	}

	s.WriteString("\n\n" + infoStyle.Render("Press q or Ctrl+C to quit"))
	return s.String()
}

func (m model) renderBar() string {
	return m.downloadPrg.ViewAs(m.phasePercent)
}

func RunTUI(client *Client) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := model{
		client:      client,
		ctx:         ctx,
		cancel:      cancel,
		spinner:     s,
		downloadPrg: progress.New(progress.WithGradient("#FF10FF", "#10FFFF"), progress.WithWidth(40)),
		phase:       "init",
	}

	p := tea.NewProgram(m)
	program = p
	_, err := p.Run()
	return err
}
