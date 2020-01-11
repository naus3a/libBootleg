package main

import (
	"flag"
	"fmt"
	"github.com/mdp/qrterminal"
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
	bufSz        int
	port         int
	ip           string
	token        string
	pass         string
	data         string
	curMode      ToolMode
	curSecAction SecretAction
	dataType     libBootleg.DataType
	bQr          bool
}

func (cf *CliFlags) setup() {
	cf.curMode = MODE_NONE

	flag.Usage = func() {
		fmt.Printf("Usage: bootlegger [optional params] [mode]\n\n")

		fmt.Printf("Modes:\n")
		fmt.Printf("  send [yourtext]\n")
		fmt.Printf("  transfer [path/to/file]")
		fmt.Printf("\tsend data to a receiver\n")
		fmt.Printf("  receive\n")
		fmt.Printf("\tlisten for data from a sender\n")
		fmt.Printf("  discover\n")
		fmt.Printf("\tdiscover listening bootleggers\n")
		fmt.Printf("  secret [action]\n")
		fmt.Printf("\tmake: forge (make random if you don't specify a token), print and save new token\n")
		fmt.Printf("\tclear: delete saved token\n")
		fmt.Printf("\tshow [qr]: print saved token (as a QR code if you specify the qr option)\n")

		fmt.Printf("\nParams:\n")
		flag.PrintDefaults()
	}
	flag.IntVar(&cf.bufSz, "bf", 100, "buffer size in bytes")
	flag.IntVar(&cf.port, "port", 6666, "port listening")
	flag.StringVar(&cf.ip, "ip", libBootleg.GetOutboundIp(), "IP listening")
	flag.StringVar(&cf.token, "token", "", "the token to use (use saved token if blank)")
	flag.StringVar(&cf.pass, "pass", "", "the password to make or load your saved token (unencrypted if blank)")
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
			if nAdded > 0 {
				cf.data = cf.data + " "
			}
			cf.data = cf.data + _args[i]
			nAdded++
		}
	}
	return (nAdded > 0)
}

func (cf *CliFlags) parseTransferData(_args []string, sId int) bool {
	if len(_args) < (sId + 2) {
		return false
	}
	cf.data = _args[sId+1]
	return true
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
		cf.bQr = false
		if len(_args) >= sId+3 {
			if _args[sId+2] == "qr" {
				cf.bQr = true
			}
		}
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
				cf.dataType = libBootleg.DATA_TEXT
			}
			i = len(args) + 2
		case "transfer":
			if cf.parseTransferData(args, i) {
				cf.curMode = MODE_SENDER
				cf.dataType = libBootleg.DATA_FILE
			}
			i = len(args) + 2
		case "discover":
			cf.curMode = MODE_SENDER
			cf.dataType = libBootleg.DATA_PROBE
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
	if len(cf.token) < 32 {
		fmt.Println("Using saved token")
		return false
	} else {
		return true
	}
}

func (cf *CliFlags) hasPassword() bool {
	return (len(cf.pass) > 0)
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
			if cf.hasPassword() {
				err = libBootleg.LoadSecretEncrypted(pth, _secret, cf.pass)
			} else {
				err = libBootleg.LoadSecret(pth, _secret)
			}
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
		runSecret(&cliFlags)
	case MODE_SENDER:
		runSender(&cliFlags)
	case MODE_RECEIVER:
		runReceiver(&cliFlags)
	case MODE_NONE:
		flag.Usage()
	}
}

//secret handling---
func runSecret(cf *CliFlags) {
	switch cf.curSecAction {
	case SECRET_MAKE:
		makeSecret(cf)
	case SECRET_SHOW:
		showSecret(cf)
	case SECRET_CLEAR:
		clearSecret()
	default:
		flag.Usage()
	}
}

func printQR(_s string) {
	config := qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}
	qrterminal.GenerateWithConfig(_s, config)
}

func makeSecret(cf *CliFlags) {
	var err error
	var pthDot string
	var s []byte
	var rs string
	if cf.isGoodFlagToken() {
		rs = cf.token
		s, _ = libBootleg.DecodeReadableSecret(rs)
	} else {
		s, _ = libBootleg.MakeSecret()
		rs = libBootleg.MakeSecretReadable(s)
	}
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
	if cf.hasPassword() {
		err = libBootleg.SaveSecretEncrypted(s, libBootleg.PathJoin(pthDot, "token"), cf.pass)
	} else {
		err = libBootleg.SaveSecret(s, libBootleg.PathJoin(pthDot, "token"))
	}

	if err != nil {
		fmt.Println("Could not save secret: ", err)
		return
	}
	fmt.Println("New token created and saved:")
	fmt.Println(rs)
}

func showSecret(cf *CliFlags) {
	pth, err := loadSecretPath()
	var s []byte
	if cf.hasPassword() {
		err = libBootleg.LoadSecretEncrypted(pth, &s, cf.pass)
	} else {
		err = libBootleg.LoadSecret(pth, &s)
	}
	if err != nil {
		fmt.Println("Cannot find a saved secret: ", err)
	} else {
		rs := libBootleg.MakeSecretReadable(s)
		fmt.Println(rs)
		if cf.bQr {
			printQR(rs)
		}
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
	switch cf.dataType {
	case libBootleg.DATA_TEXT:
		libBootleg.SendText(&ni, s, cf.data)
	case libBootleg.DATA_FILE:
		libBootleg.SendFilePath(&ni, s, cf.data)
	case libBootleg.DATA_PROBE:
		ips := libBootleg.DiscoverReceivers(&ni, s)
		fmt.Println("Valid receivers:")
		for i := 0; i < len(ips); i++ {
			fmt.Println("\t", ips[i])
		}
	default:
		break
	}
}

//---sender

//receiver---
func runReceiver(cf *CliFlags) {
	var data libBootleg.DataPack
	var s []byte
	err := getSecret(cf, &s)
	if err != nil {
		fmt.Println("Cannot start a receiver: ", err)
	}
	cData := make(chan libBootleg.DataPack)
	var l libBootleg.Listener
	l.BufSize = cf.bufSz
	l.SetupAndListen(cf.ip, cf.port, s, cData)
	data = <-cData
	switch data.Header.GetType() {
	case libBootleg.DATA_TEXT:
		fmt.Println(string(data.Data))
	case libBootleg.DATA_FILE:
		err = data.SaveFile()
		if err != nil {
			fmt.Println("Cannot save file: ", err)
		} else {
			fmt.Println("File saved to ", data.Header.GetFileName())
		}
	}
}

//---receiver
