package main

import (
	"fmt"
	"time"

	"github.com/naus3a/libBootleg"
)

/*

func onDiscovered(d peerdiscovery.Discovered) {
	fmt.Println("I got one:")
	fmt.Println(d.Address)
}*/

func main() {
	s := ""
	var d libBootleg.Discoverable
	d.Init()
	d.StartPublishing()
	for i := 0; i < 5; i++ {
		if d.IsPublishing() {
			s = "publishing"
		} else {
			s = "not publishing"
		}
		fmt.Println(i, ": ", s)
		time.Sleep(1 * time.Second)
	}
	d.StopPublishing()
	if d.IsPublishing() {
		s = "publishing"
	} else {
		s = "not publishing"
	}
	fmt.Println(s)
	/*fmt.Println("Starting publishing myself")

	s := peerdiscovery.Settings{
		Limit:     -1,
		TimeLimit: time.Second * 60,
		Notify:    onDiscovered,
		//DisableBroadcast: true,
	}

	discoveries, err = peerdiscovery.Discover(s)
	if err != nil {
		fmt.Println(err)
		return
	}*/
	/*for _, d := range discoveries {
		fmt.Printf("discovered '%s'\n", d.Address)
	}*/
}
