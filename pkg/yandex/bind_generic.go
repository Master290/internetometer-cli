// havent tested it yet
// probably some bs 

package yandex

import (
	"context"
	"net"
	"time"
)

func dialContext(ctx context.Context, network, addr, ifaceName string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	iface, err := net.InterfaceByName(ifaceName)
	if err == nil {
		addrs, err := iface.Addrs()
		if err == nil {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				host = addr
			}

		
			ips, _ := net.LookupIP(host)
			
			isIPv6 := false
			if len(ips) > 0 {
				for _, ip := range ips {
					if ip.To4() == nil {
						isIPv6 = true
						break
					}
				}
			}

			var localIP net.IP
			for _, a := range addrs {
				var ip net.IP
				switch v := a.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip == nil || ip.IsLoopback() {
					continue
				}
				if isIPv6 && ip.To4() == nil {
					localIP = ip
					break
				}
				if !isIPv6 && ip.To4() != nil {
					localIP = ip
					break
				}
			}

			if localIP == nil {
				for _, a := range addrs {
					var ip net.IP
					switch v := a.(type) {
					case *net.IPNet:
						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}
					if ip == nil || ip.IsLoopback() {
						continue
					}
					localIP = ip
					break
				}
			}

			if localIP != nil {
				dialer.LocalAddr = &net.TCPAddr{IP: localIP}
			}
		}
	}

	return dialer.DialContext(ctx, network, addr)
}
