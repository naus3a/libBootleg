package libBootleg

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/mimoo/disco/libdisco"
	"io/ioutil"
)

const szSecret int = 32
const szHash int = 64

//make a random 32bytes secret
func MakeSecret() ([]byte, error) {
	s := make([]byte, szSecret)
	_, err := rand.Read(s)
	return s, err
}

func MakeSecretReadable(_secret []byte) string {
	return base64.URLEncoding.EncodeToString(_secret)
}

func DecodeReadableSecret(_readable string) ([]byte, error) {
	s, err := base64.URLEncoding.DecodeString(_readable)
	if len(s) != szSecret {
		return s, errors.New("corrupted secret")
	} else {
		return s, err
	}

}

func SaveSecret(_secret []byte, _path string) error {
	var err error
	err = ResetFile(_path)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_path, _secret, 0644)
	return err
}

func MakeHash(_secret []byte) []byte {
	return libdisco.Hash(_secret, szHash)
}

//TokenHash---
type TokenHash struct {
	name string
	hash []byte
}

func (_th *TokenHash) SetFromSecret(_name string, _s []byte) {
	_th.name = _name
	_th.hash = MakeHash(_s)
}

func (_th *TokenHash) SetFromReadableSecret(_name string, _readable string) {
	_s, _ := DecodeReadableSecret(_readable)
	_th.SetFromSecret(_name, _s)
}

//---TokenHash

//HashTank---
type TokenTank struct {
	tokens []TokenHash
}

func (_tt *TokenTank) AddTokenHash(_th *TokenHash) {
	_tt.tokens = append(_tt.tokens, *_th)
}

func (_tt *TokenTank) AddSecret(_name string, _s []byte) {
	var _th TokenHash
	_th.SetFromSecret(_name, _s)
	_tt.AddTokenHash(&_th)
}

func (_tt *TokenTank) AddReadable(_name string, _s string) {
	var _th TokenHash
	_th.SetFromReadableSecret(_name, _s)
	_tt.AddTokenHash(&_th)
}

func (_tt *TokenTank) CheckToken(_s []byte) (bool, string) {
	_hash := MakeHash(_s)
	for i := 0; i < len(_tt.tokens); i++ {
		if bytes.Compare(_hash, _tt.tokens[i].hash) == 0 {
			return true, _tt.tokens[i].name
		}
	}
	return false, ""
}

func (_tt *TokenTank) CheckReadableToken(_readable string) (bool, string) {
	_s, _ := DecodeReadableSecret(_readable)
	return _tt.CheckToken(_s)
}

//---HashTank
