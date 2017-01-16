package mp4parser

import (
	"math"
	"os"
	"testing"
	"time"
)

var testFile *os.File

func init() {
	testFile, _ = os.Open(`./sample/sample.mp4`)
}

func TestNewParser(t *testing.T) {
	p := NewParser(testFile)
	if p.file == nil || p.rootBox == nil || p.dataBoxs == nil || p.mediaInfo == nil {
		t.Errorf("got %#v", p)
	}
}

func TestParse(t *testing.T) {
	const eps = 1e-7

	want := &MediaInfo{
		width:  560,
		height: 320,
	}
	t1 := time.Date(2010, 3, 20, 21, 29, 11, 0, time.UTC)
	want.creationTime = &t1
	t2 := time.Date(2010, 3, 20, 21, 29, 12, 0, time.UTC)
	want.modifTime = &t2

	t3, _ := time.ParseDuration("5s")
	want.duration = &t3

	p := NewParser(testFile)
	got, _ := p.Parse()

	if math.Abs(got.width-want.width) > eps ||
		math.Abs(got.height-want.height) > eps ||
		!got.creationTime.Equal(*want.creationTime) ||
		!got.modifTime.Equal(*want.modifTime) ||
		*got.duration != *want.duration {
		t.Errorf("want:\n%v\ngot:\n%v", want, got)
	}

}

func BenchmarkParse(b *testing.B) {
	p := NewParser(testFile)
	for i := 0; i < b.N; i++ {
		p.Parse()
	}
}
