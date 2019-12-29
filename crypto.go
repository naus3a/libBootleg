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
	listener   net.Listener
	BufSize    int
	bListening bool
	bNetInfo   bool
	bProtocol  bool
	bListener  bool
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

func (_l Listener) HasListener() bool {
	return _l.bListener
}

func (_l Listener) IsReady() bool {
	return _l.HasNetInfo() && _l.HasSecret() && _l.HasListener()
}

func (_l *Listener) SetNetInfo(_ip string, _port int) {
	_l.netInfo = NetInfo{_ip, _port}
	_l.bNetInfo = true
	fmt.Println("Listener net info: ", _l.netInfo.String())
}

func (_l *Listener) SetSecret(_secret []byte) {
	_l.cc = makeConfig(_secret)
	_l.bProtocol = true
	fmt.Println("Listener secret set")
}

func (_l *Listener) StartListening(_data chan []byte) bool {
	if !_l.HasNetInfo() || !_l.HasSecret() {
		fmt.Println("Listener NOT ready: cannot setup")
		return false
	}
	if _l.BufSize < 1 {
		_l.BufSize = 100
	}
	var err error
	_l.listener, err = libdisco.Listen("tcp", _l.netInfo.String(), &_l.cc)
	if err != nil {
		fmt.Println("cannot setup listener: ", err)
	} else {
		fmt.Println("Listener setup and listening on ", _l.netInfo, "...")
		go loopListener(_l, _data)
	}
	return true
}

func (_l *Listener) SetupAndListen(_ip string, _port int, _secret []byte, _data chan []byte) bool {
	_l.SetNetInfo(_ip, _port)
	_l.SetSecret(_secret)
	return _l.StartListening(_data)
}

func (_l *Listener) StopListening() {
	if !_l.IsListening() {
		return
	}
	//TODO
}

func loopListener(_l *Listener, _data chan []byte) {
	for {
		var err error
		server, err := _l.listener.Accept()
		if err != nil {
			fmt.Println("server cannot accept: ", err)
			server.Close()
			continue
		}
		fmt.Println("server accepted connection from ", server.RemoteAddr())
		go readSocket(server, _data, _l.BufSize)
	}
}

func readSocket(_srv net.Conn, _data chan []byte, _bufSz int) {
	buf := make([]byte, _bufSz)
	for {
		_, err := _srv.Read(buf)
		if err != nil {
			fmt.Println("server cannot read on socket", err)
			break
		}
		_data <- buf
	}
	fmt.Println("shutting down connection")
	_srv.Close()
}

//---Listener
