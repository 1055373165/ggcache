package validate

import (
	"net"
	"strconv"
	"strings"
)

func ValidPeerAddr(addr string) bool {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}
	ip := parts[0]
	port := parts[1]
	if (net.ParseIP(ip) == nil && ip != "localhost") || !isValidPort(port) {
		return false
	}
	return true
}

func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p > 0 && p < 65536
}
