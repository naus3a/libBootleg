package main

import (
	"fmt"
	"time"

	"github.com/naus3a/libBootleg"
)

func main() {
	s := ""
	secret := []byte("123456")
	var d libBootleg.Discoverable
	d.Init(&secret)
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
}
