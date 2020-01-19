package libBootleg

import (
	"encoding/binary"
	"errors"
)

//packet anatony
//Probe:
//0
//p
//Text:
//0|1..5|6..6+d
//t|d   |D
//File:
//0|1|2..2+n|2+n..2+n+4|2+n+4..2+n+4+d
//f|n|N     |d         |D

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
	szData     uint32
}

func (dh *DataHeader) setText() {
	dh.dataType = DATA_TEXT
	dh.szData = 0
}

func (dh *DataHeader) setTextWithSize(_szText uint32) {
	dh.setText()
	dh.szData = _szText
}

func (dh *DataHeader) setProbe() {
	dh.dataType = DATA_PROBE
}

func (dh *DataHeader) setFile(_fName string) {
	dh.dataType = DATA_FILE
	dh.szFileName = byte(len(_fName))
	dh.fileName = _fName
	dh.szData = 0
}

func (dh *DataHeader) setFileWithSize(_fName string, _szData uint32) {
	dh.setFile(_fName)
	dh.szData = _szData
}

func (dh *DataHeader) SetFromData(_d *[]byte) {

	switch (*_d)[0] {
	case byte(DATA_PROBE):
		dh.dataType = DATA_PROBE
	case byte(DATA_TEXT):
		var err error
		dh.dataType = DATA_TEXT
		dh.szData, err = dh.getDataSz(_d)
		if err != nil {
			dh.dataType = DATA_NONE
			return
		}
	case byte(DATA_FILE):
		var err error
		dh.dataType = DATA_FILE
		dh.szFileName, err = dh.getFileNameSz(_d)
		if err != nil {
			dh.dataType = DATA_NONE
			return
		}
		dh.fileName, err = dh.getFileName(_d)
		dh.szData, err = dh.getDataSz(_d)
		if err != nil {
			dh.dataType = DATA_NONE
			return
		}
	default:
		dh.dataType = DATA_NONE
	}
}

func (dh *DataHeader) getFileNameSz(_d *[]byte) (_sz byte, err error) {
	if len(*_d) < 2 {
		err = errors.New("malformed data")
		return
	}
	_sz = (*_d)[1]
	err = nil
	return
}

func (dh *DataHeader) getFileName(_d *[]byte) (_fn string, err error) {
	if len(*_d) < (2 + int(dh.szFileName)) {
		err = errors.New("malformed data")
		return
	}
	_fn = string((*_d)[2:(int(dh.szFileName) + 2)])
	err = nil
	return
}

func (dh *DataHeader) getDataSz(_d *[]byte) (_sz uint32, err error) {
	dt := Byte2DataType((*_d)[0])
	switch dt {
	case DATA_TEXT:
		if len(*_d) < 6 {
			_sz = 0
			err = errors.New("malformed data")
		} else {
			_sz, err = Bytes2Uint32((*_d)[1:5])
		}
	case DATA_FILE:
		idx := 2 + int(dh.szFileName) + 2 + 1
		if len(*_d) < idx {
			_sz = 0
			err = errors.New("malformed data")
		} else {
			_sz, err = Bytes2Uint32((*_d)[idx : idx+4])
		}
	default:
		_sz = 0
		err = errors.New("malformed data")
	}
	return
}

func (dh *DataHeader) GetType() DataType {
	return dh.dataType
}

func (dh *DataHeader) GetFileName() string {
	return dh.fileName
}

func (dh *DataHeader) GetSize() int {
	switch dh.dataType {
	case DATA_NONE:
		return 0
	case DATA_PROBE:
		return 1
	case DATA_TEXT:
		return 6
	case DATA_FILE:
		return (6 + int(dh.szFileName))
	}
	return 0
}

func (dh *DataHeader) GetRaw() []byte {
	switch dh.dataType {
	case DATA_TEXT:
		h := []byte{byte(DATA_TEXT)}
		h = append(h, Uint322Bytes(dh.szData)...)
		return h
	case DATA_FILE:
		h := []byte{byte(DATA_FILE), dh.szFileName}
		h = append(h, []byte(dh.fileName)...)
		h = append(h, Uint322Bytes(dh.szData)...)
		return h
	case DATA_PROBE:
		h := []byte{byte(DATA_PROBE)}
		return h
	default:
		h := []byte{byte(DATA_NONE)}
		return h
	}
}

func Byte2DataType(_byte byte) DataType {
	switch _byte {
	case byte(DATA_TEXT):
		return DATA_TEXT
	case byte(DATA_FILE):
		return DATA_FILE
	case byte(DATA_PROBE):
		return DATA_PROBE
	default:
		return DATA_NONE
	}
}

func Bytes2Uint32(_d []byte) (val uint32, err error) {
	if len(_d) < 4 {
		val = 0
		err = errors.New("makformed data")
		return
	} else {
		val = binary.BigEndian.Uint32(_d)
		err = nil
		return
	}
}

func Uint322Bytes(_val uint32) []byte {
	_d := make([]byte, 4)
	binary.BigEndian.PutUint32(_d, _val)
	return _d
}

//---header
