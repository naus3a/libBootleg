package main

import (
	"fmt"
	"github.com/naus3a/libBootleg"
)

func main() {
	b, _ := libBootleg.MakeSecret()
	fmt.Println(libBootleg.MakeSecretReadable(b))
}
