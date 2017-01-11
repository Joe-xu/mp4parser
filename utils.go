package mp4parser

import "time"

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
