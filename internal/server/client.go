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
	"fmt"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/session"
	"github.com/yunqi/lighthouse/internal/xio"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Connecting = iota
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
	}
)

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
	panic("implement me")
}

func newClient(server *server, conn net.Conn) *client {
	reader := xio.NewBufReaderSize(conn, 2048)
	writer := xio.NewBufWriterSize(conn, 2048)
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
	}
	return c
}
func (c *client) listen() {
	fmt.Println("监听该连接")
	waitGroup := sync.WaitGroup{}
	//read conn
	waitGroup.Add(1)
	c.readConn()
}

func (c *client) readConn() {
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
				//zaplog.Error("read error", zap.String("packet_type", reflect.TypeOf(packet).String()))
			}
			return
		}
		fmt.Println("收到数据")
		fmt.Println(p.String())
		c.in <- p

	}
}
