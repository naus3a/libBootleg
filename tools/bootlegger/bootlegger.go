package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mdp/qrterminal"
	"github.com/naus3a/libBootleg"
)

type ToolMode int

const (
	MODE_SENDER ToolMode = iota
	MODE_RECEIVER
	MODE_SECRET
	MODE_INFO
	MODE_ENCRYPT
	MODE_DECRYPT
	MODE_NONE
)

type SecretAction int

const (
	SECRET_MAKE = iota
	SECRET_CLEAR
	SECRET_SHOW
	SECRET_NONE
)

type InfoAction int

const (
	INFO_SECRET = iota
	INFO_IP
	INFO_NONE
)

//parsing---

type CliFlags struct {
	bufSz         int
	port          int
	ip            string
	defaultIp     string
	token         string
	pass          string
	data          string
	curMode       ToolMode
	curSecAction  SecretAction
	curInfoAction InfoAction
	dataType      libBootleg.DataType
	bQr           bool
}

//this is kinda ugly; maybe let's get rid of the flag lib
var flagQrPtr *bool

func (cf *CliFlags) setup() {
	cf.curMode = MODE_NONE

	cf.defaultIp = libBootleg.GetOutboundIp()

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
		fmt.Printf("\tshow: print saved token\n")
		fmt.Printf("  encrypt [text]\n")
		fmt.Printf("  decrypt [cipher text]\n")
		fmt.Printf("  info [item to show]\n")
		fmt.Printf("\t	secret: print saved token\n")
		fmt.Printf("\t	ip: print local ip\n")

		fmt.Printf("\nParams:\n")
		flag.PrintDefaults()
	}

	flag.IntVar(&cf.bufSz, "bf", 100, "buffer size in bytes")
	flag.IntVar(&cf.port, "port", 6666, "port listening")
	flag.StringVar(&cf.ip, "ip", cf.defaultIp, "IP listening")
	flag.StringVar(&cf.token, "token", "", "the token to use (use saved token if blank)")
	flag.StringVar(&cf.pass, "pass", "", "the password to make or load your saved token (unencrypted if blank)")

	flagQrPtr = flag.Bool("qr", false, "render output as QR code")
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
		return SECRET_SHOW
	default:
		return SECRET_NONE
	}

}

func (cf *CliFlags) parseInfo(_args []string, sId int) InfoAction {
	if len(_args) < (sId + 2) {
		return INFO_NONE
	}
	cf.bQr = false
	switch _args[sId+1] {
	case "secret":
		return INFO_SECRET
	case "ip":
		return INFO_IP
	default:
		return INFO_NONE
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
		case "info":
			cf.curMode = MODE_INFO
			cf.curInfoAction = cf.parseInfo(args, i)
			i = len(args) + 2
		case "encrypt":
			if len(args) > i {
				cf.curMode = MODE_ENCRYPT
				cf.data = args[i+1]
			}
			i = len(args) + 2
		case "decrypt":
			if len(args) > i {
				cf.curMode = MODE_DECRYPT
				cf.data = args[i+1]
			}
			i = len(args) + 2
		}
	}

	flag.Parse()
	cf.bQr = *flagQrPtr
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

func (cf *CliFlags) validateIp() {
	if cf.ip == cf.defaultIp {
		return
	}
	splitIp := strings.Split(cf.ip, ".")
	numTokens := len(splitIp)
	if numTokens < 4 {
		if numTokens == 1 {
			addr, err := strconv.Atoi(splitIp[0])
			if err != nil {
				fmt.Printf("malformed IP: using default")
				cf.ip = cf.defaultIp
			} else {
				if addr < 0 {
					addr = 0
				} else if addr > 255 {
					addr = 255
				}
				splitDefaultIp := strings.Split(cf.defaultIp, ".")
				cf.ip = splitDefaultIp[0] + "." + splitDefaultIp[1] + "." + splitDefaultIp[2] + "." + strconv.Itoa(addr)
			}
		} else {
			fmt.Printf("malformed IP: using default")
			cf.ip = cf.defaultIp
		}
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
	case MODE_INFO:
		runInfo(&cliFlags)
	case MODE_ENCRYPT:
		runEncrypt(&cliFlags)
	case MODE_DECRYPT:
		runDecrypt(&cliFlags)
	case MODE_NONE:
		flag.Usage()
	}
}

//info handling---

func runInfo(cf *CliFlags) {
	switch cf.curInfoAction {
	case INFO_SECRET:
		showSecret(cf)
	case INFO_IP:
		showIp(cf)
	default:
		flag.Usage()
	}
}

func showIp(cf *CliFlags) {
	fmt.Printf(cf.defaultIp + "\n")
	if cf.bQr {
		printQR(cf.defaultIp)
	}
}

//---info handling

//encrypt/decrypt---

func runEncrypt(cf *CliFlags) {
	var s []byte
	err := getSecret(cf, &s)
	if err != nil {
		fmt.Print("")
	}
	cipherText := libBootleg.EncryptText(s, cf.data)
	fmt.Println(cipherText)
	if cf.bQr {
		printQR(cipherText)
	}
}

func runDecrypt(cf *CliFlags) {
	var s []byte
	err := getSecret(cf, &s)
	if err != nil {
		fmt.Print("")
	}
	var plainText string
	plainText, err = libBootleg.DecryptText(s, cf.data)
	fmt.Println(plainText)
	if cf.bQr {
		printQR(plainText)
	}
}

//---encrypt/decrypt

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

func discoverFirstReceiver(_ip *string, _timeout int, _secret *[]byte) {
	var d libBootleg.Discoverer
	d.Init(_secret)
	discovered, err := d.Discover(_timeout)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(discovered) > 0 {
		*_ip = discovered[0].Address
		fmt.Println("Found receiver @ " + discovered[0].Address)
	} else {
		fmt.Println("No receivers found")
	}
}

func runSender(cf *CliFlags) {
	var s []byte
	err := getSecret(cf, &s)
	if err != nil {
		return
	}

	cf.validateIp()
	if (cf.ip == cf.defaultIp) && (cf.dataType != libBootleg.DATA_PROBE) {
		discoverFirstReceiver(&cf.ip, 5, &s)
	}
	ni := libBootleg.NetInfo{
		cf.ip,
		cf.port,
	}

	switch cf.dataType {
	case libBootleg.DATA_TEXT:
		libBootleg.SendText(&ni, s, cf.data)
	case libBootleg.DATA_FILE:
		libBootleg.SendFilePath(&ni, s, cf.data)
	case libBootleg.DATA_PROBE:
		/*var d libBootleg.Discoverer
		var l libBootleg.DiscoveryListener
		l.Secret = &s
		d.Secret = &s
		l.Start(ni.Ip)
		d.Start()
		fmt.Println("Found receivers:")
		for {
			ip := <-l.CIp
			fmt.Println("\t", ip)
		}*/
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

	var d libBootleg.Discoverable
	d.Init(&s)
	d.StartPublishing()

	var bLoop bool
	bLoop = true
	for bLoop {
		data = <-cData
		switch data.Header.GetType() {
		case libBootleg.DATA_TEXT:
			d.StopPublishing()
			//text works, but it's a bit ugly atm
			var sOutput string
			sOutput = ""
			if len(data.Data) >= 4 {
				var textSz int
				textSz = int(data.Data[3])
				for i := 4; i <= (textSz + 3); i++ {
					sOutput += string(data.Data[i])
				}
			}
			fmt.Println(sOutput)
			bLoop = false
		case libBootleg.DATA_FILE:
			d.StopPublishing()
			err = data.SaveFile()
			if err != nil {
				fmt.Println("Cannot save file: ", err)
			} else {
				fmt.Println("File saved to ", data.Header.GetFileName())
			}
			bLoop = false
		}
	}
}

//---receiver
