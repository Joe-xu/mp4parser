//Package mp4parser parses mp4 file into boxs structure
package mp4parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	seekFromStart = iota
	seekFromCurrent
	seekFromEnd
)

//Parse parses the mp4 file , return root box and an error ,if any
func Parse(file *os.File) (*RootBox, error) {
	rootBox := new(RootBox)
	rootBox.Box = newBox()

	size, _ := file.Seek(0, seekFromEnd)
	file.Seek(0, seekFromStart)
	rootBox.headerSize = 0
	rootBox.size = uint64(size)
	rootBox.boxType = "root"

	err := parseInnerBox(rootBox.Box, file)

	return rootBox, err
}

//parseBoxHeadr parses b's size and type in header, return an error,if any,and resumes file seeker
func parseBoxHeadr(h *header, file *os.File) (err error) {
	savedOffset, err := file.Seek(0, seekFromCurrent)
	if err != nil {
		return
	}
	defer file.Seek(savedOffset, seekFromStart)

	temp := new([8]byte)

	if n, err := file.Read(temp[:8]); err != nil || n != 8 {
		if err == io.EOF {
			return err
		}
		return fmt.Errorf("parseBoxHeadr: %d-th byte  %v\n", n, err)
	}

	h.boxType = string(temp[4:8])

	size := uint64(binary.BigEndian.Uint32(temp[:4]))

	if size == 1 { //if size == 1 get largeSize in next 8 bytes
		if n, err := file.Read(temp[:8]); err != nil || n != 8 {
			if err == io.EOF {
				return err
			}
			return fmt.Errorf("parseBoxHeadr: %d-th byte  %v\n", n, err)
		}

		h.headerSize = largeHeaderSize
		size = binary.BigEndian.Uint64(temp[:8])

	}

	if size <= 0 {
		return fmt.Errorf("parseBoxHeadr: %v \t temp:%v", h, temp)
	}

	h.size = size

	return
}

//parseInnerBox parses b's inner boxs , return an error,if any,and resumes file seeker
func parseInnerBox(b *Box, file *os.File) (err error) {
	savedOffset, err := file.Seek(0, seekFromCurrent)
	defer file.Seek(savedOffset, seekFromStart)
	var endOffset, offsetTmp int64

	offsetTmp, err = file.Seek(int64(b.headerSize), seekFromCurrent) //skip  box header
	if err != nil {
		return
	}

	endOffset = savedOffset + int64(b.size-uint64(b.headerSize))

	for {
		innerBoxTmp := newBox()
		innerBoxTmp.nth = b.nth + 1
		innerBoxTmp.offset = offsetTmp
		b.addInnerBox(innerBoxTmp)
		if err = parseBoxHeadr(innerBoxTmp.header, file); err != nil {
			if err == io.EOF {
				return
			}
			return fmt.Errorf("parseInnerBox:%v %v\n", b, err)
		}

		if innerBoxTmp.isContainer() {

			if err = parseInnerBox(innerBoxTmp, file); err != nil { //
				return
			}

		}

		scanBoxData(innerBoxTmp, file)

		if offsetTmp, err = file.Seek(int64(innerBoxTmp.size), seekFromCurrent); err != nil || offsetTmp >= endOffset {
			return
		}
	}

}

//scanBoxData scans box data from file, return an error , if any ,and resumes file seeker
func scanBoxData(b *Box, file *os.File) (err error) {
	savedOffset, _ := file.Seek(0, seekFromCurrent)
	defer file.Seek(savedOffset, seekFromStart)

	switch b.boxType {
	case "stsc":
		stscBox := newSTSC(b)
		err = stscBox.scan(file)
		// fmt.Print(stscBox)
	case "stco":
		stcoBox := newSTCO(b)
		err = stcoBox.scan(file)
		// fmt.Print(stcoBox)
	default:
		return fmt.Errorf("unexcepted box type:%v", b.boxType)
	}
	return
}
