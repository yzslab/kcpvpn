package main

import (
	"fmt"
	"github.com/urfave/cli"
	go_tuntap "github.com/yzslab/go-tuntap"
	"github.com/yzslab/goipam"
	"github.com/yzslab/kcpvpn/libbrctl4go"
	"log"
	"os"
	"strings"
)

type CommonConfig struct {
	IP                  string
	Port                uint16
	KCPMode             string
	SendWindowSize      int
	ReceiveWindowSize   int
	Datashard           int
	Parityshard         int
	EnableRapidFec      bool
	DSCP                int
	AckNodelay          bool
	Nodelay             int
	Interval            int
	Resend              int
	NC                  int
	SocketBufferSize    int
	SMuxBufferSize      int
	KeepaliveInerval    int
	Secret              string
	Crypt               string
	VNIMTU              go_tuntap.VirtualNetworkInterfaceMTU
	FullFrameMTU        go_tuntap.VirtualNetworkInterfaceMTU
	UDPMTU              uint16
	LocalIP             uint32
	IP4Netmask          uint32
	VNIMode             go_tuntap.VirtualNetworkInterfaceMode
	EnableTCPSimulation bool
	BRCtl4Go            libbrctl4go.BRCtl4Go
}

type ServerConfig struct {
	CommonConfig
	VNINamePrefix              string
	ConnectionCounter          uint32
	IP4AM                      goipam.IP4AddressManager
	HookDirectory              string
	ConnectedHookInvokeChannel chan *VPNClient
}

type ClientConfig struct {
	CommonConfig
	PeerIP          uint32
	VNIName         string
	isVNIPersistent bool
	ClientIdLength  uint8
	ClientId        string
	AutoReconnect   bool
	OnConnectedHook string
	ClientIPMode
}

// CommonConfig

func (c *CommonConfig) GetIP() string {
	return c.IP
}

func (c *CommonConfig) SetIP(ip string) {
	c.IP = ip
}

func (c *CommonConfig) GetPort() uint16 {
	return c.Port
}

func (c *CommonConfig) SetPort(port uint16) {
	c.Port = port
}

func (c *CommonConfig) SetKCPMode(mode string) {
	c.KCPMode = mode
}

func (c *CommonConfig) GetKCPMode() string {
	return c.KCPMode
}

func (c *CommonConfig) SetSendWindowSize(size int) {
	c.SendWindowSize = size
}

func (c *CommonConfig) GetSendWindowSize() int {
	return c.SendWindowSize
}

func (c *CommonConfig) SetReceiveWindowSize(size int) {
	c.ReceiveWindowSize = size
}

func (c *CommonConfig) GetReceiveWindowSize() int {
	return c.ReceiveWindowSize
}

func (c *CommonConfig) SetDatashard(value int) {
	c.Datashard = value
}

func (c *CommonConfig) GetDatashard() int {
	return c.Datashard
}

func (c *CommonConfig) SetParityshard(value int) {
	c.Parityshard = value
}

func (c *CommonConfig) GetParityshard() int {
	return c.Parityshard
}

func (c *CommonConfig) SetDSCP(value int) {
	c.DSCP = value
}

func (c *CommonConfig) GetDSCP() int {
	return c.DSCP
}

func (c *CommonConfig) SetAckNodelay(nodelay bool) {
	c.AckNodelay = nodelay
}

func (c *CommonConfig) GetAckNodelay() bool {
	return c.AckNodelay
}

func (c *CommonConfig) SetNodelay(value int) {
	c.Nodelay = value
}

func (c *CommonConfig) GetNodelay() int {
	return c.Nodelay
}

func (c *CommonConfig) SetInterval(interval int) {
	c.Interval = interval
}

func (c *CommonConfig) GetInterval() int {
	return c.Interval
}

func (c *CommonConfig) SetResend(value int) {
	c.Resend = value
}

func (c *CommonConfig) GetResend() int {
	return c.Resend
}

func (c *CommonConfig) SetNoCongestion(value int) {
	c.NC = value
}

func (c *CommonConfig) GetNoCongestion() int {
	return c.NC
}

func (c *CommonConfig) SetSocketBufferSize(size int) {
	c.SocketBufferSize = size
}

func (c *CommonConfig) GetSocketBufferSize() int {
	return c.SocketBufferSize
}

func (c *CommonConfig) SetSMuxBufferSize(size int) {
	c.SMuxBufferSize = size
}

func (c *CommonConfig) GetSMuxBufferSize() int {
	return c.SMuxBufferSize
}

func (c *CommonConfig) SetKeepaliveInterval(interval int) {
	c.KeepaliveInerval = interval
}

func (c *CommonConfig) GetKeepaliveInterval() int {
	return c.KeepaliveInerval
}

func (c *CommonConfig) GetSecret() string {
	return c.Secret
}

func (c *CommonConfig) SetSecret(secret string) {
	c.Secret = secret
}

func (c *CommonConfig) GetCrypt() string {
	return c.Crypt
}

func (c *CommonConfig) SetCrypt(crypt string) {
	c.Crypt = crypt
}

func (c *CommonConfig) GetVNIMTU() go_tuntap.VirtualNetworkInterfaceMTU {
	return c.VNIMTU
}

func (c *CommonConfig) SetVNIMTU(mtu go_tuntap.VirtualNetworkInterfaceMTU) {
	c.VNIMTU = mtu
}

func (c *CommonConfig) GetUDPMTU() uint16 {
	return c.UDPMTU
}

func (c *CommonConfig) SetUDPMTU(mtu uint16) {
	c.UDPMTU = mtu
}

func (c *CommonConfig) GetLocalIP() uint32 {
	return c.LocalIP
}

func (c *CommonConfig) SetLocalIP(ip uint32) {
	c.LocalIP = ip
}

func (c *CommonConfig) GetIP4Netmask() uint32 {
	return c.IP4Netmask
}

func (c *CommonConfig) SetIP4Netmask(netmask uint32) {
	c.IP4Netmask = netmask
}

func (c *CommonConfig) GetVNIMode() go_tuntap.VirtualNetworkInterfaceMode {
	return c.VNIMode
}

func (c *CommonConfig) SetVNIMode(mode go_tuntap.VirtualNetworkInterfaceMode) {
	c.VNIMode = mode
}

func (c *CommonConfig) PrintSummary() {
	log.Printf("datashard: %d", c.GetDatashard())
	log.Printf("parityshard: %d", c.GetParityshard())
	log.Printf("dscp: %d", c.GetDSCP())
	log.Printf("socket buffer size: %d", c.GetSocketBufferSize())
	log.Printf("nodelay: %d, interval: %d, resend: %d, no congestion: %d", c.GetNodelay(), c.GetInterval(), c.GetResend(), c.GetNoCongestion())
	log.Printf("udp mtu: %d", c.GetUDPMTU())
	log.Printf("send windows size: %d, receive windows size: %d", c.GetSendWindowSize(), c.GetReceiveWindowSize())
	log.Printf("ack nodelay: %t", c.GetAckNodelay())

	log.Printf("vni mtu: %d", c.GetVNIMTU())
	if c.FullFrameMTU > 0 {
		log.Printf("full frame mtu: %d", c.FullFrameMTU)
	}
}

// ServerConfig

func (c *ServerConfig) GetVNINamePrefix() string {
	return c.VNINamePrefix
}

func (c *ServerConfig) SetVNINamePrefix(name string) {
	c.VNINamePrefix = name
}

func (c *ServerConfig) IncreaseConnectionCounter() uint32 {
	currentValue := c.ConnectionCounter
	c.ConnectionCounter++
	return currentValue
}

func (c *ServerConfig) GetIP4AM() goipam.IP4AddressManager {
	return c.IP4AM
}

// ClientConfig

func (c *ClientConfig) GetPeerIP() uint32 {
	return c.PeerIP
}

func (c *ClientConfig) SetPeerIP(ip uint32) {
	c.PeerIP = ip
}

func (c *ClientConfig) GetVNIName() string {
	return c.VNIName
}

func (c *ClientConfig) SetVNIName(name string) {
	c.VNIName = name
}

func (c *ClientConfig) IsVNIPersistent() bool {
	return c.isVNIPersistent
}

func (c *ClientConfig) SetVNIPersistent(value bool) {
	c.isVNIPersistent = value
}

func createConfigFromCLI(extraFlags []cli.Flag, config *CommonConfig, cliCallback func(c *cli.Context) error) error {
	app := cli.NewApp()
	app.Name = "KCPVPN"
	app.Usage = "KCP based VPN"
	app.Version = "2019112001"

	commonFlags := []cli.Flag{
		cli.StringFlag{
			Name:     "ip",
			Usage:    "the ip address to listen in server mode, or the ip address to connect in client mode",
			Required: true,
			Value:    "127.0.0.1",
		},
		cli.UintFlag{
			Name:     "port",
			Usage:    "the port to listen in server mode, or the port to connect in client mode",
			Required: true,
		},
		cli.UintFlag{
			Name:     "udp-mtu",
			Usage:    "set maximum transmission unit for udp packets",
			Required: false,
			Value:    1350,
		},
		cli.StringFlag{
			Name:  "kcp-mode",
			Value: "fast",
			Usage: "profiles: fast3, fast2, fast, normal, manual",
		},
		cli.IntFlag{
			Name:  "sndwnd",
			Value: 1024,
			Usage: "set send window size(num of packets)",
		},
		cli.IntFlag{
			Name:  "rcvwnd",
			Value: 1024,
			Usage: "set receive window size(num of packets)",
		},
		cli.IntFlag{
			Name:  "datashard,ds",
			Value: 10,
			Usage: "set reed-solomon erasure coding - datashard",
		},
		cli.IntFlag{
			Name:  "parityshard,ps",
			Value: 3,
			Usage: "set reed-solomon erasure coding - parityshard",
		},
		cli.BoolFlag{
			Name:  "rapid-fec",
			Usage: "enable rapid fec mode",
		},
		cli.IntFlag{
			Name:  "dscp",
			Value: 0,
			Usage: "set DSCP(6bit)",
		},
		cli.BoolFlag{
			Name:   "acknodelay",
			Usage:  "flush ack immediately when a packet is received",
			Hidden: true,
		},
		cli.IntFlag{
			Name:   "nodelay",
			Value:  0,
			Hidden: true,
		},
		cli.IntFlag{
			Name:   "interval",
			Value:  50,
			Hidden: true,
		},
		cli.IntFlag{
			Name:   "resend",
			Value:  0,
			Hidden: true,
		},
		cli.IntFlag{
			Name:   "nc",
			Value:  0,
			Hidden: true,
		},
		cli.IntFlag{
			Name:  "sockbuf",
			Value: 4194304, // socket buffer size in bytes
			Usage: "per-socket buffer in bytes",
		},
		cli.IntFlag{
			Name:  "smuxbuf",
			Value: 4194304,
			Usage: "the overall de-mux buffer in bytes",
		},
		cli.IntFlag{
			Name:  "keepalive",
			Value: 10, // nat keepalive interval in seconds
			Usage: "seconds between heartbeats",
		},
		cli.StringFlag{
			Name:     "secret",
			Usage:    "pre-shared secret between client and server",
			EnvVar:   "KCPVPN_KEY",
			Required: true,
		},
		cli.StringFlag{
			Name:     "crypt",
			Usage:    "aes, aes-128, aes-192, salsa20, blowfish, twofish, cast5, 3des, tea, xtea, xor, sm4, none",
			Required: false,
			Value:    "aes",
		},
		cli.StringFlag{
			Name:     "local-ip",
			Usage:    "local ip address for tun (tun mode only)",
			Value:    "",
			Required: false,
		},
		cli.StringFlag{
			Name:  "netmask",
			Usage: "example: 255.255.255.0, only required in tap mode",
			Value: "",
		},
		cli.IntFlag{
			Name:     "vni-mtu",
			Usage:    "set maximum transmission unit for virtual network interface packets",
			Required: false,
			Value:    1500,
		},
		cli.StringFlag{
			Name:  "bridge",
			Usage: "auto add interface to specific bridge, only available on tap mode",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "tcp",
			Usage: "enable tcp simulation",
		},
	}

	app.Flags = append(commonFlags, extraFlags...)

	app.Action = func(c *cli.Context) error {
		config.SetIP(c.String("ip"))
		config.SetPort(uint16(c.Uint("port")))
		config.SetUDPMTU(uint16(c.Uint("udp-mtu")))
		config.SetSecret(c.String("secret"))
		config.SetCrypt(c.String("crypt"))

		config.SetKCPMode(c.String("kcp-mode"))
		config.SetSendWindowSize(c.Int("sndwnd"))
		config.SetReceiveWindowSize(c.Int("rcvwnd"))
		config.SetDatashard(c.Int("datashard"))
		config.SetParityshard(c.Int("parityshard"))
		config.EnableRapidFec = c.Bool("rapid-fec")
		config.SetDSCP(c.Int("dscp"))
		config.SetAckNodelay(c.Bool("acknodelay"))
		config.SetNodelay(c.Int("nodelay"))
		config.SetInterval(c.Int("interval"))
		config.SetResend(c.Int("resend"))
		config.SetNoCongestion(c.Int("nc"))
		config.SetSocketBufferSize(c.Int("sockbuf"))
		config.SetSMuxBufferSize(c.Int("smuxbuf"))
		config.SetKeepaliveInterval(c.Int("keepalive"))

		config.SetVNIMTU(go_tuntap.VirtualNetworkInterfaceMTU(c.Uint("vni-mtu")))

		switch config.GetKCPMode() {
		case "normal":
			config.SetNodelay(0)
			config.SetInterval(40)
			config.SetResend(2)
			config.SetNoCongestion(1)
		case "fast":
			config.SetNodelay(0)
			config.SetInterval(30)
			config.SetResend(2)
			config.SetNoCongestion(1)
		case "fast2":
			config.SetNodelay(1)
			config.SetInterval(20)
			config.SetResend(2)
			config.SetNoCongestion(1)
		case "fast3":
			config.SetNodelay(1)
			config.SetInterval(10)
			config.SetResend(2)
			config.SetNoCongestion(1)
		}

		localIPString := c.String("local-ip")
		if localIPString != "" {
			localIP, err := ip2long(localIPString)
			if err != nil {
				return cli.NewExitError(err, -1)
			}
			config.SetLocalIP(localIP)
		}

		netmaskString := c.String("netmask")
		if netmaskString != "" {
			netmask, err := ip2long(netmaskString)
			if err != nil {
				return cli.NewExitError(fmt.Sprintf("invalid netmask: %s", err.Error()), -1)
			}
			config.SetIP4Netmask(netmask)
		}

		bridgeName := c.String("bridge")
		if bridgeName != "" {
			brctl4go, err := libbrctl4go.OpenLinuxBridge(bridgeName, true)
			if err != nil {
				return cli.NewExitError(err, -1)
			}
			config.BRCtl4Go = brctl4go
		}

		config.EnableTCPSimulation = c.Bool("tcp")

		err := cliCallback(c)
		if err != nil {
			return cli.NewExitError(err, -1)
		}

		return nil
	}

	return app.Run(os.Args[1:])
}

func createServerConfig(onCreated func(serverConfig *ServerConfig)) error {
	vpnServerConfig := ServerConfig{}
	err := createConfigFromCLI([]cli.Flag{
		cli.StringFlag{
			Name:  "vni-name-prefix",
			Usage: "virtual network interface name",
			Value: "kvs",
		},
		cli.StringFlag{
			Name:     "vni-mode",
			Usage:    "tun or tap",
			Required: true,
		},
		cli.StringFlag{
			Name:  "assignable-ips",
			Usage: "192.168.0.0/24 or 192.168.0.0-192.168.0.255",
			Value: "",
		},
		cli.StringFlag{
			Name:  "hook-dir",
			Value: "",
		},
		cli.IntFlag{
			Name:  "full-frame-mtu",
			Usage: "decide automatically if is 0",
			Value: 0,
		},
	}, &vpnServerConfig.CommonConfig, func(c *cli.Context) error {
		vpnServerConfig.VNINamePrefix = c.String("vni-name-prefix")

		// parse vni mode
		vniMode := c.String("vni-mode")
		if vniMode == "tun" {
			vpnServerConfig.SetVNIMode(go_tuntap.TUN)
		} else if vniMode == "tap" {
			vpnServerConfig.SetVNIMode(go_tuntap.TAP)
		} else {
			return cli.NewExitError("invalid --vni-mode value", -1)
		}

		// parse assignable ips
		assignableIPs := c.String("assignable-ips")
		if assignableIPs == "" {
			vpnServerConfig.IP4AM = NewConstantIP4AM(0)
		} else {
			var err error
			if strings.Index(assignableIPs, "-") != -1 {
				ipRange := strings.Split(assignableIPs, "-")
				vpnServerConfig.IP4AM, err = goipam.NewIP4BitmapFromStringRange(ipRange[0], ipRange[1])
				if err != nil {
					return cli.NewExitError(err, -1)
				}
			} else {
				vpnServerConfig.IP4AM, err = goipam.NewIP4BitmapFromSubnet(assignableIPs)
				if err != nil {
					return cli.NewExitError(err, -1)
				}
			}

			if vpnServerConfig.LocalIP > 0 {
				_ = vpnServerConfig.IP4AM.AssignSpecificIP(vpnServerConfig.LocalIP)
			}
		}

		if (assignableIPs != "" || vpnServerConfig.GetLocalIP() != 0) && vpnServerConfig.GetIP4Netmask() == 0 && vpnServerConfig.GetVNIMode() == go_tuntap.TAP {
			return cli.NewExitError("--netmask is required in tap mode with --assignable-ips or --local-ip", -1)
		}

		vpnServerConfig.HookDirectory = c.String("hook-dir")

		fullFrameMTU := c.Uint("full-frame-mtu")
		if fullFrameMTU == 0 {
			vpnServerConfig.FullFrameMTU = vpnServerConfig.GetVNIMTU() + 22
		} else {
			vpnServerConfig.FullFrameMTU = go_tuntap.VirtualNetworkInterfaceMTU(fullFrameMTU)
		}

		onCreated(&vpnServerConfig)
		return nil
	})

	return err
}

func createClientConfig(onCreated func(clientConfig *ClientConfig)) error {
	clientFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "client-id",
			Usage: "client id for some specific configuration on server, max length 16",
			Value: "",
		},
		cli.StringFlag{
			Name:  "vni-name",
			Usage: "virtual network interface name (client mode only)",
			Value: "kvc0",
		},
		cli.BoolFlag{
			Name:     "persistent-vni",
			Required: false,
		},
		cli.BoolFlag{
			Name:  "auto-reconnect",
			Usage: "auto reconnect to server on connection closed",
		},
		cli.BoolFlag{
			Name:  "no-ip-configuration",
			Usage: "useful if a DHCP server exists, or bridge enabled",
		},
		cli.StringFlag{
			Name:  "on-connected-hook",
			Usage: "path to hook file",
		},
	}

	vpnClientConfig := ClientConfig{}

	err := createConfigFromCLI(clientFlags, &vpnClientConfig.CommonConfig, func(c *cli.Context) error {
		vpnClientConfig.ClientId = c.String("client-id")
		clientIdLength := len(vpnClientConfig.ClientId)
		if clientIdLength > 16 {
			return cli.NewExitError("then length of client id must not > 16", -1)
		}
		vpnClientConfig.ClientIdLength = uint8(clientIdLength)

		vpnClientConfig.VNIName = c.String("vni-name")
		vpnClientConfig.SetVNIPersistent(c.Bool("persistent-vni"))

		if len(vpnClientConfig.VNIName) == 0 {
			return cli.NewExitError("vni-name values is required in client mode", -1)
		}

		if vpnClientConfig.CommonConfig.GetLocalIP() == 0 {
			vpnClientConfig.ClientIPMode = ClientIPModeServerAssign
		} else {
			vpnClientConfig.ClientIPMode = ClientIPModeClientSet
		}

		vpnClientConfig.AutoReconnect = c.Bool("auto-reconnect")

		if c.Bool("no-ip-configuration") {
			vpnClientConfig.ClientIPMode = ClientIPModeOther
		}

		vpnClientConfig.OnConnectedHook = c.String("on-connected-hook")

		onCreated(&vpnClientConfig)
		return nil
	})

	return err
}
