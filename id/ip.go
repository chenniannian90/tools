package id

import (
	"net"
	"os"
)

func ResolveExposedIP() net.IP {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = os.Getenv("HOSTNAME")
	}

	addrList, _ := net.LookupIP(hostname)

	for i := range addrList {
		addr := addrList[len(addrList)-1-i]

		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4
		}
	}

	return nil
}
