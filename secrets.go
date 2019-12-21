package libBootleg

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

const szSecret int = 32

//const minReadableChar int = 33
//const maxReadableChar int = 176

//make a random 32bytes secret
func MakeSecret() ([]byte, error) {
	s := make([]byte, szSecret)
	_, err := rand.Read(s)
	return s, err
}

func MakeSecretReadable(secret []byte) string {
	return base64.URLEncoding.EncodeToString(secret)
}

func DecodeReadableSecret(readable string) ([]byte, error) {
	s, err := base64.URLEncoding.DecodeString(readable)
	if len(s) != szSecret {
		return s, errors.New("corrupted secret")
	} else {
		return s, err
	}

}
