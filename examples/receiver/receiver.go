package main

import (
	"github.com/naus3a/libBootleg"
)

func main() {
	var token string
	token = "B56zvdbX_dY6FJEP-s7ipwtG4DtnRlOhCxReSnbpnkA="

	s, _ := libBootleg.DecodeReadableSecret(token)

	ni := libBootleg.NetInfo{
		"127.0.0.1",
		6666,
	}

	var l libBootleg.Listener

	l.SetupAndListen(ni.Ip, ni.Port, s)

}
