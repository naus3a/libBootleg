package main

import (
	"fmt"
	"github.com/naus3a/libBootleg"
)

func main() {
	var bMsg []byte
	var sMsg string

	var token string = "B56zvdbX_dY6FJEP-s7ipwtG4DtnRlOhCxReSnbpnkA="

	s, _ := libBootleg.DecodeReadableSecret(token)

	ni := libBootleg.NetInfo{
		libBootleg.GetOutboundIp(),
		6666,
	}

	cMsg := make(chan []byte)

	var l libBootleg.Listener
	l.SetupAndListen(ni.Ip, ni.Port, s, cMsg)

	bMsg = <-cMsg
	sMsg = string(bMsg[:len(bMsg)])
	fmt.Println("received: ", sMsg)
}
