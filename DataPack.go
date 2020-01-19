package libBootleg

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

//data---
type DataPack struct {
	Header DataHeader
	Data   []byte
}

func (dp *DataPack) SetProbe() {
	dp.Header.setProbe()
}

func (dp *DataPack) SetText(_txt string) {
	dp.Header.setTextWithSize(uint32(len(_txt)))
	dp.Data = []byte(_txt)
}

func (dp *DataPack) SetFile(_fName string, _d []byte) {
	dp.Header.setFileWithSize(_fName, uint32(len(_d)))
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
	dp.Header.setFileWithSize(filepath.Base(_pth), uint32(len(dp.Data)))
	return
}

func (dp *DataPack) SetFromRaw(_d *[]byte) {
	if *_d == nil || len(*_d) < 1 {
		fmt.Println("Data format: null")
		dp.Header.dataType = DATA_NONE
		return
	}

	switch (*_d)[0] {
	case byte(DATA_PROBE):
		fmt.Println("Data format: probe")
		dp.Header.dataType = DATA_PROBE
		return
	case byte(DATA_TEXT):
		fmt.Println("Data format: text")
		dp.Header.dataType = DATA_TEXT
		dp.Data = (*_d)[1:]
		return
	case byte(DATA_FILE):
		fmt.Println("Data format: file")
		dp.Header.SetFromData(_d)
		if dp.Header.dataType == DATA_NONE {
			return
		}
		var iFrom int
		iFrom = 2 + int(dp.Header.szFileName) + 4
		dp.Data = (*_d)[iFrom:]
	default:
		fmt.Println("Data format: unknown (", string((*_d)[0]), ")")
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
