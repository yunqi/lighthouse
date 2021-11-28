/*
 *    Copyright 2021 chenquan
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package server

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
	"github.com/yunqi/lighthouse/internal/persistence/session"
	"github.com/yunqi/lighthouse/internal/persistence/session/memery"
	"go.uber.org/zap"
	"net"
	"os"
	"time"
)

var poolGo, _ = ants.NewPool(ants.DefaultAntsPoolSize)

type (
	Server interface {
		Stop(ctx context.Context) error
		Run() error
	}
	Option func(server *Options)

	Options struct {
		tcpListen       string
		websocketListen string
		storeType       string
	}
	server struct {
		tcpListen         string
		websocketListen   string
		tcpListener       net.Listener //tcp listeners
		websocketListener *websocket.Conn
		sessions          session.Store
	}
)

func WithTcpListen(tcpListen string) Option {
	return func(opts *Options) {
		opts.tcpListen = tcpListen
	}
}
func WithWebsocketListen(websocketListen string) Option {
	return func(opts *Options) {
		opts.websocketListen = websocketListen
	}
}

func NewServer(opts ...Option) *server {
	options := loadServerOptions(opts...)
	s := &server{}

	s.init(options)
	return s
}
func loadServerOptions(opts ...Option) *Options {
	options := new(Options)
	for _, opt := range opts {
		opt(options)
	}
	if options.tcpListen == "" {
		options.tcpListen = ":1883"
	}
	return options
}

func (s *server) ServeTCP() {
	defer func() {
		err := s.tcpListener.Close()
		if err != nil {
			zap.L().Error("tcpListener close", zap.Error(err))
		}
	}()
	var tempDelay time.Duration

	for {
		accept, err := s.tcpListener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		// 创建一个客户端连接

		c := newClient(s, accept)
		zap.L().Debug("创建一个新的客户端连接")
		// 监听该连接
		err = poolGo.Submit(func() {
			c.listen()
		})
		if err != nil {
			zap.L().Error("资源耗尽", zap.Error(err))
			continue
		}

	}
}

func (s *server) init(opts *Options) {
	s.tcpListen = opts.tcpListen
	s.websocketListen = opts.websocketListen

	ln, err := net.Listen("tcp", s.tcpListen)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	s.tcpListener = ln
	// TODO 创建一个存储客户端会话的容器
	s.sessions = memery.New()

}
func handleGoroutineErr(err error) (isErr bool) {
	if err != nil {
		zap.L().Error("资源耗尽", zap.Error(err))
		return true
	}
	return false
}
