//go:build linux
// +build linux

package yandex

import (
	"context"
	"net"
	"syscall"
	"time"
)

func dialContext(ctx context.Context, network, addr, iface string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, iface)
			})
		},
	}
	return dialer.DialContext(ctx, network, addr)
}
