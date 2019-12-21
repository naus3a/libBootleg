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
}
