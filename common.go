package main

import (
	"crypto/sha1"
	"github.com/xtaci/smux"
	go_tuntap "github.com/yzslab/go-tuntap"
	"github.com/yzslab/kcp-go"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"log"
	"net"
	"time"
	"unsafe"
)

type VNIConfig struct {
	Mode         go_tuntap.VirtualNetworkInterfaceMode
	Name         string
	MTU          go_tuntap.VirtualNetworkInterfaceMTU
	Netmask      uint32
	LocalIP      uint32
	PeerIP       uint32
	IsPersistent bool
}

var ServerConfigurationSize = unsafe.Sizeof(go_tuntap.VirtualNetworkInterfaceMode(0)) + (unsafe.Sizeof(uint32(0)) * 2) + (unsafe.Sizeof(uint16(0)) * 2)

var logChannel = make(chan string, 16)

const ControlMessageBufferSize = 64

func createSMux(conn net.Conn, bufferSize int, keepAliveInterval int) (*smux.Session, error) {
	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = bufferSize
	smuxConfig.KeepAliveInterval = time.Duration(keepAliveInterval) * time.Second

	mux, err := smux.Server(conn, smuxConfig)
	return mux, err
}

func createBlock(config *CommonConfig) (block kcp.BlockCrypt) {
	pass := pbkdf2.Key([]byte(config.GetSecret()), []byte("KCPVPN"), 4096, 32, sha1.New)
	var err error

	switch config.GetCrypt() {
	case "sm4":
		block, err = kcp.NewSM4BlockCrypt(pass[:16])
	case "tea":
		block, err = kcp.NewTEABlockCrypt(pass[:16])
	case "xor":
		block, err = kcp.NewSimpleXORBlockCrypt(pass)
	case "none":
		block, err = kcp.NewNoneBlockCrypt(pass)
	case "aes-128":
		block, err = kcp.NewAESBlockCrypt(pass[:16])
	case "aes-192":
		block, err = kcp.NewAESBlockCrypt(pass[:24])
	case "blowfish":
		block, err = kcp.NewBlowfishBlockCrypt(pass)
	case "twofish":
		block, err = kcp.NewTwofishBlockCrypt(pass)
	case "cast5":
		block, err = kcp.NewCast5BlockCrypt(pass[:16])
	case "3des":
		block, err = kcp.NewTripleDESBlockCrypt(pass[:24])
	case "xtea":
		block, err = kcp.NewXTEABlockCrypt(pass[:16])
	case "salsa20":
		block, err = kcp.NewSalsa20BlockCrypt(pass)
	case "aes":
		block, err = kcp.NewAESBlockCrypt(pass)
	default:
		log.Fatal("Unsupported crypt")
	}

	if err != nil {
		log.Fatal(err)
	}

	return block
}

func CreateVNI(config *VNIConfig, logger func(string)) (go_tuntap.VirtualNetworkInterface, error) {
	var vni go_tuntap.VirtualNetworkInterface
	var err error

	logger("creating vni")
	vni, err = go_tuntap.NewLinuxVirtualNetworkInterface(config.Mode, config.Name, config.IsPersistent)
	if err != nil {
		goto ReturnError
	}
	logger("vni created")
	logger("vni setting mtu")
	err = vni.SetMTU(config.MTU)
	if err != nil {
		goto CloseVNI
	}
	logger("vni mtu set")
	if config.LocalIP > 0 {
		logger("setting local ip")
		netmask := config.Netmask
		// do not set netmask in tun mode
		if config.Mode == go_tuntap.TUN {
			netmask = 0
		}
		err = vni.SetBinaryAddress(go_tuntap.Htonl(config.LocalIP), go_tuntap.Htonl(netmask))
		if err != nil {
			goto CloseVNI
		}
		logger("local ip set")
	}
	if config.Mode == go_tuntap.TUN {
		if tunVNI, ok := vni.(go_tuntap.VirtualNetworkTUN); ok {
			if config.PeerIP > 0 {
				logger("setting peer ip")
				err = tunVNI.SetBinaryDestinationAddress(go_tuntap.Htonl(config.PeerIP))
				if err != nil {
					goto CloseVNI
				}
				logger("peer ip set")
			}
		}
	}

	return vni, nil
CloseVNI:
	_ = vni.Close
ReturnError:
	return nil, err
}

func readThenWrite(reader io.Reader, writer io.Writer, bufferSize uint) error {
	var err error
	var nRead int
	buffer := make([]byte, bufferSize)
	for {
		nRead, err = reader.Read(buffer)
		// log.Printf("%d byte(s) read", nRead)
		if err != nil {
			return err
		}
		_, err = writer.Write(buffer[:nRead])
		// log.Printf("%d byte(s) write", nWrite)
		if err != nil {
			return err
		}
	}
}

func readThenWriteWithLogger(reader io.Reader, writer io.Writer, bufferSize uint, logger func(message string), onError func(err error)) {
	err := readThenWrite(reader, writer, bufferSize)
	if err != nil {
		logger(err.Error())
		onError(err)
	}
}

func startReadThenWriteWithLogger(reader io.Reader, writer io.Writer, bufferSize uint, logger func(message string)) chan struct{} {
	ch := make(chan struct{})
	go readThenWriteWithLogger(reader, writer, bufferSize, logger, func(err error) {
		close(ch)
	})
	return ch
}

func startReadWriterExchange(a io.ReadWriter, b io.ReadWriter, bufferSize uint, logger func(message string)) chan struct{} {
	ch := make(chan struct{})
	a2b := startReadThenWriteWithLogger(a, b, bufferSize, logger)
	b2a := startReadThenWriteWithLogger(b, a, bufferSize, logger)
	go func() {
		select {
		case <-a2b:
		case <-b2a:
		}
		close(ch)
	}()
	return ch
}

func StartLogRoutine() {
	go func() {
		for {
			log.Println(<-logChannel)
		}
	}()
}

func SendLog(log string) {
	logChannel <- log
}
