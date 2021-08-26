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
	"github.com/chenquan/lighthouse/internal/xlog"
	"github.com/gorilla/websocket"
	"net"
	"os"
	"time"
)

var log = xlog.Logger()

type (
	Server interface {
		Stop(ctx context.Context) error
		Run() error
	}
	Option func(server *server)

	server struct {
		tcpListen         string
		websocketListen   string
		tcpListener       net.Listener //tcp listeners
		websocketListener *websocket.Conn
	}
)

func WithTcpListen(tcpListen string) Option {
	return func(server *server) {
		server.tcpListen = tcpListen
	}
}
func WithWebsocketListen(websocketListen string) Option {
	return func(server *server) {
		server.websocketListen = websocketListen
	}
}

func NewServer(opts ...Option) *server {
	s := &server{}
	for _, opt := range opts {
		opt(s)
	}
	if s.tcpListen == "" {
		s.tcpListen = ":1883"
	}
	s.init()
	return s
}

func (s *server) serveTCP() {
	defer func() {
		_ = s.tcpListener.Close()
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
		_ = accept
		c := newClient(s, accept)
		log.Info("创建一个新的客户端连接")
		// 监听该连接
		go c.listen()
	}
}

func (s *server) init() {

	ln, err := net.Listen("tcp", s.tcpListen)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	s.tcpListener = ln
}
