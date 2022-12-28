package main

import (
	"fmt"

	"github.com/naus3a/libBootleg"
)

func main() {
	fmt.Println("Starting discovery...")
	var d libBootleg.Discoverer
	discovered, err := d.Discover(10)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(discovered) > 0 {
		fmt.Println("discovered " + discovered[0].Address)
	} else {
		fmt.Println("Could not find a peer")
	}
}
