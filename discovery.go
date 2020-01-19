package libBootleg

import (
	"fmt"
	"net"
	"time"
)

func MakeDefaultMulticastNetInfo() NetInfo {
	var m NetInfo
	m.Ip = "239.6.6.6"
	m.Port = 6666
	return m
}

func MakeDiscoverPacket() []byte {
	var p []byte
	p = append(p, '0')
	return p
}

func Discover(_ni *NetInfo) error {
	a, err := _ni.UDPAddr()
	if err != nil {
		fmt.Println("Malformed discovery address: ", err)
		return err
	}
	c, err := net.DialUDP("udp", nil, a)
	if err != nil {
		fmt.Println("Cannot send discovery packet: ", err)
		return err
	}
	defer c.Close()
	for i := 0; i < 5; i++ {
		c.Write(MakeDiscoverPacket())
		time.Sleep(1 * time.Second)
	}
	return nil
}

func DiscoverDefaultNetInfo() error {
	m := MakeDefaultMulticastNetInfo()
	return Discover(&m)
}

func ReceiveProbes(_ni *NetInfo) error {
	a, err := _ni.UDPAddr()
	if err != nil {
		fmt.Println("Malformed discovery address: ", err)
		return err
	}
	l, err := net.ListenMulticastUDP("udp", nil, a)
	if err != nil {
		fmt.Println("Cannot start multicast listener: ", err)
		return err
	}
	l.SetReadBuffer(1)
	defer l.Close()
	for {
		b := make([]byte, 1)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			fmt.Println(n, " ", src)
		}
	}
	return nil
}

func ReceiveProbesDefault() error {
	m := MakeDefaultMulticastNetInfo()
	return ReceiveProbes(&m)
}

func ReceiveReply() error {
	return nil
}
