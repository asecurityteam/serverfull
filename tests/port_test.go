// +build integration

package tests

import (
	"fmt"
	"net"
)

// Copied from https://github.com/phayes/freeport.
// BSD Licensed: https://github.com/phayes/freeport/blob/master/LICENSE.md
func getPort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	return fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port), nil
}
