package libBootleg

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
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

func password2Key(_password string) []byte {
	_key := []byte(_password)
	if len(_key) < 16 {
		var diff int
		diff = 16 - len(_key)
		for i := 0; i < diff; i++ {
			_key = append(_key, '0')
		}
	}
	return _key
}

func EncryptSecret(_secret []byte, _password string) []byte {
	return libdisco.Encrypt(password2Key(_password), _secret)
}

func DecryptSecret(_secret []byte, _password string) ([]byte, error) {
	return libdisco.Decrypt(password2Key(_password), _secret)
}

func IsEncryptedSecret(_secret []byte) bool {
	if len(_secret) > 0 {
		return _secret[0] == '1'
	}
	return false
}

func MarkSecretEncrypted(_secret *[]byte) {
	*_secret = Insert(*_secret, 0, '1')
}

func MarkSecretPlainText(_secret *[]byte) {
	*_secret = Insert(*_secret, 0, '0')
}

func SaveSecret(_secret []byte, _path string) error {
	var err error
	err = ResetFile(_path)
	if err != nil {
		return err
	}
	MarkSecretPlainText(&_secret)
	err = ioutil.WriteFile(_path, _secret, 0644)
	return err
}

func SaveSecretEncrypted(_secret []byte, _path string, _pass string) error {
	var err error
	err = ResetFile(_path)
	if err != nil {
		return err
	}
	_e := EncryptSecret(_secret, _pass)
	MarkSecretEncrypted(&_e)
	err = ioutil.WriteFile(_path, _e, 0644)
	return err
}

func LoadSecret(_path string, _secret *[]byte) (err error) {
	if DoesFileExist(_path) {
		*_secret, err = ioutil.ReadFile(_path)
		if err != nil {
			return err
		} else {
			if len(*_secret) < 1 {
				return errors.New("no secret or corrupted secret")
			} else {
				if IsEncryptedSecret(*_secret) {
					return errors.New("encrypted secret: password needed")
				}
			}
			*_secret = (*_secret)[1:]

			if len(*_secret) < 32 {
				return errors.New("no secret or corrupted secret")
			} else {
				return err
			}
		}
	} else {
		return errors.New("path does not exists")
	}
}

func LoadSecretEncrypted(_path string, _secret *[]byte, _pass string) (err error) {
	if DoesFileExist(_path) {
		*_secret, err = ioutil.ReadFile(_path)
		if err != nil {
			return err
		} else {
			if len(*_secret) < 1 {
				return errors.New("no secret or corrupted secret")
			} else {
				if IsEncryptedSecret(*_secret) {
					*_secret = (*_secret)[1:]
					*_secret, err = DecryptSecret(*_secret, _pass)
					if err != nil {
						return err
					}
				} else {
					fmt.Println("Your saved token in unencrypted")
					*_secret = (*_secret)[1:]
				}
				if len(*_secret) < 32 {
					return errors.New("no secret or corrupted secret")
				} else {
					return err
				}
			}
		}
	} else {
		return errors.New("path does not exists")
	}
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
