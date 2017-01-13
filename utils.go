package mp4parser

import (
	"fmt"
	"time"
)

//getFixTime excepts number of secondselapsed from 1904-Jan-01 00:00:00 UTC ,
//  return fix time and an error,if any
func getFixTime(sec uint32) (*time.Time, error) {

	fixZeroTime, err := time.Parse("2006-Jan-02", "1904-Jan-01")
	if err != nil {
		return nil, err
	}
	t := time.Unix(int64(sec)+fixZeroTime.Unix(), 0).In(time.UTC)

	return &t, nil
}

//dottedNotationToF convert dotted notation ,[8.8]/[16.16] , to float
//	len(n)<=4 and should not be odd length
//  e.g.
//	0xFF11 is the number 255.17
//	0x0104 is the number 1.4
//	0x2356 is the number 35.86
func dottedNotationToF(n []byte) (float64, error) {

	nLen := len(n)
	if nLen == 0 {
		return 0, nil
	}
	if nLen&1 == 1 || nLen > 4 {
		return 0, fmt.Errorf("dottedNotationToF:invalid length of bytes = %d", nLen)
	}

	head := byteToUint(n[:nLen/2])
	tail := byteToUint(n[nLen/2:])

	base := uint64(1)
	for tail/base > 0 {
		base *= 10
	}

	t := float64(tail) / float64(base)

	return float64(head) + t, nil
}

//byteToUint convert bytes into uint64
func byteToUint(buf []byte) uint64 {
	var res uint64
	var count int

	for _, b := range buf {
		res = uint64(b) | res<<8

		if res > 0 {
			count++
			if count > 8 {
				panic("byteToUint:overflow")
			}

		}

	}

	return res

}
