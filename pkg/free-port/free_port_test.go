package freeport

import (
	"net"
	"strconv"
	"testing"
)

func TestGetFreePort(t *testing.T) {
	port, err := GetFreePort()
	if err != nil {
		t.Errorf("GetFreePort() error = %v", err)
	}
	if port == 0 {
		t.Errorf("GetFreePort() = %v, want a free port", port)
	}
	l, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		t.Errorf("net.Listen() error = %v", err)
	}
	defer l.Close()
}
