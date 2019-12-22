package libBootleg

import (
	"fmt"
)

type NetInfo struct {
	Ip   string
	Port int
}

func (_ni NetInfo) String() string {
	return fmt.Sprintf("%v:%v", _ni.Ip, _ni.Port)
}
