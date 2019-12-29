package main

import (
	"github.com/naus3a/libBootleg"
)

func main() {
	var token string
	token = "vj1o6DrtmxYnlwavDdMaFEV87L6dByUNzKFN7TJmnsQ="

	s, _ := libBootleg.DecodeReadableSecret(token)

	ni := libBootleg.NetInfo{
		libBootleg.GetOutboundIp(),
		6666,
	}

	libBootleg.Send(&ni, s, "cippa")
}
