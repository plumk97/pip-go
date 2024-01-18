package tests_test

import (
	"encoding/hex"
	"testing"

	"github.com/plumk97/pip-go/types"
)

func TestIPv4Header(t *testing.T) {
	if bytes, err := hex.DecodeString("45000034a2bc40003406d0afcc2c461cc0a80067"); err == nil {
		hdr := types.IPHdr(bytes)
		// header := pipgo.NewIPHeader(bytes)
		hdr.SetVersion(6)
		t.Log(hdr.Version())
		t.Log(hdr.IHL())
	} else {
		t.Error(err)
	}
}

func TestIPv6Header(t *testing.T) {
	if bytes, err := hex.DecodeString("6000000000200001fe80000000000000caeaf8fffee1356dff0200000000000000000000000000013a00050200000000"); err == nil {
		hdr := types.IP6Hdr(bytes)
		t.Logf("Version: %d", hdr.Version())
		t.Logf("TrafficClass: %d", hdr.TrafficClass())
		t.Logf("Flow: %d", hdr.Flow())
		t.Logf("PayloadLen: %d", hdr.PayloadLen())
		t.Logf("NextHeader: %d", hdr.NextHeader())
		t.Logf("HopLimit: %d", hdr.HopLimit())
		t.Logf("Src: %s", hdr.Src().String())
		t.Logf("Dst: %s", hdr.Dst().String())
	} else {
		t.Error(err)
	}
}
