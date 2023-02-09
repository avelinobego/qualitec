package model

import "time"

type QualitecTime struct {
	time.Time
}

func (c *QualitecTime) Scan(value interface{}) error {
	switch t := value.(type) {
	case string:
		if v, e := time.Parse(time.RFC3339, t); e == nil {
			c.Time = v
		} else {
			return e
		}
	case []uint8:
		if v, e := time.Parse(time.RFC3339, string(t)); e == nil {
			c.Time = v
		} else {
			return e
		}
	case time.Time:
		c.Time = t
	default:
		c.Time = time.Now()
	}
	return nil
}
