package util

type Log func(...any)

func Check[T any](value T, err error) func(log Log) T {

	return func(log Log) T {
		if err != nil {
			log(err)
		}
		return value
	}
}
