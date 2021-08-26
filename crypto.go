package libBootleg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/mimoo/disco/libdisco"
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

//OTP---
func MakeOtp(_secret []byte, _ephemeral []byte) (otp string, err error) {
	otp = ""
	hash := libdisco.Hash(append(_secret, _ephemeral...), 32)

	//get last nibble/half byte, which is always <15
	offset := (hash[31] & 15)

	var chunk uint32
	r := bytes.NewReader(hash[offset : offset+4])
	err = binary.Read(r, binary.BigEndian, &chunk)
	if err != nil {
		return
	}
	//as per RFC 4226 ignore most significant bits
	//and divide by 1million to get a reaminder <7digits
	h12 := (int(chunk) & 0x7fffffff) % 1000000
	otp = strconv.Itoa(int(h12))
	if len(otp) < 6 {
		n0 := 6 - len(otp)
		s0 := ""
		for i := 0; i < n0; i++ {
			s0 += "0"
		}
		otp = s0 + otp
	}
	return
}

func MakeTotp(_secret []byte) (otp string, err error) {
	t := math.Floor(float64(time.Now().Unix() / 30))
	sT := fmt.Sprintf("%f", t)
	return MakeOtp(_secret, []byte(sT))
}

//---OTP

//text encryption---

func EncryptText(_secret []byte, _text string) string {
	dataText := []byte(_text)
	cipher := libdisco.Encrypt(_secret, dataText)
	return MakeSecretReadable(cipher)
}

func DecryptText(_secret []byte, _cipherText string) (text string, err error) {
	var dataCipher []byte
	dataCipher, err = DecodeReadableSecret(_cipherText)
	var dataText []byte
	dataText, err = libdisco.Decrypt(_secret, dataCipher)
	text = string(dataText)
	return
}

//---text encryption
