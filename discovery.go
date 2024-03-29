package libBootleg

import (
	"fmt"
	"net"
	"time"

	"github.com/schollz/peerdiscovery"
)

func MakeDefaultMulticastNetInfo() NetInfo {
	var m NetInfo
	m.Ip = "239.6.6.6"
	m.Port = 6666
	return m
}

func MakeDiscoverPacket(_secret *[]byte) []byte {
	var p []byte
	if _secret != nil {
		s, err := MakeTotp(*_secret)
		if err != nil {
			p = make([]byte, 6)
		} else {
			p = []byte(s)
		}
	} else {
		p = make([]byte, 6)
	}
	return p
}

func isGoodDiscoveryPacket(_pkt []byte, _secret *[]byte) bool {
	if len(_pkt) != 6 {
		return false
	}

	cmp, err := MakeTotp(*_secret)
	if err != nil {
		return false
	}
	return string(_pkt) == string(cmp)
}

//Discoverer ---

// Discoverer tries to find a listening bootleg instance
type Discoverer struct {
	discoveries   []peerdiscovery.Discovered
	err           error
	secret        *[]byte
	cStopDiscover chan struct{}
}

// Init initializez the discoverer
func (d *Discoverer) Init(secret *[]byte) {
	d.secret = secret
	d.cStopDiscover = make(chan struct{})
}

// Discover discovers listening bootleg instances
func (d *Discoverer) Discover(timeout int) (discovered []peerdiscovery.Discovered, err error) {
	d.err = nil

	s := peerdiscovery.Settings{
		Limit:            -1,
		TimeLimit:        time.Second * time.Duration(timeout),
		DisableBroadcast: true,
		Notify:           d.onDiscovered,
		StopChan:         d.cStopDiscover,
	}

	_, d.err = peerdiscovery.Discover(s)

	err = d.err
	if len(d.discoveries) > 0 {
		discovered = d.discoveries
	}
	return
}

func (d *Discoverer) onDiscovered(discovered peerdiscovery.Discovered) {
	if isGoodDiscoveryPacket(discovered.Payload, d.secret) {
		d.discoveries = append(d.discoveries, discovered)
		close(d.cStopDiscover)
	}
}

//Discoverable---

// Discoverable makes itself discoverable
type Discoverable struct {
	bPublising    bool
	discoveries   []peerdiscovery.Discovered
	err           error
	cStopDiscover chan struct{}
	secret        *[]byte
}

// Init initializes the discoverable object
func (d *Discoverable) Init(secret *[]byte) {
	d.bPublising = false
	d.secret = secret
	d.cStopDiscover = make(chan struct{})
}

//IsPublishing returns true if the discoverable object is publishing itself
func (d *Discoverable) IsPublishing() bool {
	return d.bPublising
}

// StartPublishing starts to publish the discoverable object
func (d *Discoverable) StartPublishing() {
	if d.bPublising {
		return
	}

	d.bPublising = true
	go d.discover()
}

// StopPublishing stops the discoverable object
func (d *Discoverable) StopPublishing() {
	if !d.bPublising {
		return
	}
	close(d.cStopDiscover)
	d.bPublising = false
}

func (d *Discoverable) discover() {
	s := peerdiscovery.Settings{
		Limit:       -1,
		TimeLimit:   -1,
		PayloadFunc: d.makePayload,
		StopChan:    d.cStopDiscover,
	}

	d.discoveries, d.err = peerdiscovery.Discover(s)
}

func (d *Discoverable) makePayload() []byte {
	return MakeDiscoverPacket(d.secret)
}

//---Discoverable

/*type Discoverer struct {
	bRunning bool
	cStop    chan struct{}
	Secret   *[]byte
}

func MakeDefaultMulticastNetInfo() NetInfo {
	var m NetInfo
	m.Ip = "239.6.6.6"
	m.Port = 6666
	return m
}

func MakeDiscoverPacket(_secret *[]byte) []byte {
	var p []byte
	if _secret != nil {
		s, err := MakeTotp(*_secret)
		if err != nil {
			p = make([]byte, 6)
		} else {
			p = []byte(s)
		}
	} else {
		p = make([]byte, 6)
	}
	return p
}

func (d *Discoverer) Discover(_ni *NetInfo) error {
	a, err := _ni.UDPAddr()
	if err != nil {
		fmt.Println("Malformed discovery address: ", err)
		d.bRunning = false
		return err
	}
	c, err := net.DialUDP("udp", nil, a)
	if err != nil {
		fmt.Println("Cannot send discovery packet: ", err)
		d.bRunning = false
		return err
	}

	d.bRunning = true
	defer c.Close()
	for {
		select {
		case <-d.cStop:
			break
		default:
			c.Write(MakeDiscoverPacket(d.Secret))
			time.Sleep(1 * time.Second)
		}
	}
	d.bRunning = false
	return nil
}

func (d *Discoverer) DiscoverDefaultNetInfo() error {
	m := MakeDefaultMulticastNetInfo()
	return d.Discover(&m)
}

func (d *Discoverer) IsRunning() bool {
	return d.bRunning
}

func (d *Discoverer) Start() {
	if d.IsRunning() {
		return
	}
	d.cStop = make(chan struct{})
	go d.DiscoverDefaultNetInfo()
}

func (d *Discoverer) Stop() {
	if !d.IsRunning() {
		return
	}
	close(d.cStop)
	d.bRunning = false
}*/

//---Discoverer

//DiscoveryListener---

type DiscoveryListener struct {
	bRunning bool
	ips      []string
	CIp      chan string
	Secret   *[]byte
}

func ReceiveProbes(_ni *NetInfo, _secret *[]byte) error {
	a, err := _ni.UDPAddr()
	if err != nil {
		fmt.Println("Malformed discovery address: ", err)
		return err
	}
	l, err := net.ListenMulticastUDP("udp", nil, a)
	if err != nil {
		fmt.Println("Cannot start multicast listener: ", err)
		return err
	}
	l.SetReadBuffer(1)
	defer l.Close()
	for {
		b := make([]byte, 6)
		_, src, err := l.ReadFromUDP(b)
		if err == nil {
			if isGoodDiscoveryPacket(b, _secret) {
				sendDiscoveryReply(src.IP.String(), _secret)
			}
		} else {
			fmt.Println("discovery error ", err)
		}
	}
	return nil
}

func ReceiveProbesDefault(_secret *[]byte) error {
	m := MakeDefaultMulticastNetInfo()
	return ReceiveProbes(&m, _secret)
}

func (dl *DiscoveryListener) ReceiveReply(_ip string) ([]string, error) {
	var ips []string
	var ni NetInfo
	ni.Ip = _ip
	ni.Port = 9999
	a, err := ni.UDPAddr()
	if err != nil {
		fmt.Println("Malformed reply  address: ", err)
		dl.bRunning = false
		return ips, err
	}
	l, err := net.ListenUDP("udp", a)
	if err != nil {
		fmt.Println("cannot start receiver: ", err)
		dl.bRunning = false
		return ips, err
	}
	dl.bRunning = true
	for {
		b := make([]byte, 6)
		_, src, err := l.ReadFromUDP(b)
		if err == nil && !alreadyHasString(&ips, src.IP.String()) {
			sIp := src.IP.String()
			ips = append(ips, sIp)
			dl.CIp <- sIp
		}
	}
	dl.bRunning = false
	dl.ips = make([]string, len(ips))
	copy(dl.ips, ips)
	return ips, nil
}

func (dl *DiscoveryListener) IsRunning() bool {
	return dl.bRunning
}

func (dl *DiscoveryListener) GetFoundIps() []string {
	return dl.ips
}

func (dl *DiscoveryListener) Start(_ip string) {
	if dl.IsRunning() {
		return
	}
	dl.CIp = make(chan string)
	dl.ips = nil
	go dl.ReceiveReply(_ip)
}

func (dl *DiscoveryListener) Stop() {
	if !dl.IsRunning() {
		return
	}
	dl.bRunning = false
	close(dl.CIp)
}

func alreadyHasString(_ss *[]string, _s string) bool {
	for i := 0; i < len(*_ss); i++ {
		if (*_ss)[i] == _s {
			return true
		}
	}
	return false
}

func sendDiscoveryReply(_ip string, _secret *[]byte) error {
	var ni NetInfo
	ni.Ip = _ip
	ni.Port = 9999
	a, err := ni.UDPAddr()
	if err != nil {
		fmt.Println("Malformed reply  address: ", err)
		return err
	}
	c, err := net.DialUDP("udp", nil, a)
	if err != nil {
		fmt.Println("Cannot send reply packet: ", err)
		return err
	}
	defer c.Close()
	for i := 0; i < 5; i++ {
		c.Write(MakeDiscoverPacket(_secret))
		time.Sleep(1 * time.Second)
	}
	return nil
}

//---DiscoveryListener
