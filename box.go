package mp4parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
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

//RootBox contains all box in file
type RootBox struct {
	*Box
	videoTracks []*trak
	soundTracks []*trak
	// hintTracks  []*trak
}

//box header
//box data size = size - headerSize
type header struct {
	size       uint64 //size of box include header
	headerSize int
	boxType    string
}

//newBox return new Box
func newBox() *Box {

	return &Box{
		header:    &header{headerSize: normalHeaderSize},
		innerBoxs: make(map[string][]*Box),
	}
}

//newRootBox return new RootBox
func newRootBox() *RootBox {
	return &RootBox{
		Box:         newBox(),
		videoTracks: make([]*trak, 0, 1),
		soundTracks: make([]*trak, 0, 1),
	}
}

//newInnerBox  add new innerbox
func (b *Box) addInnerBox(inner *Box) {
	if _, ok := b.innerBoxs[inner.boxType]; !ok { //not exist yet
		b.innerBoxs[inner.boxType] = make([]*Box, 0, 1)

	}
	b.innerBoxs[inner.boxType] = append(b.innerBoxs[inner.boxType], inner)

}

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
	if strings.TrimSpace(b.boxType) == "" {
		return false
	}
	return strings.Contains("moov trak mdia minf dinf stbl", b.boxType)
}

//=====specified box types=======//

//boxs contain meta data
type dataBox interface {
	scan(*os.File) error
}

//moov movie
type moov struct {
	*Box
}

//mvhd  movie header
/**
*header				normalHeaderSize/largeHeaderSize
*version 			1
*flags 				3
*creation_time		4
*modification_time	4
*time_scale			4
*duration 			4
*rate 				4
*volume 			2
*_reserved 			10
*matrix 			36
*pre-defined		24
*next_track_id		4
 */
type mvhd struct {
	*Box
	version      uint8
	creationTime *time.Time
	modifTime    *time.Time
	duration     uint32
	timeScale    uint32
	// nextTrackID uint32
}

func newMVHD(b *Box) *mvhd {
	return &mvhd{
		Box: b,
	}
}

//scan mvhd data in file , return an error ,if any , and resume file seeker
func (b *mvhd) scan(file *os.File) (err error) {
	savedOffset, _ := file.Seek(0, seekFromCurrent)
	defer file.Seek(savedOffset, seekFromStart)

	temp := new([16]byte)

	_, err = file.ReadAt(temp[:], b.offset+int64(b.headerSize)+4) //read at creation_time
	if err != nil {
		return
	}

	b.creationTime, err = getFixTime(binary.BigEndian.Uint32(temp[:4]))
	if err != nil {
		return
	}

	b.modifTime, err = getFixTime(binary.BigEndian.Uint32(temp[4:8]))
	if err != nil {
		return
	}

	b.timeScale = binary.BigEndian.Uint32(temp[8:12])

	b.duration = binary.BigEndian.Uint32(temp[12:16])

	return
}

//trak track
type trak struct {
	*Box
}

func newTRAK(b *Box) *trak {
	return &trak{
		Box: b,
	}
}

//tkhd track header
const (
	trackEnabled   = 0x000001
	trackInMovie   = 0x000002
	trackInPreview = 0x000004
) //track flags
/**
*header				normalHeaderSize/largeHeaderSize
*version			1
*flags				3
*creation_time		4
*modification_time	4
*track_id			4
*_reserved			4
*duration			4
*_reserved			8
*layer				2
*alternate_group	2
*volume				2	//[8.8]  e.g. max value 0x0100 = 1.0
*_reserved			2
*matrix				36
*width				4	//[16.16]
*height				4
 */
type tkhd struct {
	*Box
	// version      uint8
	flags        uint32
	creationTime *time.Time
	modifTime    *time.Time
	duration     uint32
	// trackID      uint32
	volume float64
	width  float64
	height float64
}

func newTKHD(b *Box) *tkhd {
	return &tkhd{
		Box: b,
	}
}

//scan tkhd data in file , return an error ,if any , and resume file seeker
func (b *tkhd) scan(file *os.File) (err error) {
	savedOffset, _ := file.Seek(0, seekFromCurrent)
	defer file.Seek(savedOffset, seekFromStart)

	temp := new([40]byte)

	_, err = file.ReadAt(temp[:40], b.offset+int64(b.headerSize)) //read at version
	if err != nil {
		return
	}

	// if version := binary.BigEndian.Uint16(temp[:1]); version == 1 {
	// 	panic("scan tkhd: version = 1")
	// }

	// b.flags=temp[1:4]
	b.creationTime, err = getFixTime(binary.BigEndian.Uint32(temp[4:8]))
	if err != nil {
		return
	}

	b.modifTime, err = getFixTime(binary.BigEndian.Uint32(temp[8:12]))
	if err != nil {
		return
	}
	b.duration = binary.BigEndian.Uint32(temp[20:24])
	b.volume, err = dottedNotationToF(temp[36:38])
	if err != nil {
		return
	}

	_, err = file.ReadAt(temp[:8], b.offset+int64(b.headerSize)+76) //read at width
	if err != nil {
		return
	}
	b.width, err = dottedNotationToF(temp[:4])
	if err != nil {
		return
	}
	b.height, err = dottedNotationToF(temp[4:8])

	return
}

//mdia  media
type mdia struct {
	*Box
}

//mdhd media header
/**
*header					normalHeaderSize/largeHeaderSize
*version				1
*flags					3
*creation_time			4
*modification_time		4
*time_scale				4
*duration				4
*language				2
*pre-defined			2
 */
type mdhd struct {
	*Box
	// version uint8
	// creationTime *time.Time
	// modifTime    *time.Time
	// duration  *time.Duration
	timeScale uint32
}

func newMDHD(b *Box) *mdhd {
	return &mdhd{
		Box: b,
	}
}

func (b *mdhd) scan(file *os.File) (err error) {
	savedOffset, _ := file.Seek(0, seekFromCurrent)
	defer file.Seek(savedOffset, seekFromStart)

	temp := make([]byte, 0, 4)

	_, err = file.ReadAt(temp[:4], b.offset+int64(b.headerSize)+12)
	if err != nil {
		return
	}

	b.timeScale = binary.BigEndian.Uint32(temp[:4])

	return
}

//hdlr handler reference
/**
*header				normalHeaderSize/largeHeaderSize
*version			1
*flags				3
*pre-defined		4
*handler_type		4
*_reserved			12
*name				end with '\0' in file
 */
type hdlr struct {
	*Box
	// version     uint8
	handlerType string //"vide"/"soun"/"hint"
	name        string
}

func newHDLR(b *Box) *hdlr {
	return &hdlr{
		Box: b,
	}
}

func (b *hdlr) scan(file *os.File) (err error) {
	savedOffset, _ := file.Seek(0, seekFromCurrent)
	defer file.Seek(savedOffset, seekFromStart)

	temp := make([]byte, 0, 4)

	_, err = file.ReadAt(temp[:4], b.offset+int64(b.headerSize)+8)
	if err != nil {
		return
	}

	b.handlerType = string(temp[:4])

	_, err = file.Seek(b.offset+int64(b.headerSize)+24, seekFromStart)

	if err != nil {
		return
	}
	//read name
	temp = temp[:0]
	t := make([]byte, 1)

	for {

		n, err := file.Read(t)
		if err != nil || n == 0 {
			return err
		}

		if t[0] == '\000' {
			break
		}
		temp = append(temp, t[0])
	}

	b.name = string(temp)

	return
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

//stsc sample to chunk
/**
*header			normalHeaderSize/largeHeaderSize
*version 		1
*flags 			3
*entryCount		4
 */
type stsc struct {
	*Box
	// version uint8
	// flag uint32 //uint24
	entryCount uint32
	entrys     []*stscEntry
}

type stscEntry struct {
	firstChunk      uint32
	samplesPerChunk uint32
	sampleDescIndex uint32 //index to find sample description in stsd
}

func newSTSC(b *Box) *stsc {
	return &stsc{
		Box: b,
		// entrys:make([]*stscEntry,1),
	}
}

func (b *stsc) String() string {
	buffer := new(bytes.Buffer)

	buffer.WriteString(fmt.Sprintf("entry count: %d\n", b.entryCount))
	buffer.WriteString("\t\tfirstChunk\tsamplesPerChunk\tsampleDescIndex\n")
	for i, entry := range b.entrys {

		buffer.WriteString(
			fmt.Sprintf("%d\t%d\t%d\t%d\n",
				i, entry.firstChunk, entry.samplesPerChunk, entry.sampleDescIndex))
	}

	return buffer.String()
}

//scan stsc data in file , return an error ,if any , and resume file seeker
func (b *stsc) scan(file *os.File) (err error) {
	savedOffset, _ := file.Seek(0, seekFromCurrent)
	defer file.Seek(savedOffset, seekFromStart)

	temp := new([12]byte)

	_, err = file.Seek(b.offset+int64(b.headerSize)+4, seekFromStart)
	if err != nil {
		return
	}

	_, err = file.Read(temp[:4]) //read entry count
	if err != nil {
		return
	}

	b.entryCount = binary.BigEndian.Uint32(temp[:4])
	if b.entrys == nil {
		b.entrys = make([]*stscEntry, 0, b.entryCount)
	}

	for i := uint32(0); i < b.entryCount; i++ {

		_, err = file.Read(temp[:])
		if err != nil {
			return
		}
		b.entrys = append(b.entrys, &stscEntry{
			firstChunk:      binary.BigEndian.Uint32(temp[:4]),
			samplesPerChunk: binary.BigEndian.Uint32(temp[4:8]),
			sampleDescIndex: binary.BigEndian.Uint32(temp[8:12]),
		})

	}

	return
}

//stco chunk offset
/**
*header			normalHeaderSize/largeHeaderSize
*version		1
*flags 			3
*entryCount		4
*chunkOffset	entryCount*4
 */
type stco struct {
	*Box
	// version     uint8
	entryCount  uint32
	chunkOffset []uint32
}

func newSTCO(b *Box) *stco {
	return &stco{
		Box: b,
	}
}

func (b *stco) String() string {
	buffer := new(bytes.Buffer)

	buffer.WriteString(fmt.Sprintf("entry count: %d\n", b.entryCount))

	for _, offset := range b.chunkOffset {

		buffer.WriteString(
			fmt.Sprintf("\t%d\n", offset))
	}

	return buffer.String()
}

//scan stco data in file , return an error ,if any , and resume file seeker
func (b *stco) scan(file *os.File) (err error) {
	savedOffset, _ := file.Seek(0, seekFromCurrent)
	defer file.Seek(savedOffset, seekFromStart)

	temp := new([4]byte)
	_, err = file.Seek(b.offset+int64(b.headerSize)+4, seekFromStart)

	_, err = file.Read(temp[:]) //read entry count
	if err != nil {
		return
	}

	b.entryCount = binary.BigEndian.Uint32(temp[:])
	if b.chunkOffset == nil {
		b.chunkOffset = make([]uint32, 0, b.entryCount)
	}

	for i := uint32(0); i < b.entryCount; i++ {

		_, err = file.Read(temp[:])
		if err != nil {
			return
		}
		b.chunkOffset = append(b.chunkOffset, binary.BigEndian.Uint32(temp[:]))

	}

	return
}
