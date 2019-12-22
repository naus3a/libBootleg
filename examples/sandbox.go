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

	var tank libBootleg.TokenTank
	tank.AddReadable("sample", s)

	success, n := tank.CheckReadableToken(s)
	if success {
		fmt.Println("welcome " + n)
	} else {
		fmt.Println("I don't know you")
	}
}
