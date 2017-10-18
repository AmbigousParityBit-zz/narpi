package NARPImage

import (
	"testing"
)

func Test_cutBytesOfUint16(t *testing.T) {
	var bytesCutTests = []struct {
		in   uint16
		outB bool
		outL uint8
		outR uint8
	}{
		{60116, true, 234, 212},
		{14231, true, 111, 23},
		{16383, true, 255, 255},
		{255, false, 255, 255},
	}

	for _, bt := range bytesCutTests {
		if outB_, outL_, outR_ := cutBytesOfUint16(bt.in); bt.outB != outB_ &&
			bt.outL != outL_ && bt.outR != outR_ {
			t.Errorf("cutBytesOfUint16(%v) => %v,%v,%v, want %v,%v,%v", bt.in, outB_, outL_, outR_,
				bt.outB, bt.outL, bt.outR)
		}
	}
}

func Test_putBytesToUint16(t *testing.T) {
	var bytesPutTests = []struct {
		in  []uint8
		out uint16
	}{
		{[]uint8{122}, 122},
		{[]uint8{234, 212}, 60116},
		{[]uint8{111, 23}, 28439},
		{[]uint8{255, 255}, 65535},
	}

	for _, bt := range bytesPutTests {
		if v := putBytesToUint16(bt.in); v != bt.out {
			t.Errorf("putBytesToUint16(%v) => %v, want %v", bt.in, v, bt.out)
		}
	}
}
