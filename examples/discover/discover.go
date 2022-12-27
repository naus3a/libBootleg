package main

import (
	"fmt"
	"time"

	"github.com/schollz/peerdiscovery"
)

var discoveries []peerdiscovery.Discovered

func onDiscovered(d peerdiscovery.Discovered) {
	fmt.Println("I got one:")
	fmt.Println(d.Address)
}

func main() {
	s := peerdiscovery.Settings{
		Limit:     1,
		TimeLimit: time.Second * 60,
		Notify:    onDiscovered,
		//AllowSelf: true,
	}
	fmt.Println("Starting discovery")
	discoveries, _ = peerdiscovery.Discover(s)

	/*for _, d := range discoveries {
		fmt.Printf("discovered '%s'\n", d.Address)
	}*/
}
