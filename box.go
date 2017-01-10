package mp4parser

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	normalHeaderSize = 8
	largeHeaderSize  = 16 //only if  size (field of header) == 1
)

//Box  fundation unit in mp4
type Box struct {
	*header       // header / largeHeader
	nth       int //n-th inner Box
	innerBoxs map[string][]*Box
	offset    int64 //offset in file
}

//box header
//box data size = size - headerSize
type header struct {
	size       uint64 //size of box include header
	headerSize int
	boxType    string
}

//mdat media data container
type mdat []byte

//newBox create Box
func newBox() (b *Box) {
	b = &Box{
		header:    &header{headerSize: normalHeaderSize},
		innerBoxs: make(map[string][]*Box),
	}

	return
}

//newInnerBox  add new innerbox
func (b *Box) addInnerBox(inner *Box) {
	if _, ok := b.innerBoxs[inner.boxType]; !ok { //not exist yet
		b.innerBoxs[inner.boxType] = make([]*Box, 0, 1)

	}
	b.innerBoxs[inner.boxType] = append(b.innerBoxs[inner.boxType], inner)

}

//String stringly Box and inner boxs
func (b *Box) String() string {
	buffer := new(bytes.Buffer)
	buffer.WriteString(fmt.Sprintf(
		"%*stype: %s\tsize: %d\toffset: %d\n",
		b.nth, "", b.boxType, b.size, b.offset))

	if len(b.innerBoxs) > 0 {
		buffer.WriteString(fmt.Sprintf("%*sinner boxs\n", b.nth*2-1, "")) //indent
	}
	for _, innerBoxSlice := range b.innerBoxs {
		for _, innerBox := range innerBoxSlice {
			buffer.WriteString(fmt.Sprintf("%*s%v", b.nth, "", innerBox))
		}
	}

	return buffer.String()
}

func (h *header) String() string {
	return fmt.Sprintf("size: %d\theader size:%d\ttype: %s", h.size, h.headerSize, h.boxType)
}

//
func (b *Box) isContainer() bool {
	return strings.Contains("moov trak mdia minf dinf stbl", b.boxType)
}

//============
