package speedtest

import (
	"net"
	"net/http"
	"time"
)

var client = http.Client{}

func BindIP(ip net.IP) {
	localTCPAddr := net.TCPAddr{
		IP: ip,
	}

	d := net.Dialer{
		LocalAddr: &localTCPAddr,
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	tr := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                d.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client = http.Client{Transport: tr}
}
