//Package mp4parser parses mp4 file into boxs structure
package mp4parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

//Parse parses the mp4 file , return root box and an error ,if any
func Parse(file *os.File) (*Box, error) {
	rootBox := newBox()

	size, _ := file.Seek(0, 2)
	file.Seek(0, 0)
	rootBox.headerSize = 0
	rootBox.size = uint64(size)
	rootBox.boxType = "root"

	err := parseInnerBox(rootBox, file)

	return rootBox, err
}

//parseBoxHeadr parse b's size and type in header, return an error,if any
func parseBoxHeadr(h *header, file *os.File) (err error) {
	temp := new([8]byte)

	if n, err := file.Read(temp[:8]); err != nil || n != 8 {
		if err == io.EOF {
			return err
		}
		return fmt.Errorf("parseBoxHeadr: %v\n", err)
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

//parseInnerBox parses b's inner boxs , return an error,if any
func parseInnerBox(b *Box, file *os.File) (err error) {

	var endOffset, offsetTmp int64

	if offsetTmp, err = file.Seek(0, 1); err != nil { //record current offset
		return
	}
	endOffset = offsetTmp + int64(b.size-uint64(b.headerSize))

	for {
		innerBoxTmp := newBox()
		if err = parseBoxHeadr(innerBoxTmp.header, file); err != nil {
			if err == io.EOF {
				return
			}
			return fmt.Errorf("parseInnerBox:%v %v\n", b, err)
		}
		innerBoxTmp.nth = b.nth + 1
		b.addInnerBox(innerBoxTmp)

		offsetTmp, err = file.Seek(0, 1) //record  offset
		if innerBoxTmp.isContainer() {

			if err = parseInnerBox(innerBoxTmp, file); err != nil { //
				return
			}

			_, err = file.Seek(offsetTmp, 0) //reset

		}
		innerBoxDataSize := int64(innerBoxTmp.size - uint64(innerBoxTmp.headerSize))
		innerBoxTmp.offset = offsetTmp - int64(innerBoxTmp.headerSize)
		if offsetTmp, err = file.Seek(innerBoxDataSize, 1); err != nil || offsetTmp >= endOffset {
			return
		}
	}

}
