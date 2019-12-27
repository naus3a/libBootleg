package libBootleg

import (
	"os"
)

func DoesFileExist(_path string) bool {
	if _, err := os.Stat(_path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func ResetFile(_path string) error {
	var err error
	var f *os.File

	if DoesFileExist(_path) {
		err = os.Remove(_path)
		if err != nil {
			return err
		}
	}

	f, err = os.Create(_path)
	if err != nil {
		defer f.Close()
	}
	return err
}

func CheckDir(_path string) {
	if !DoesFileExist(_path) {
		os.Mkdir(_path, os.ModeDir)
	}
}
