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

type SecretAction int

const (
	SECRET_MAKE = iota
	SECRET_CLEAR
	SECRET_SHOW
	SECRET_NONE
)

//parsing---

type CliFlags struct {
	port         int
	ip           string
	token        string
	data         string
	curMode      ToolMode
	curSecAction SecretAction
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
		fmt.Printf("  secret [action]\n")
		fmt.Printf("\tmake: forge, print and save new token\n")
		fmt.Printf("\tclear: delete saved token\n")
		fmt.Printf("\tshow: print saved token\n")

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

func (cf *CliFlags) parseSecret(_args []string, sId int) SecretAction {
	if len(_args) < (sId + 2) {
		return SECRET_NONE
	}
	switch _args[sId+1] {
	case "make":
		return SECRET_MAKE
	case "clear":
		return SECRET_CLEAR
	case "show":
		return SECRET_SHOW
	default:
		return SECRET_NONE
	}

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
		case "secret":
			cf.curMode = MODE_SECRET
			cf.curSecAction = cf.parseSecret(args, i)
			i = len(args) + 2
		}
	}

	flag.Parse()
}

func (cf *CliFlags) isGoodFlagToken() bool {
	if len(cf.token) != 32 {
		fmt.Println("Using saved token")
		return false
	} else {
		return true
	}
}

func getSecret(cf *CliFlags, _secret *[]byte) error {
	var err error
	if cf.isGoodFlagToken() {
		*_secret, err = libBootleg.DecodeReadableSecret(cf.token)
		return err
	} else {
		var pth string
		pth, err = loadSecretPath()
		if err == nil {
			err = libBootleg.LoadSecret(pth, _secret)
			return err
		} else {
			return err
		}
	}
}

//---parsing

func main() {
	var cliFlags CliFlags
	cliFlags.setup()
	cliFlags.parse()

	switch cliFlags.curMode {
	case MODE_SECRET:
		runSecret(cliFlags.curSecAction)
	case MODE_SENDER:
		runSender(&cliFlags)
	case MODE_RECEIVER:
		runReceiver(&cliFlags)
	case MODE_NONE:
		flag.Usage()
	}
}

//secret handling---
func runSecret(_sa SecretAction) {
	switch _sa {
	case SECRET_MAKE:
		makeSecret()
	case SECRET_SHOW:
		showSecret()
	case SECRET_CLEAR:
		clearSecret()
	default:
		flag.Usage()
	}
}

func makeSecret() {
	var err error
	var pthDot string
	s, _ := libBootleg.MakeSecret()
	rs := libBootleg.MakeSecretReadable(s)
	pthDot, err = libBootleg.GetDotDirPath()
	if err != nil {
		fmt.Println("Could not get your home path: ", err)
		return
	}
	err = libBootleg.CheckDir(pthDot)
	if err != nil {
		fmt.Println("Could not check .bootleg dir: ", err)
		return
	}
	err = libBootleg.SaveSecret(s, libBootleg.PathJoin(pthDot, "token"))

	if err != nil {
		fmt.Println("Could not save secret: ", err)
		return
	}
	fmt.Println("New token created and saved:")
	fmt.Println(rs)
}

func showSecret() {
	pth, err := loadSecretPath()
	var s []byte
	err = libBootleg.LoadSecret(pth, &s)
	if err != nil {
		fmt.Println("Cannot find a saved secret: ", err)
	} else {
		fmt.Println(libBootleg.MakeSecretReadable(s))
	}
}

func clearSecret() {
	pth, err := loadSecretPath()
	err = libBootleg.ResetFile(pth)
	if err != nil {
		fmt.Println("Could not clear: ", err)
	} else {
		fmt.Println("Secret cleared")
	}
}

func loadSecretPath() (string, error) {
	pth, err := libBootleg.GetDotDirPath()
	if err != nil {
		fmt.Println("Cannot find a saved secret: ", err)
		return "", err
	}
	pth = libBootleg.PathJoin(pth, "token")
	return pth, err
}

//---secret handling

//sender---
func runSender(cf *CliFlags) {
	ni := libBootleg.NetInfo{
		cf.ip,
		cf.port,
	}
	var s []byte
	err := getSecret(cf, &s)
	if err != nil {
		return
	}
	libBootleg.Send(&ni, s, cf.data)
}

//---sender

//receiver---
func runReceiver(cf *CliFlags) {
	var data []byte
	var sData string
	var s []byte
	err := getSecret(cf, &s)
	if err != nil {
		return
	}
	cData := make(chan []byte)
	var l libBootleg.Listener
	l.SetupAndListen(cf.ip, cf.port, s, cData)
	data = <-cData
	sData = string(data[len(data)])
	fmt.Println(sData)
}

//---receiver
