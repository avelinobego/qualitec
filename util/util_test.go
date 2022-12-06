package util

import (
	"testing"
	"time"
)

func TestRunAsyncEvery(t *testing.T) {
	done1 := make(chan bool)
	done2 := RunAsyncEvery(time.Millisecond, func() {
		done1 <- true
	})

	i := 0
	for {
		select {
		case <-done1:
			i++
			if i > 10 {
				close(done2)
				return
			}

		case <-time.After(time.Millisecond * 100):
			t.Error("Error")
		}
	}
}

func TestRunEvery(t *testing.T) {
	i := 0
	done := make(chan bool)
	RunEvery(done, time.Millisecond, func() {
		i++
		if i == 10 {
			close(done)
		}
	})

	if i != 10 {
		t.Errorf("Expected %d, got %d", 10, i)
	}
}
