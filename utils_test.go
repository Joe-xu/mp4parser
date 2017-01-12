package mp4parser

import "testing"
import "math"

func TestDottedNotationToF(t *testing.T) {
	const eps = 1e-7

	tests := [...]struct {
		input []byte
		want  float64
	}{
		{[]byte{0xFF, 0x11}, 255.17},
		{[]byte{0x01, 0x00}, 1.0},
		{[]byte{0x01, 0x04}, 1.4},
		{[]byte{0x23, 0x56}, 35.86},
		{[]byte{0x23, 0x56, 0xff, 0x01}, 9046.65281},
	}

	for _, test := range tests {
		got, _ := dottedNotationToF(test.input)
		if math.Abs(got-test.want) > eps {
			t.Errorf("intput %v want %v , got %v", test.input, test.want, got)
		}
	}
}

func TestByteToUint(t *testing.T) {

	tests := [...]struct {
		input []byte
		want  uint64
	}{
		{[]byte{0x00}, 0},
		{[]byte{0x00, 0x80}, 128},
		{[]byte{0xFF}, 255},
		{[]byte{0xFF, 0x11}, 65297},
		{[]byte{0x01, 0x00}, 256},
	}

	for _, test := range tests {
		got := byteToUint(test.input)
		if got != test.want {
			t.Errorf("intput %v want %v , got %v", test.input, test.want, got)
		}

	}
}

func TestByteToUintOverFlow(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Error("should panic when overflow")
		}
	}()

	byteToUint([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
}
