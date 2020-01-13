package ssdp

import (
	"net"
	"time"
)

var (
	sendAddrIPv4 = "239.255.255.250:1900"
	recvAddrIPv4 = "224.0.0.0:1900"
	ssdpAddrIPv4 *net.UDPAddr
)

func init() {
	// FIXME: https://github.com/koron/go-ssdp/issues/9
	var err error
	ssdpAddrIPv4, err = net.ResolveUDPAddr("udp4", sendAddrIPv4)
	if err != nil {
		panic(err)
	}
}

type packetHandler func(net.Addr, []byte) error

func readPackets(conn *net.UDPConn, timeout time.Duration, h packetHandler) error {
	buf := make([]byte, 65535)
	conn.SetReadBuffer(len(buf))
	conn.SetReadDeadline(time.Now().Add(timeout))
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				return nil
			}
			return err
		}
		if err := h(addr, buf[:n]); err != nil {
			return err
		}
	}
}

func sendTo(to *net.UDPAddr, data []byte) (int, error) {
	conn, err := net.DialUDP("udp4", nil, to)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	n, err := conn.Write(data)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// SetMulticastSendAddrIPv4 updates a UDP address to send multicast packets.
func SetMulticastSendAddrIPv4(s string) error {
	// FIXME: https://github.com/koron/go-ssdp/issues/9
	addr, err := net.ResolveUDPAddr("udp4", s)
	if err != nil {
		return err
	}
	ssdpAddrIPv4 = addr
	return nil
}

// SetMulticastRecvAddrIPv4 updates multicast address where to receive packets.
func SetMulticastRecvAddrIPv4(s string) error {
	// FIXME: https://github.com/koron/go-ssdp/issues/9
	_, err := net.ResolveUDPAddr("udp4", s)
	if err != nil {
		return err
	}
	recvAddrIPv4 = s
	return nil
}
