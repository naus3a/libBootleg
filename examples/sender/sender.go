package main

import (
	"github.com/naus3a/libBootleg"
)

func main() {
	var token string
	token = "B56zvdbX_dY6FJEP-s7ipwtG4DtnRlOhCxReSnbpnkA="

	s, _ := libBootleg.DecodeReadableSecret(token)

	ni := libBootleg.NetInfo{
		libBootleg.GetOutboundIp(),
		6666,
	}

	libBootleg.Send(&ni, s, "cippa")
}
