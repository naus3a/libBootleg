package libBootleg

import (
	"os"
	"os/user"
	"path/filepath"
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

func CheckDir(_path string) error {
	var err error
	if !DoesFileExist(_path) {
		err = os.MkdirAll(_path, os.ModePerm)
	}
	return err
}

func GetHomePath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.HomeDir, err
}

func GetDotDirPath() (string, error) {
	pth, err := GetHomePath()
	if err != nil {
		return "", err
	}
	pth = PathJoin(pth, ".bootleg")
	return pth, err
}

func PathJoin(_path1 string, _path2 string) string {
	return (_path1 + string(filepath.Separator) + _path2)
}

func Insert(_a []byte, _idx int, _el byte) []byte {
	return append(_a[:_idx], append([]byte{_el}, _a[_idx:]...)...)
}
