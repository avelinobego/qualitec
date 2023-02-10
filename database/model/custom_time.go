package model

import (
	"time"
)

type QualitecTime struct {
	time.Time
}

func (c *QualitecTime) Scan(value interface{}) (e error) {

	switch t := value.(type) {
	case string:
		c.Time, e = toDateTime(t)
	case []uint8:
		c.Time, e = toDateTime(string(t))
	case time.Time:
		c.Time = t
	default:
		c.Time = time.Time{}
	}
	return
}

func toDateTime(v string) (result time.Time, err error) {
	result = time.Time{}
	var mapa map[string]string
	if mapa, err = ExtractDateTime(v); err != nil {
		return
	}
	result, err = time.Parse("2006-01-0215:04:05.999999999", mapa["date"]+mapa["time"])
	return
}
