package libBootleg

import (
	"errors"
	"fmt"
	"github.com/mimoo/disco/libdisco"
	"io"
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"
	"strings"
)

//header---
type DataType byte

const (
	DATA_TEXT  DataType = 't'
	DATA_FILE  DataType = 'f'
	DATA_PROBE DataType = 'p'
	DATA_NONE  DataType = 'n'
)

type DataHeader struct {
	dataType   DataType
	szFileName byte
	fileName   string
}

func (dh *DataHeader) setText() {
	dh.dataType = DATA_TEXT
}

func (dh *DataHeader) setProbe() {
	dh.dataType = DATA_PROBE
}

func (dh *DataHeader) setFile(_fName string) {
	dh.dataType = DATA_FILE
	dh.szFileName = byte(len(_fName))
	dh.fileName = _fName
}

func (dh *DataHeader) SetFromData(_d []byte) {
	switch _d[0] {
	case byte(DATA_PROBE):
		dh.dataType = DATA_PROBE
	case byte(DATA_TEXT):
		dh.dataType = DATA_TEXT
	case byte(DATA_FILE):
		var err error
		dh.dataType = DATA_FILE
		dh.szFileName, err = dh.getFileNameSz(_d)
		if err != nil {
			dh.dataType = DATA_NONE
			return
		}
		dh.fileName, err = dh.getFileName(_d)
		if err != nil {
			dh.dataType = DATA_NONE
			return
		}
	default:
		dh.dataType = DATA_NONE
	}
}

func (dh *DataHeader) getFileNameSz(_d []byte) (_sz byte, err error) {
	if len(_d) < 2 {
		err = errors.New("malformed data")
		return
	}
	_sz = _d[1]
	err = nil
	return
}

func (dh *DataHeader) getFileName(_d []byte) (_fn string, err error) {
	if len(_d) < (2 + int(dh.szFileName)) {
		err = errors.New("malformed data")
		return
	}
	_fn = string(_d[2:(int(dh.szFileName) + 2)])
	err = nil
	return
}

func (dh *DataHeader) GetType() DataType {
	return dh.dataType
}

func (dh *DataHeader) GetFileName() string {
	return dh.fileName
}

func (dh *DataHeader) GetRaw() []byte {
	switch dh.dataType {
	case DATA_TEXT:
		h := []byte{byte(DATA_TEXT)}
		return h
	case DATA_FILE:
		h := []byte{byte(DATA_FILE), dh.szFileName}
		h = append(h, []byte(dh.fileName)...)
		return h
	case DATA_PROBE:
		h := []byte{byte(DATA_PROBE)}
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

func (dp *DataPack) SetProbe() {
	dp.Header.setProbe()
}

func (dp *DataPack) SetText(_txt string) {
	dp.Header.setText()
	dp.Data = []byte(_txt)
}

func (dp *DataPack) SetFile(_fName string, _d []byte) {
	dp.Header.setFile(_fName)
	dp.Data = _d
}

func (dp *DataPack) LoadFile(_pth string) (err error) {
	if !DoesFileExist(_pth) {
		err = errors.New("file does not exist")
		return
	}
	dp.Data, err = ioutil.ReadFile(_pth)
	if err != nil {
		return
	}
	dp.Header.setFile(filepath.Base(_pth))
	return
}

func (dp *DataPack) SetFromRaw(_d []byte) {
	if _d == nil || len(_d) < 1 {
		fmt.Println("Data format: null")
		dp.Header.dataType = DATA_NONE
		return
	}
	switch _d[0] {
	case byte(DATA_PROBE):
		fmt.Println("Data format: probe")
		dp.Header.dataType = DATA_PROBE
		return
	case byte(DATA_TEXT):
		fmt.Println("Data format: text")
		dp.Header.dataType = DATA_TEXT
		dp.Data = _d[1:]
		return
	case byte(DATA_FILE):
		fmt.Println("Data format: file")
		dp.Header.SetFromData(_d)
		if dp.Header.dataType == DATA_NONE {
			return
		}
		dp.Data = _d[int((dp.Header.szFileName)+2):]
	default:
		fmt.Println("Data format: unknown (", string(_d[0]), ")")
		dp.Header.dataType = DATA_NONE
		return
	}
}

func (dp *DataPack) SaveFile() error {
	var err error = nil
	if dp.Header.dataType != DATA_FILE {
		err = errors.New("wrong data type")
		return err
	}
	err = ioutil.WriteFile(dp.Header.fileName, dp.Data, 0644)
	return err
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

func SendDataPack(_ni *NetInfo, _secret []byte, _dp *DataPack, _bVerbose bool) error {
	cc := makeConfig(_secret)
	client, err := libdisco.Dial("tcp", _ni.String(), &cc)
	if err != nil {
		if _bVerbose {
			fmt.Println("Cannot connect to server: ", err)
		}
		return err
	}
	defer client.Close()
	_, err = client.Write(_dp.GetRaw())
	if err != nil {
		if _bVerbose {
			fmt.Println("Cannot write on socket: ", err)
		}
	}
	return err
}

func SendText(_ni *NetInfo, _secret []byte, _msg string) error {
	var dp DataPack
	dp.SetText(_msg)
	return SendDataPack(_ni, _secret, &dp, true)
}

func SendFile(_ni *NetInfo, _secret []byte, _fName string, _d []byte) error {
	var dp DataPack
	dp.SetFile(_fName, _d)
	return SendDataPack(_ni, _secret, &dp, true)
}

func SendFilePath(_ni *NetInfo, _secret []byte, _pth string) error {
	var dp DataPack
	err := dp.LoadFile(_pth)
	if err != nil {
		fmt.Println("Cannot send file: ", err)
		return err
	}
	return SendDataPack(_ni, _secret, &dp, true)
}

func SendProbe(_ni *NetInfo, _secret []byte) error {
	var dp DataPack
	dp.SetProbe()
	err := SendDataPack(_ni, _secret, &dp, false)
	return err
}

func DiscoverReceivers(_ni *NetInfo, _secret []byte) []string {
	//naif scanner; gonna polish it later
	var ips []string
	if SendProbe(_ni, _secret) == nil {
		ips = append(ips, _ni.Ip)
	} else {
		splitIp := strings.Split(_ni.Ip, ".")
		var sLoc string
		var s3 string
		var iLoc int
		for i := 0; i < 3; i++ {
			s3 += splitIp[i]
			s3 += "."
		}
		sLoc = splitIp[len(splitIp)-1]
		iLoc, _ = strconv.Atoi(sLoc)
		for i := 1; i < 254; i++ {
			if i != iLoc {
				var sCur string
				sCur = s3 + strconv.Itoa(i)
				var cni NetInfo
				cni.Ip = sCur
				cni.Port = _ni.Port
				if SendProbe(&cni, _secret) == nil {
					ips = append(ips, cni.Ip)
					i = 300
				}
			}
		}
	}

	return ips
}

//Listener---
type Listener struct {
	netInfo    NetInfo
	cc         libdisco.Config
	listener   net.Listener
	server     net.Conn
	BufSize    int
	bListening bool
	bNetInfo   bool
	bProtocol  bool
	bListener  bool
}

func (_l *Listener) resetFlags() {
	_l.resetListeningFlags()
	_l.bNetInfo = false
	_l.bProtocol = false
}

func (_l *Listener) resetListeningFlags() {
	_l.bListener = false
	_l.bListening = false
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

func (_l *Listener) StartListening(_data chan DataPack) bool {
	_l.resetListeningFlags()
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
		_l.bListener = true
		_l.bListening = true
		fmt.Println("Listener setup and listening on ", _l.netInfo, "...")
		go loopListener(_l, _data)
	}
	return true
}

func (_l *Listener) SetupAndListen(_ip string, _port int, _secret []byte, _data chan DataPack) bool {
	_l.resetFlags()
	_l.SetNetInfo(_ip, _port)
	_l.SetSecret(_secret)
	return _l.StartListening(_data)
}

func (_l *Listener) StopListening() {
	if !_l.IsListening() {
		return
	}
	_l.server.Close()
	_l.resetListeningFlags()
}

func loopListener(_l *Listener, _data chan DataPack) {
	for {
		var err error
		_l.server, err = _l.listener.Accept()
		if err != nil {
			fmt.Println("Listener cannot accept: ", err)
			_l.StopListening()
			continue
		}
		fmt.Println("Listener accepted connection from ", _l.server.RemoteAddr())
		go readSocket(_l, _data, _l.BufSize)
	}
}

func readSocket(_l *Listener, _data chan DataPack, _bufSz int) {
	buf := make([]byte, _bufSz)
	for {
		_, err := _l.server.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Listener cannot read on socket", err)
			}
			break
		}
		var dp DataPack
		dp.SetFromRaw(buf)
		_data <- dp
	}
	fmt.Println("Transfer completed")
	_l.StopListening()
}

//---Listener
