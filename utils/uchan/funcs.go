package uchan

// DrainForever will drain the channels in separate go routines in a loop forever
// Intended for tests only
func DrainForever[T any](chs ...<-chan T) {
	for _, ch := range chs {
		go func() {
			for {
				<-ch
			}
		}()
	}
}

// Nudger can be used to make a goroutine ('A') sleep, and have another goroutine ('B') wake him up
// A will not block if B is not asleep.
type Nudger struct {
	C chan struct{} // Receive on C to sleep
}

func NewNudger() *Nudger {
	return &Nudger{make(chan struct{})}
}

// Nudge wakes up the waiting thread if any. Non blocking.
func (w Nudger) Nudge() {
	select {
	case w.C <- struct{}{}:
	default:
	}
}
