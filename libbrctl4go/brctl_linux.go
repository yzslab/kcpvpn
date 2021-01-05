package libbrctl4go

/*
#cgo CFLAGS: -I${SRCDIR}/libs
#cgo LDFLAGS: -L${SRCDIR}/libs -l:libbridge.a
#include "string.h"
#include "stdlib.h"
#include "libbridge.h"
#include "libbridge_private.h"
*/
import "C"
import (
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"
)

func init() {
	ret := C.br_init()
	if int(ret) < 0 {
		log.Fatalf("br_init(): can not init libbridge")
	}
}

type LinuxBridge struct {
	name             string
	isPersistent     bool
	alreadyDestroied bool
}

func NewLinuxBridge(bridgeName string, isPersistent bool) (*LinuxBridge, error) {
	err := addBridge(bridgeName)
	if err != nil {
		return nil, err
	}
	return OpenLinuxBridge(bridgeName, isPersistent)
}

func OpenLinuxBridge(bridgeName string, isPersistent bool) (*LinuxBridge, error) {
	linuxBridge := LinuxBridge{
		name:             bridgeName,
		isPersistent:     isPersistent,
		alreadyDestroied: false,
	}
	return &linuxBridge, nil
}

func (l *LinuxBridge) AddInterface(interfaceName string) error {
	return addInterface(l.name, interfaceName)
}

func (l *LinuxBridge) DeleteInterface(interfaceName string) error {
	return deleteInterface(l.name, interfaceName)
}

func (l *LinuxBridge) DestroyBridge() error {
	if l.alreadyDestroied == false {
		err := deleteBridge(l.name)
		l.alreadyDestroied = true
		return err
	}
	return nil
}

func (l *LinuxBridge) Close() error {
	if l.isPersistent || l.alreadyDestroied {
		return nil
	}
	l.DestroyBridge()
	return nil
}

func addBridge(name string) error {
	cname := goString2CString(name)
	defer freeCString(cname)
	return retCheck("br_add_bridge", C.br_add_bridge(cname))
}

func deleteBridge(name string) error {
	cname := goString2CString(name)
	defer freeCString(cname)
	return retCheck("br_del_bridge", C.br_del_bridge(cname))
}

func addInterface(bridgeName string, interfaceName string) error {
	cBridgeName := goString2CString(bridgeName)
	defer freeCString(cBridgeName)
	cInterfaceName := goString2CString(interfaceName)
	defer freeCString(cInterfaceName)
	ret := C.br_add_interface(cBridgeName, cInterfaceName)
	return retCheck("br_add_interface", ret)
}

func deleteInterface(bridgeName string, interfaceName string) error {
	cBridgeName := goString2CString(bridgeName)
	defer freeCString(cBridgeName)
	cInterfaceName := goString2CString(interfaceName)
	defer freeCString(cInterfaceName)
	ret := C.br_del_interface(cBridgeName, cInterfaceName)
	return retCheck("br_del_interface", ret)
}

func retCheck(syscallName string, ret C.int) error {
	if int(ret) != 0 {
		return getSyscallError(fmt.Sprintf("%s()", syscallName), int(ret))
	}
	return nil
}

func getSyscallError(syscallName string, errno int) error {
	err := syscall.Errno(errno)
	return os.NewSyscallError(syscallName, err)
}

func goString2CString(goString string) *C.char {
	return C.CString(goString)
}

func freeCString(cString *C.char) {
	C.free(unsafe.Pointer(cString))
}
