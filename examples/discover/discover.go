package main

import (
	"fmt"
	"time"

	"github.com/schollz/peerdiscovery"
)

var discoveries []peerdiscovery.Discovered
var err error

func onDiscovered(d peerdiscovery.Discovered) {
	fmt.Println("I got one:")
	fmt.Println(d.Address)
}

func main() {
	fmt.Println("Starting publishing myself")

	s := peerdiscovery.Settings{
		Limit:            1,
		TimeLimit:        time.Second * 60,
		Notify:           onDiscovered,
		DisableBroadcast: true,
	}

	discoveries, err = peerdiscovery.Discover(s)
	if err != nil {
		fmt.Println(err)
		return
	}
	/*for _, d := range discoveries {
		fmt.Printf("discovered '%s'\n", d.Address)
	}*/
}
