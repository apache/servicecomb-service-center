package utils

import (
	"fmt"
	"net"
)

func SplitHostPort(address string, defaultPort int) (string, int, error) {
	_, _, err := net.SplitHostPort(address)
	if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
		address = fmt.Sprintf("%s:%d", address, defaultPort)
		_, _, err = net.SplitHostPort(address)
	}
	if err != nil {
		return "", 0, err
	}

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return "", 0, err
	}

	return addr.IP.String(), addr.Port, nil
}
