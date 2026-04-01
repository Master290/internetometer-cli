//go:build !linux

package yandex

import "net/http"

func interfaceTransport(_ string) http.RoundTripper {
	return nil
}
