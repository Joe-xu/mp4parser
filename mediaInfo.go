package mp4parser

import "time"
import "fmt"

//MediaInfo contain media information
type MediaInfo struct {
	width  uint32 //
	height uint32 //found in tkhd

	creationTime *time.Time
	modifTime    *time.Time
	duration     *time.Duration // result of duration/time_scale(field in mvhd)
}

func (m *MediaInfo) String() string {
	return fmt.Sprintf("creationTime:%v\nmodifTime:%v\nduration:%v", m.creationTime, m.modifTime, m.duration)
}
