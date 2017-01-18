//Package mp4parser parses mp4 file into boxs structure
package mp4parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	seekFromStart = iota
	seekFromCurrent
	seekFromEnd
)

//Parser parses file into media meta infos
type Parser struct {
	file      *os.File
	rootBox   *RootBox
	dataBoxs  map[string][]dataBox
	mediaInfo *MediaInfo
}

//NewParser return new Parser
func NewParser(file *os.File) *Parser {
	return &Parser{
		file:      file,
		rootBox:   newRootBox(),
		dataBoxs:  make(map[string][]dataBox),
		mediaInfo: new(MediaInfo),
	}
}

//Parse parses the mp4 file , return mediaInfo and an error ,if any
func (p *Parser) Parse() (*MediaInfo, error) {

	fileInfo, err := p.file.Stat() //get file size
	if err != nil {
		return nil, err
	}

	p.rootBox.headerSize = 0
	p.rootBox.size = uint64(fileInfo.Size())
	p.rootBox.boxType = "root"

	p.file.Seek(0, seekFromStart) //ensure parsing start from file head
	err = p.parseInnerBox(p.rootBox.Box)

	err = rangeBox(p.rootBox.Box, p.scanBoxData) // //TODO:handle err

	return p.mediaInfo, err
}

//parseBoxHeadr parses b's size and type in header, return an error,if any,and resumes file seeker
func (p *Parser) parseBoxHeadr(h *header) (err error) {
	savedOffset, err := p.file.Seek(0, seekFromCurrent)
	if err != nil {
		return
	}
	defer p.file.Seek(savedOffset, seekFromStart)

	temp := new([8]byte)

	if n, err := p.file.Read(temp[:8]); err != nil || n != 8 {
		if err == io.EOF {
			return err
		}
		return fmt.Errorf("parseBoxHeadr: %d-th byte  %v\n", n, err)
	}

	h.boxType = string(temp[4:8])

	size := uint64(binary.BigEndian.Uint32(temp[:4]))

	if size == 1 { //if size == 1 get largeSize in next 8 bytes
		if n, err := p.file.Read(temp[:8]); err != nil || n != 8 {
			if err == io.EOF {
				return err
			}
			return fmt.Errorf("parseBoxHeadr: %d-th byte  %v\n", n, err)
		}

		h.headerSize = largeHeaderSize
		size = binary.BigEndian.Uint64(temp[:8])

	}

	if size <= 0 {
		panic(fmt.Errorf("parseBoxHeadr: size <=0 : %v \t temp:%v", h, temp))
	}

	h.size = size

	return
}

//parseInnerBox parses b's inner boxs , return an error,if any,and resumes file seeker
func (p *Parser) parseInnerBox(b *Box) (err error) {
	savedOffset, err := p.file.Seek(0, seekFromCurrent)
	defer p.file.Seek(savedOffset, seekFromStart)
	var endOffset, offsetTmp int64

	offsetTmp, err = p.file.Seek(int64(b.headerSize), seekFromCurrent) //skip  box header
	if err != nil {
		return
	}

	endOffset = savedOffset + int64(b.size-uint64(b.headerSize))

	for {
		innerBox := newBox()
		innerBox.nth = b.nth + 1
		innerBox.offset = offsetTmp
		if err = p.parseBoxHeadr(innerBox.header); err != nil {
			if err == io.EOF {
				return
			}
			return fmt.Errorf("parseInnerBox:%v %v\n", b, err)
		}
		b.addInnerBox(innerBox)

		if innerBox.isContainer() {

			if err = p.parseInnerBox(innerBox); err != nil { //
				return
			}

		}

		if offsetTmp, err = p.file.Seek(int64(innerBox.size), seekFromCurrent); err != nil || offsetTmp >= endOffset {
			return
		}
	}

}

//rangeBox run func do for each box contained below b and return an error , if any
func rangeBox(b *Box, do func(*Box) error) (err error) {
	for _, innerBoxSlice := range b.innerBoxs {
		for _, inner := range innerBoxSlice {

			err = do(inner)
			if err != nil {
				return err
			}

			err = rangeBox(inner, do)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

//scanBoxData scans box data from file, return an error , if any
func (p *Parser) scanBoxData(b *Box) (err error) {
	// savedOffset, _ := p.file.Seek(0, seekFromCurrent)
	// defer p.file.Seek(savedOffset, seekFromStart)

	switch b.boxType {

	case "trak": //get track data
		tkhdBox := newTKHD(b.innerBoxs["tkhd"][0])
		err = tkhdBox.scan(p.file)

		hdlrBox := newHDLR(b.innerBoxs["mdia"][0].innerBoxs["hdlr"][0])
		err = hdlrBox.scan(p.file)

		if hdlrBox.handlerType == "vide" {
			p.mediaInfo.height = tkhdBox.height
			p.mediaInfo.width = tkhdBox.width
			p.rootBox.videoTracks = append(p.rootBox.videoTracks, newTRAK(b))
		} else if hdlrBox.handlerType == "soun" {
			mdhdBox := newMDHD(b.innerBoxs["mdia"][0].innerBoxs["mdhd"][0])
			mdhdBox.scan(p.file)

			p.mediaInfo.soundSamplingRate = mdhdBox.timeScale
			p.rootBox.soundTracks = append(p.rootBox.soundTracks, newTRAK(b))
		}

	case "mvhd":
		mvhdBox := newMVHD(b)
		err = mvhdBox.scan(p.file)

		duration, _ := time.ParseDuration(fmt.Sprintf("%ds", mvhdBox.duration/mvhdBox.timeScale))
		p.mediaInfo.duration = &duration
		p.mediaInfo.creationTime = mvhdBox.creationTime
		p.mediaInfo.modifTime = mvhdBox.modifTime

		p.dataBoxs[b.boxType] = append(p.dataBoxs[b.boxType], mvhdBox)
	case "stsc":
		stscBox := newSTSC(b)
		err = stscBox.scan(p.file)
		p.dataBoxs[b.boxType] = append(p.dataBoxs[b.boxType], stscBox)

	case "stco":
		stcoBox := newSTCO(b)
		err = stcoBox.scan(p.file)
		p.dataBoxs[b.boxType] = append(p.dataBoxs[b.boxType], stcoBox)

		// default:
		// 	return fmt.Errorf("unexcepted box type:%v", b.boxType)
	}
	return
}
