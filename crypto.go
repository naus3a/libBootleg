package libBootleg

import (
	"fmt"
	"github.com/mimoo/disco/libdisco"
	"net"
)

//header---
type DataType byte

const (
	DATA_TEXT DataType = iota
	DATA_FILE
	DATA_PROBE
	DATA_NONE
)

type DataHeader struct {
	dataType DataType
}

func (dh *DataHeader) setText() {
	dh.dataType = DATA_TEXT
}

func (dh *DataHeader) setProbe() {
	dh.dataType = DATA_PROBE
}

func (dh *DataHeader) SetFromData(_d []byte) {
	switch _d[0] {
	case byte(DATA_PROBE):
		dh.dataType = DATA_PROBE
	case byte(DATA_TEXT):
		dh.dataType = DATA_TEXT

	default:
		dh.dataType = DATA_NONE
	}
}

func (dh *DataHeader) GetRaw() []byte {
	switch dh.dataType {
	case DATA_TEXT:
		h := []byte{byte(DATA_TEXT)}
		return h
	default:
		h := []byte{byte(DATA_NONE)}
		return h
	}
}

//---header

//data---
type DataPack struct {
	Header DataHeader
	Data   []byte
}

func (dp *DataPack) setProbe() {
	dp.Header.setProbe()
}

func (dp *DataPack) setText(_txt string) {
	dp.Header.setText()
	dp.Data = []byte(_txt)
}

func (dp *DataPack) SetFromRaw(_d []byte) {
	if _d == nil || len(_d) < 1 {
		dp.Header.dataType = DATA_NONE
		return
	}
	switch _d[0] {
	case byte(DATA_PROBE):
		dp.Header.dataType = DATA_PROBE
		return
	case byte(DATA_TEXT):
		dp.Header.dataType = DATA_TEXT
		dp.Data = _d[1:]
		return
	default:
		dp.Header.dataType = DATA_NONE
		return
	}
}

func (dp *DataPack) GetRaw() []byte {
	if dp.Data == nil {
		return dp.Header.GetRaw()
	}
	return append(dp.Header.GetRaw(), dp.Data...)
}

//---data

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
	var dp DataPack
	dp.setText(_msg)
	_, err = client.Write(dp.GetRaw())
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
			fmt.Println("Listener cannot accept: ", err)
			server.Close()
			continue
		}
		fmt.Println("Listener accepted connection from ", server.RemoteAddr())
		go readSocket(server, _data, _l.BufSize)
	}
}

func readSocket(_srv net.Conn, _data chan []byte, _bufSz int) {
	buf := make([]byte, _bufSz)
	for {
		_, err := _srv.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Listener cannot read on socket", err)
			}
			break
		}
		var dp DataPack
		dp.SetFromRaw(buf)
		_data <- dp.Data
	}
	fmt.Println("Transfer completed")
	_srv.Close()
}

//---Listener
