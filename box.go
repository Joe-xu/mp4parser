package mp4parser

import (
	"bytes"
	"fmt"
	"strings"
	"time"
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

//RootBox contains all box in file and media status
type RootBox struct {
	*Box
	tracks       []*trak
	creationTime time.Time
	modifTime    time.Time
	duration     time.Time
}

//box header
//box data size = size - headerSize
type header struct {
	size       uint64 //size of box include header
	headerSize int
	boxType    string
}

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

//String stringify Box and inner boxs
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

func (r *RootBox) String() string {
	return r.Box.String()
}

func (h *header) String() string {
	return fmt.Sprintf("size: %d\theader size:%d\ttype: %s", h.size, h.headerSize, h.boxType)
}

//
func (b *Box) isContainer() bool {
	return strings.Contains("moov trak mdia minf dinf stbl", b.boxType)
}

//=====specified box types=======//

//moov movie
type moov struct {
	*Box
}

//mvhd  movie header
type mvhd struct {
	*Box
	version      uint8
	creationTime time.Time
	modifTime    time.Time
	duration     time.Time
	// timeScale    uint32
	nextTrackID uint32
}

//trak track
type trak struct {
	*Box
}

//tkhd track header
const (
	trackEnabled   = 0x000001
	trackInMovie   = 0x000002
	trackInPreview = 0x000004
) //track flags

type tkhd struct {
	*Box
	version      uint8
	flags        uint32
	creationTime time.Time
	modifTime    time.Time
	duration     time.Time
	trackID      uint32
	width        uint32
	height       uint32
}

//mdia  media
type mdia struct {
	*Box
}

//mdhd media header
type mdhd struct {
	*Box
	version      uint8
	creationTime time.Time
	modifTime    time.Time
	duration     time.Time
	// timeScale    uint32
}

//hdlr handler reference
type hdlr struct {
	*Box
	version     uint8
	handlerType uint32 //"vide"/"soun"/"hint"
	name        string //end with '\0' in file
}

//minf media information
type minf struct {
	*Box
}

//media information header include vmhd/smhd/hmhd/nmhd

//vmhd video media header
type vmhd struct {
	*Box
	version uint8
	// opcolor [3]uint16
}

//smhd sound media header
type smhd struct {
	*Box
	version uint8
	// balance uint16
}

//stbl  sample table
type stbl struct {
	*Box
}
