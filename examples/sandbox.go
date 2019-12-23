package main

import (
	"fmt"
	"github.com/naus3a/libBootleg"
)

func main() {
	var s string
	b, err := libBootleg.MakeSecret()
	if err == nil {
		s = libBootleg.MakeSecretReadable(b)
	} else {
		fmt.Println("cannot make a secret")
	}
	fmt.Println(s)

	ni := libBootleg.NetInfo{
		"127.0.0.1",
		6666,
	}
	fmt.Println(ni.String())

	var l libBootleg.Listener
	l.SetNetInfo(ni.Ip, ni.Port)
	l.SetSecret(b)
	l.StartListening()
}
