package util

import (
	"time"
)

// RunAsyncEvery executa assincronamente uma função f a cada período
// A função pode ser finalizada por meio do canal retornado
func RunAsyncEvery(d time.Duration, f func()) chan<- bool {
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-time.After(d):
				f()
			case <-done:
				return
			}
		}
	}()
	return done
}

// RunEvery executa sincronamente uma função f a cada período d.
// A função pode ser finalizada sinalizando ou fechando o parâmetro done
func RunEvery(done <-chan bool, d time.Duration, f func()) {
	for {
		select {
		case <-time.After(d):
			f()
		case <-done:
			return
		}
	}
}
