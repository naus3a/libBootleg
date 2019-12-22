package libBootleg

import (
	"fmt"
	"github.com/mimoo/disco/libdisco"
	"net"
)

func makeConfig(_secret []byte) libdisco.Config {
	return libdisco.Config{
		HandshakePattern: libdisco.NoiseNNpsk2,
		PreSharedKey:     _secret,
	}
}

func Send(_ni *NetInfo, _secret []byte, _msg string) {
	cc := makeConfig(_secret)
	client, err := libdisco.Dial("tcp", _ni.String(), &cc)
	if err != nil {
		fmt.Println("Cannot connect to server: ", err)
		return
	}
	defer client.Close()
	_, err = client.Write([]byte(_msg))
	if err != nil {
		fmt.Println("Cannot write on socket: ", err)
	}
}

//Listener---
type Listener struct {
	netInfo    NetInfo
	cc         libdisco.Config
	listener   net.Conn
	bListening bool
	bNetInfo   bool
	bProtocol  bool
}

func (_l Listener) IsListening() bool {
	return _l.bListening
}

func (_l Listener) HasNetInfo() bool {
	return _l.bNetInfo
}

func (_l Listener) HasSecret() bool {
	return _l.bProtocol
}

func (_l Listener) IsReady() bool {
	return _l.HasNetInfo() && _l.HasSecret()
}

func (_l *Listener) SetNetInfo(_ip string, _port int) {
	_l.netInfo = NetInfo{_ip, _port}
	_l.bNetInfo = true
}

func (_l *Listener) SetSecret(_secret []byte) {
	_l.cc = makeConfig(_secret)
	_l.bProtocol = true
}

func (_l *Listener) Setup() {
	if !_l.IsReady() {
		fmt.Println("Listener NOT ready: cannot setup")
		return
	}

}

func (_l *Listener) SetupAll(_ip string, _port int, _secret []byte) {
	_l.SetNetInfo(_ip, _port)
	_l.SetSecret(_secret)
	_l.Setup()
}

//---Listener
