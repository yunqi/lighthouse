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
	"errors"
	"fmt"
	"github.com/chenquan/go-pkg/xio"
	"github.com/yunqi/lighthouse/internal/code"
	"github.com/yunqi/lighthouse/internal/goroutine"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"github.com/yunqi/lighthouse/internal/persistence/queue"
	"github.com/yunqi/lighthouse/internal/session"
	"github.com/yunqi/lighthouse/internal/xerror"
	"github.com/yunqi/lighthouse/internal/xlog"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Connecting Status = iota
	Connected
)

type (
	Status byte
	// Client represent a mqtt client.
	Client interface {
		// ClientOption return a reference of ClientOption. Do not edit.
		// This is mainly used in hooks.
		ClientOption() *ClientOption
		// Session return a reference of session information of the client. Do not edit.
		// Session info will be available after the client has passed OnSessionCreated or OnSessionResume.
		Session() *session.Session
		// Version return the protocol version of the used client.
		Version() packet.Version
		// ConnectedAt returns the connected time
		ConnectedAt() time.Time
		// Connection returns the raw net.Conn
		Connection() net.Conn
		// Close closes the client connection.
		Close() error
		// Disconnect sends a disconnect packet to client, it is use to close v5 client.
		Disconnect(disconnect *packet.Disconnect)
		Deliverer
	}

	// ClientOption is the options which controls how the server interacts with the client.
	// It will be set after the client has connected.
	ClientOption struct {
		// ClientId is the client id for the client.
		ClientId string
		// Username is the username for the client.
		Username string
		// KeepAlive is the keep alive time in seconds for the client.
		// The server will close the client if no there is no packet has been received for 1.5 times the KeepAlive time.
		KeepAlive uint16
		// SessionExpiry is the session expiry interval in seconds.
		// If the client version is v5, this value will be set into CONNACK Session Expiry Interval property.
		// See: https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901082
		SessionExpiry uint32
		// MaxInflight limits the number of QoS 1 and QoS 2 publications that the client is willing to process concurrently.
		// For v3 client, it is default to config.MQTT.MaxInflight.
		// For v5 client, it is the minimum of config.MQTT.MaxInflight and Receive Maximum property in CONNECT packet.
		MaxInflight uint16
		// ReceiveMax limits the number of QoS 1 and QoS 2 publications that the server is willing to process concurrently for the Client.
		// If the client version is v5, this value will be set into Receive Maximum property in CONNACK packet.
		// See: https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901083
		ReceiveMax uint16
		// ClientMaxPacketSize is the maximum packet size that the client is willing to accept.
		// The server will drop the packet if it exceeds ClientMaxPacketSize.
		// See: https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901050
		ClientMaxPacketSize uint32
		// ServerMaxPacketSize is the maximum packet size that the server is willing to accept from the client.
		// See: https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901086
		ServerMaxPacketSize uint32
		// ClientTopicAliasMax is highest value that the client will accept as a Topic Alias sent by the server.
		// See: https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901051
		ClientTopicAliasMax uint16
		// ServerTopicAliasMax is highest value that the server will accept as a Topic Alias sent by the client.
		// See: https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901088
		ServerTopicAliasMax uint16
		// RequestProblemInfo is the value to indicate whether the Reason String or User Properties should be sent in the case of failures.
		// See: https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901053
		RequestProblemInfo bool
	}
	client struct {
		clientId      string
		connectedAt   int64
		clientConn    net.Conn
		bufReader     io.Reader
		bufWriter     io.Writer
		packetReader  *packet.Reader
		packetWriter  *packet.Writer
		status        Status
		server        *server
		in            chan packet.Packet
		out           chan packet.Packet
		session       *session.Session
		cleanWillFlag bool // whether to remove will Msg
		version       packet.Version
		opt           *ClientOption //set up before OnConnect()
		disconnect    *packet.Disconnect
		closed        chan struct{}
		connected     chan struct{}
		wg            sync.WaitGroup
		queueStore    queue.Queue
		limit         *packetIdLimiter
		//ctx           context.Context
		log        *xlog.Log
		remoteAddr net.Addr
	}
)

func (c *client) ClientOption() *ClientOption {
	panic("implement me")
}

func (c *client) Deliver(message message.Message) error {
	panic("implement me")
}

func (c *client) ClientOptions() *ClientOption {
	return c.opt
}

func (c *client) Session() *session.Session {
	return c.session
}

func (c *client) Version() packet.Version {
	return c.version
}

func (c *client) ConnectedAt() time.Time {
	return time.Unix(atomic.LoadInt64(&c.connectedAt), 0)
}

func (c *client) Connection() net.Conn {
	return c.clientConn
}

func (c *client) Close() error {
	defer func() {
		c.log.Debug("关闭客户端")
	}()
	if c.clientConn != nil {
		return c.clientConn.Close()
	}

	return nil
}
func (c *client) Status() Status {
	return c.status
}
func (c *client) IsConnected() bool {
	return c.status == Connected
}
func (c *client) IsConnecting() bool {
	return c.status == Connecting
}
func (c *client) Disconnect(disconnect *packet.Disconnect) {
}

func newClient(server *server, conn net.Conn) *client {
	reader := xio.GetBufferReaderSize(conn, 2048)
	writer := xio.GetBufferWriterSize(conn, 2048)
	c := &client{
		server:       server,
		clientConn:   conn,
		bufReader:    reader,
		bufWriter:    writer,
		packetReader: packet.NewReader(reader),
		packetWriter: packet.NewWriter(writer),
		connectedAt:  time.Now().UnixMilli(),
		in:           make(chan packet.Packet, 8),
		out:          make(chan packet.Packet, 8),
		closed:       make(chan struct{}),
		connected:    make(chan struct{}),
		log:          xlog.LoggerModule("client"),
		remoteAddr:   conn.RemoteAddr(),
	}
	return c
}

func (c *client) listen() {
	ctx, span := c.server.tracer.Start(context.Background(), "listen")
	logger := c.log.WithContext(ctx)
	logger.Debug("create a new client connection", zap.Any("IP", c.remoteAddr.String()))

	c.wg.Add(1)
	goroutine.Go(func() {
		defer c.wg.Done()
		c.writeConn()
	})

	// 认证
	if !c.auth(ctx) {
		span.End()
		return
	}
	span.End()

	readWg := sync.WaitGroup{}
	readWg.Add(1)
	goroutine.Go(func() {
		//read conn
		defer readWg.Done()
		c.readConn()
	})

	// 拉取消息
	c.wg.Add(1)
	goroutine.Go(func() {
		c.pollMessageHandler()
		c.wg.Done()
	})

	c.wg.Add(1)
	goroutine.Go(func() {
		defer c.wg.Done()
		c.handleConn()
	})

	readWg.Wait()

	c.wg.Wait()

}

func (c *client) auth(ctx context.Context) bool {
	ctx, span := c.server.tracer.Start(ctx, "auth")
	defer span.End()
	logger := c.log.WithContext(ctx)

	logger.Debug("开始认证")

	var p packet.Packet
	if c.IsConnected() {
		if keepAlive := c.opt.KeepAlive; keepAlive != 0 { //KeepAlive
			_ = c.clientConn.SetReadDeadline(time.Now().Add(time.Duration(keepAlive/2+keepAlive) * time.Second))
		}
	}

	p, err := c.packetReader.Read()
	if err != nil {
		if err != io.EOF && p != nil {
			logger.Error("read error", zap.String("packet_type", reflect.TypeOf(p).String()))
		}
		select {
		case <-c.closed:
			logger.Debug("客户端退出，关闭连接")
		default:
			logger.Debug("连接超时，自动关闭")
		}
		return false
	}

	if connect, ok := p.(*packet.Connect); ok {
		if !c.connectAuthentication(ctx, connect) {
			logger.Debug("authentication failed", zap.String("IP", c.remoteAddr.String()))
			return false
		}
		return true
	}

	logger.Debug("invalid package", zap.String("package", p.String()))
	_ = c.Close()
	logger.Debug("close connection", zap.String("IP", c.remoteAddr.String()))
	return false
}

func (c *client) readConn() {
	defer func() {
		// 关闭 in 通道
		_ = c.Close()
		close(c.in)
	}()
	go func() {
		select {
		case <-c.closed:
			// 立即关闭
			_ = c.clientConn.SetReadDeadline(time.Now())
			return
		}
	}()
	for {
		var p packet.Packet
		if c.IsConnected() {
			if keepAlive := c.opt.KeepAlive; keepAlive != 0 { //KeepAlive
				_ = c.clientConn.SetReadDeadline(time.Now().Add(time.Duration(keepAlive/2+keepAlive) * time.Second))
			}
		}
		p, err := c.packetReader.Read()
		if err != nil {
			if err != io.EOF && p != nil {
				c.log.Error("read error", zap.String("packet_type", reflect.TypeOf(p).String()))
			}
			select {
			case <-c.closed:
				c.log.Debug("客户端退出，关闭连接")
			default:
				c.log.Debug("连接超时，自动关闭")
			}
			return
		}
		//if connect, ok := p.(*packet.Connect); ok {
		//	c.log.Debug("接收认证信息", zap.String("ClientId", string(connect.ClientId)))
		//} else {
		//	//c.log.Debug("Rec data", zap.String("packet", p.String()))
		//}
		c.in <- p

		// 等待连接认证完成
		//c.waitConnection()

	}
}

func (c *client) writeConn() {

	defer func() {
	}()
	for p := range c.out {
		//c.log.Debug("Ret data", zap.String("packet", p.String()))
		err := c.packetWriter.WritePacketAndFlush(p)
		if err != nil {
			return
		}
	}
	c.log.Debug("写入操作退出")

}
func (c *client) write(ctx context.Context, packet packet.Packet) {
	c.log.WithContext(ctx).Debug("write packet", zap.String("packet", packet.String()))
	c.out <- packet
}

//func (c *client) waitConnection() {
//	<-c.connected
//}

func (c *client) connectionDone() {
	close(c.connected)
}

func (c *client) connection() (ok bool) {
	defer func() {
		c.connectionDone()
	}()
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case p := <-c.in:
			//c.log.Debug("从in通道中读出数据", zap.Any("packet", p))
			if p == nil {
				return
			}

			switch conn := p.(type) {
			case *packet.Connect:
				if conn == nil {
					//err := xerror.ErrProtocol
					break
				}
				return c.connectAuthentication(context.Background(), conn)
			default:
			}
		case <-timeout.C:
			return
		}

	}

}

// TODO 验证客户端连接
// connectAuthentication 连接验证
func (c *client) connectAuthentication(ctx context.Context, conn *packet.Connect) (ok bool) {
	logger := c.log.WithContext(ctx)

	// 根据报文进行认证
	var connack *packet.Connack
	connack = conn.NewConnackPacket(code.Success, true)
	c.clientId = string(conn.ClientId)
	logger.Debug("认证成功", zap.String("clientId", c.clientId))

	c.status = Connected
	var msg *message.Message
	if conn.WillFlag {
		msg = &message.Message{
			Dup:                    false,
			QoS:                    conn.WillQoS,
			Retained:               conn.WillRetain,
			Topic:                  string(conn.WillTopic),
			Payload:                conn.WillMessage,
			PacketId:               0,
			ContentType:            "",
			CorrelationData:        nil,
			MessageExpiry:          0,
			PayloadFormat:          packet.PayloadFormatBytes,
			ResponseTopic:          "",
			SubscriptionIdentifier: nil,
		}
	}
	c.session = &session.Session{
		ClientId:          c.clientId,
		Will:              msg,
		WillDelayInterval: 0,
		ConnectedAt:       time.Now(),
		ExpiryInterval:    0,
	}
	// client session
	_ = c.server.sessions.Set(c.session)
	c.version = conn.Version
	c.opt = &ClientOption{
		ClientId:  c.clientId,
		Username:  string(conn.Username),
		KeepAlive: conn.KeepAlive,
		//SessionExpiry:       conn,
		MaxInflight:         0,
		ReceiveMax:          0,
		ClientMaxPacketSize: 0,
		ServerMaxPacketSize: 0,
		ClientTopicAliasMax: 0,
		ServerTopicAliasMax: 0,
		RequestProblemInfo:  false,
	}
	c.newPacketIdLimiter(c.opt.MaxInflight)
	c.write(ctx, connack)
	return true
}

func (c *client) handleConn() {
	defer func() {
		close(c.closed)
		close(c.out)
	}()
	var err *xerror.Error
	// in 通道关闭时，自动退出
	for p := range c.in {
		switch packetData := p.(type) {
		case *packet.Publish:
			err = c.handlePublish(packetData)
		case *packet.Pingreq:
			c.handlePingreq(packetData)
		case *packet.Pubrel:
			c.handlePubrel(packetData)
		case *packet.Subscribe:
			c.handleSubscribe(packetData)
		case *packet.Unsubscribe:
			c.handleUnsubscribe(packetData)
		case *packet.Disconnect:
			break
		default:
		}
		if err != nil {
			break
		}
	}
}
func (c *client) getTraceLog(spanName string) (context.Context, trace.Span, *zap.Logger) {
	ctx, span := c.server.tracer.Start(context.Background(), spanName)
	logger := c.log.WithContext(ctx)
	return ctx, span, logger
}
func (c *client) handlePublish(publish *packet.Publish) *xerror.Error {
	ctx, span, logger := c.getTraceLog("publish")
	defer span.End()
	logger.Debug("received publish packet", zap.String("packet", publish.String()))
	//msg := message.FromPublish(publish)
	var ackPacket packet.Packet
	switch publish.QoS {
	case packet.QoS1:
		ackPacket = publish.CreatePuback()
	case packet.QoS2:
		ackPacket = publish.CreatePubrec()
	}

	if ackPacket != nil {
		// 返回响应
		//c.log.Debug("返回响应", zap.Any("packet", ackPacket))
		c.write(ctx, ackPacket)
	}

	return nil
}

func (c *client) handlePingreq(pingreq *packet.Pingreq) {
	ctx, span, logger := c.getTraceLog("ping request")
	defer span.End()
	logger.Debug("received ping request packet", zap.String("packet", pingreq.String()))
	c.write(ctx, pingreq.CreatePingresp())
}

func (c *client) handlePubrel(pubrel *packet.Pubrel) {
	ctx, span, logger := c.getTraceLog("publish release")
	defer span.End()

	logger.Debug("received publish release packet", zap.String("packet", pubrel.String()))
	c.write(ctx, pubrel.CreatePubcomp())
}

func (c *client) handleSubscribe(subscribe *packet.Subscribe) {
	ctx, span, logger := c.getTraceLog("subscribe")
	defer span.End()

	logger.Debug("received subscribe packet", zap.String("packet", subscribe.String()))
	c.write(ctx, &packet.Suback{
		Version:  subscribe.Version,
		PacketId: subscribe.PacketId,
		Payload:  make([]code.Code, len(subscribe.Topics)),
	})
}

func (c *client) handleUnsubscribe(unsubscribe *packet.Unsubscribe) {
	ctx, span, logger := c.getTraceLog("unsubscribe")
	defer span.End()
	logger.Debug("received unsubscribe packet", zap.String("packet", unsubscribe.String()))

	c.write(ctx, &packet.Unsuback{
		Version:  unsubscribe.Version,
		PacketId: unsubscribe.PacketId,
	})
}

func (c *client) pollMessageHandler() {
	var err error
	defer func() {
		if re := recover(); re != nil {
			err = errors.New(fmt.Sprint(re))
		}
		//c.setError(err)
	}()
	cont := true
	for cont {

		cont, err = c.pollInFlights()

		if err != nil {
			return
		}
	}
	var ids []packet.PacketId
	for {
		max := uint16(100)
		if c.opt.MaxInflight < max {
			max = c.opt.MaxInflight
		}
		ids = c.limit.pollPacketIds(max)
		if ids == nil {
			return
		}
		ids, err = c.pollNewMessages(ids)
		if err != nil {
			return
		}
		c.limit.batchRelease(ids)
	}
}
func (c *client) pollNewMessages(ids []packet.PacketId) (unused []packet.PacketId, err error) {
	var elems []*queue.Element
	elems, err = c.queueStore.Read(ids)
	if err != nil {
		return nil, err
	}
	for _, v := range elems {
		switch m := v.Message.(type) {
		case *queue.Publish:
			if m.QoS != packet.QoS0 {
				ids = ids[1:]
			}
			c.write(context.Background(), message.ToPublish(m.Message, c.version))
		case *queue.Pubrel:
		}
	}
	return ids, err
}
func (c *client) pollInFlights() (bool, error) {
	var elems []*queue.Element
	elems, err := c.queueStore.ReadInflight(uint(c.opt.MaxInflight))
	if err != nil || len(elems) == 0 {
		return false, err
	}
	c.limit.lock()
	defer c.limit.unlock()
	for _, elem := range elems {
		id := elem.Message.Id()
		switch m := elem.Message.(type) {
		case *queue.Publish:
			m.Dup = true

			m.SubscriptionIdentifier = nil
			c.limit.markUsedLocked(id)
			c.write(context.Background(), message.ToPublish(m.Message, c.version))
		case *queue.Pubrel:
			c.write(context.Background(), &packet.Pubrel{PacketId: id})
		}
	}
	return false, nil
}

func (c *client) newPacketIdLimiter(limit uint16) {
	c.limit = newPacketIDLimiter(limit)
}
