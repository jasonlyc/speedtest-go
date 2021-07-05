package speedtest

import (
	"net"
	"net/http"
	"time"
)

var client = http.Client{}
var localAddr *net.TCPAddr

func BindIP(ip net.IP) {
	localAddr = &net.TCPAddr{
		IP: ip,
	}

	d := net.Dialer{
		LocalAddr: localAddr,
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

func GetSockets(remote string, parallel int) ([]*net.TCPConn, error) {
	remoteAddr, err := net.ResolveTCPAddr("tcp4", remote)
	if err != nil {
		return nil, err
	}
	sockets := []*net.TCPConn{}
	for i := 0; i < parallel; i++ {
		conn, _ := net.DialTCP("tcp", localAddr, remoteAddr)
		if conn != nil {
			sockets = append(sockets, conn)
		}
	}
	return sockets, nil
}

func CloseSockets(sockets []*net.TCPConn) {
	for _, socket := range sockets {
		socket.Close()
	}
}
