package utils

import "net"

func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) { _ = conn.Close() }(conn)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
