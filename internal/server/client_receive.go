package server

import (
	"github.com/yunqi/lighthouse/internal/code"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/message"
	sub "github.com/yunqi/lighthouse/internal/subscription"
	"github.com/yunqi/lighthouse/internal/xerror"
	"go.uber.org/zap"
	"io"
	"reflect"
	"time"
)

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

func (c *client) handleReceivePublish(publish *packet.Publish) *xerror.Error {
	ctx, span, logger := c.getTraceLog("publish")
	defer span.End()
	logger.Debug("received publish packet", zap.String("packet", publish.String()))

	var (
		dup       bool
		ackPacket packet.Packet
	)
	switch publish.QoS {
	case packet.QoS1:
		ackPacket = publish.CreatePuback()
	case packet.QoS2:
		exist, err := c.unackStore.Set(ctx, publish.PacketId)
		if err != nil {
			return convertError(err)
		}

		if exist {
			dup = true
		}
		ackPacket = publish.CreatePubrec()
	}

	// 第一次收到数据
	if !dup {
		// 分发数据
		topicName := string(publish.TopicName)
		msg := message.FromPublish(publish)

		options := defaultIterateOptions(topicName)
		_ = c.deliverMessage(c.clientId, msg, options)

		if publish.Retain {
			if len(publish.Payload) == 0 {
				c.server.retainedStore.Remove(topicName)
			} else {
				c.server.retainedStore.AddOrReplace(msg)
			}
		}
	}

	if ackPacket != nil {
		// 返回响应
		c.write(ctx, ackPacket)
	}

	return nil
}

func (c *client) handleReceivePingreq(pingreq *packet.Pingreq) {
	ctx, span, logger := c.getTraceLog("ping request")
	defer span.End()
	logger.Debug("received ping request packet", zap.String("packet", pingreq.String()))
	c.write(ctx, pingreq.CreatePingresp())
}

func (c *client) handleReceivePubrel(pubrel *packet.Pubrel) {
	ctx, span, logger := c.getTraceLog("publish release")
	defer span.End()
	logger.Debug("received publish release packet", zap.String("packet", pubrel.String()))
	c.write(ctx, pubrel.CreatePubcomp())
}

func (c *client) handleReceiveSubscribe(subscribe *packet.Subscribe) {
	ctx, span, logger := c.getTraceLog("subscribe")
	defer span.End()

	logger.Debug("received subscribe packet", zap.String("packet", subscribe.String()))

	var subs = make([]*sub.Subscription, 0, len(subscribe.Topics))

	for _, topic := range subscribe.Topics {
		subs = append(subs, &sub.Subscription{
			//ShareName:         topic.Name,
			TopicFilter: topic.Name,
			//ID:                subscribe.PacketId,
			QoS:               topic.QoS,
			NoLocal:           topic.NoLocal,
			RetainAsPublished: topic.RetainAsPublished,
			RetainHandling:    topic.RetainHandling,
		})
	}
	subscribeResult, err := c.subscriptionStore.Subscribe(ctx, c.clientId, subs...)
	if err != nil {
		logger.Error("err", zap.Error(err))
		return

	} else {
		logger.Info("", zap.Any("subscribeResult", subscribeResult))
	}
	c.write(ctx, &packet.Suback{
		Version:  subscribe.Version,
		PacketId: subscribe.PacketId,
		Payload:  make([]code.Code, len(subscribe.Topics)),
	})
}

func (c *client) handleReceiveUnsubscribe(unsubscribe *packet.Unsubscribe) {
	ctx, span, logger := c.getTraceLog("unsubscribe")
	defer span.End()
	logger.Debug("received unsubscribe packet", zap.String("packet", unsubscribe.String()))

	c.write(ctx, &packet.Unsuback{
		Version:  unsubscribe.Version,
		PacketId: unsubscribe.PacketId,
	})
}
