package main

import (
	"fmt"
	"time"

	"github.com/schollz/peerdiscovery"
)

func main() {
	fmt.Println("Starting publishing myself")

	s := peerdiscovery.Settings{
		Limit:            -1,
		TimeLimit:        time.Second * 60,
		DisableBroadcast: true,
	}

	discoveries, err := peerdiscovery.Discover(s)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, d := range discoveries {
		fmt.Printf("discovered '%s'\n", d.Address)
	}
}
