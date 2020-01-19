package libBootleg

import (
	"fmt"
	"github.com/mimoo/disco/libdisco"
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
