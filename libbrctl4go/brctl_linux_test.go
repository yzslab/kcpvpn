package libbrctl4go

import (
	"testing"
)

func TestAddBridge(t *testing.T) {
	err := addBridge("testbr0")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteBridge(t *testing.T) {
	err := deleteBridge("testbr0")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLinuxBridge_AddDeleteInterface(t *testing.T) {
	l, err := NewLinuxBridge("testbr0", false)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	err = l.AddInterface("tap0")
	if err != nil {
		t.Fatal(err)
	}
	err = l.DeleteInterface("tap0")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetSyscallError(t *testing.T) {
	println(getSyscallError("read()", 0xb).Error())
}