package util

import "strconv"

func SliceStrToInt16(s []string, ignoreErr bool) (result []int16) {
	for _, v := range s {
		converted, err := strconv.ParseInt(v, 10, 16)
		if !ignoreErr || (ignoreErr && err == nil) {
			result = append(result, int16(converted))
		}
	}
	return
}

func SliceIntersectInt16(s1 []int16, s2 []int16) (result []int16) {
	values := make(map[int16]bool, 0)
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if v1 == v2 {
				_, ok := values[v1]
				if !ok {
					values[v1] = true
					result = append(result, v1)
				}
			}
		}
	}
	return
}
