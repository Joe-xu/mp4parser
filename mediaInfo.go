package mp4parser

import "time"
import "fmt"

//MediaInfo contain media information
type MediaInfo struct {
	width             float64 //
	height            float64 //found in tkhd
	soundSamplingRate uint32

	creationTime *time.Time
	modifTime    *time.Time
	duration     *time.Duration // result of duration/time_scale(field in mvhd)
}

func (m *MediaInfo) String() string {
	return fmt.Sprintf(
		"creationTime:%v\nmodifTime:%v\nduration:%v\nwidth:%.2f\theight:%.2f\tsound samlping rate:%dHz",
		m.creationTime, m.modifTime, m.duration, m.width, m.height, m.soundSamplingRate)
}

func (m *MediaInfo) Width() float64 {
	return m.width
}

func (m *MediaInfo) Height() float64 {
	return m.height
}

func (m *MediaInfo) SamplingRate() uint32 {
	return m.soundSamplingRate
}

func (m *MediaInfo) CreationTime() time.Time {
	return *m.creationTime
}

func (m *MediaInfo) ModifiedTime() *time.Time {
	return m.modifTime
}

func (m *MediaInfo) Duration() *time.Duration {
	return m.duration
}
