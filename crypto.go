package libBootleg

import (
	"errors"
	"fmt"
	"github.com/mimoo/disco/libdisco"
	"io"
	"net"
	"strconv"
	"strings"
)

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
	//infinite loop to accept multiple clients
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
	var transfer []byte
	var nPkts int
	var totPkts int
	var bIdx int
	var dt DataType
	bIdx = 0
	nPkts = 0
	dt = DATA_NONE
	buf := make([]byte, _bufSz)
	//infinite loop listening to data coming from 1 client
	for {
		_, err := _l.server.Read(buf)

		if nPkts == 0 {
			err = parse1stPacket(&buf, &transfer, &dt, &bIdx, &totPkts)
		} else {
			appendData(&buf, &transfer, &bIdx)
		}

		nPkts++
		if nPkts >= totPkts {
			break
		}
		if err != nil {
			if err != io.EOF {
				fmt.Println("Listener cannot read on socket", err)
			}
			break
		}
	}

	var dp DataPack
	dp.SetFromRaw(&transfer)
	_data <- dp

	fmt.Println("Transfer completed")
	_l.StopListening()
}

func parse1stPacket(_buf *[]byte, _transfer *[]byte, _dt *DataType, _bIdx *int, _totPkts *int) (err error) {
	var szTransfer int
	szTransfer = 1
	*_dt = Byte2DataType((*_buf)[0])
	switch *_dt {
	case DATA_TEXT:
		szTransfer += 4
		var szData uint32
		szData, err = Bytes2Uint32((*_buf)[1:5])
		szTransfer += int(szData)
	case DATA_FILE:
		szName := int((*_buf)[1])
		var szData uint32
		szData, err = Bytes2Uint32((*_buf)[2+szName : 2+szName+4])
		szTransfer += 1
		szTransfer += szName
		szTransfer += 4
		szTransfer += int(szData)
	case DATA_NONE:
		err = errors.New("malformed data")
	}

	*_totPkts = calcNumPkts(len(*_buf), szTransfer)

	if szTransfer >= len(*_buf) {
		*_transfer = make([]byte, szTransfer)
		appendData(_buf, _transfer, _bIdx)
	} else {
		*_transfer = (*_buf)[0:len(*_buf)]
		*_bIdx = len(*_transfer)
	}
	return
}

func calcNumPkts(_szBuf int, _szTransfer int) int {
	var n int
	var m int
	n = _szTransfer / _szBuf
	m = _szTransfer % _szBuf
	if m > 0 {
		n++
	}
	return n
}

func appendData(_buf *[]byte, _transfer *[]byte, _bIdx *int) {
	for i := 0; i < len(*_buf); i++ {
		if *_bIdx < len(*_transfer) {
			(*_transfer)[*_bIdx] = (*_buf)[i]
			*_bIdx++
		}
	}
}

//---Listener:w
