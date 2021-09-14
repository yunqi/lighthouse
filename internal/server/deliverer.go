package server

import (
	"github.com/yunqi/lighthouse/internal/persistence/message"
)

// Deliverer 表示具备投递信息功能的一类对象
type Deliverer interface {
	Deliver(message message.Message) error
}
