//go:build linux

package yandex

import (
	"net"
	"net/http"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func interfaceTransport(iface string) http.RoundTripper {
	if iface == "" {
		return nil
	}
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			var sockErr error
			err := c.Control(func(fd uintptr) {
				sockErr = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, iface)
			})
			if err != nil {
				return err
			}
			return sockErr
		},
	}
	return &http.Transport{
		DialContext: dialer.DialContext,
	}
}
