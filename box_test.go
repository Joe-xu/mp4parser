package mp4parser

import (
	"testing"
)

func TestNewBox(t *testing.T) {
	b := newBox()
	if b == nil {
		t.Error("newBox(),got nil\n")
	}

}

func TestIsContainer(t *testing.T) {
	b := newBox()
	tests := [...]struct {
		boxType string
		want    bool
	}{
		{"moov", true},
		{"trak", true},
		{"mdia", true},
		{"minf", true},
		{"dinf", true},
		{"stbl", true},
		{"stsd", false},
		{"fytp", false},
	}

	for _, test := range tests {
		b.boxType = test.boxType
		if b.isContainer() != test.want {
			t.Errorf("type:%q,isContainer(),want %t", test.boxType, test.want)
		}
	}

}
