package goroutine

import "github.com/panjf2000/ants/v2"

var pool, _ = ants.NewPool(-1)

// Go starts a goroutine.
func Go(f func()) {
	_ = pool.Submit(f)
}
