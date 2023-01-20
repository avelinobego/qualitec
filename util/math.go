package util

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0 // return correctly abs(-0)
	}
	return x
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func Min64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func Max64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
