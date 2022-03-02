package goroutine

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/yunqi/lighthouse/internal/xlog"
	"go.uber.org/zap"
)

var pool, _ = ants.NewPool(-1, ants.WithLogger(log{}))

type log struct{}

func (l log) Printf(format string, args ...interface{}) {
	xlog.Info("[Ants] logger", zap.String("msg", fmt.Sprintf(format, args...)))
}

// Go starts a goroutine.
func Go(f func()) {
	_ = pool.Submit(f)
}
