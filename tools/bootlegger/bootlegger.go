package main

import (
	"flag"
	"fmt"
	"github.com/naus3a/libBootleg"
	"os"
)

type ToolMode int

const (
	MODE_SENDER ToolMode = iota
	MODE_RECEIVER
	MODE_SECRET
	MODE_NONE
)

type CliFlags struct {
	port    int
	ip      string
	token   string
	data    string
	curMode ToolMode
}

func (cf *CliFlags) setup() {
	cf.curMode = MODE_NONE

	flag.Usage = func() {
		fmt.Printf("Usage: bootlegger [mode] [params]\n\n")

		fmt.Printf("Modes:\n")
		fmt.Printf("  send [yourtext]\n")
		fmt.Printf("\tsend data to a receiver\n")
		fmt.Printf("  receive\n")
		fmt.Printf("\tlisten for data from a sender\n")
		fmt.Printf("  make-secret\n")
		fmt.Printf("\tforge a new token, print it and save it\n")

		fmt.Printf("\nParams:\n")
		flag.PrintDefaults()
	}
	flag.IntVar(&cf.port, "port", 6666, "port listening")
	flag.StringVar(&cf.ip, "ip", libBootleg.GetOutboundIp(), "IP listening")
	flag.StringVar(&cf.token, "token", "whatever token you saved", "the token to use")
}

func (cf *CliFlags) parseSenderData(_args []string, sId int) bool {
	if len(_args) < (sId + 2) {
		return false
	}
	cf.data = ""
	var nAdded int
	nAdded = 0
	for i := (sId + 1); i < len(_args); i++ {
		if _args[i][0] == '-' {
			i = len(_args) + 2
		} else {
			cf.data = cf.data + " " + _args[i]
			nAdded++
		}
	}
	return (nAdded > 0)
}

func (cf *CliFlags) parse() {
	args := os.Args[1:]
	if len(args) < 1 {
		return
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "send":
			if cf.parseSenderData(args, i) {
				cf.curMode = MODE_SENDER
			}
			i = len(args) + 2
		case "receive":
			cf.curMode = MODE_RECEIVER
			i = len(args) + 2
		case "set-secret":
			cf.curMode = MODE_SECRET
			i = len(args) + 2
		}
	}

	flag.Parse()
}

func main() {
	var cliFlags CliFlags
	cliFlags.setup()
	cliFlags.parse()

	switch cliFlags.curMode {
	case MODE_SECRET:
		runSecret()
	case MODE_SENDER:
		runSender(&cliFlags)
	case MODE_RECEIVER:
		runReceiver(&cliFlags)
	case MODE_NONE:
		flag.Usage()
	}
}

func runSecret()               {}
func runSender(cf *CliFlags)   {}
func runReceiver(cf *CliFlags) {}
