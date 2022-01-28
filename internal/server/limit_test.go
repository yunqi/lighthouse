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
	"github.com/stretchr/testify/assert"
	"github.com/yunqi/lighthouse/internal/packet"
	"testing"
	"time"
)

func Test_packetIDLimiter(t *testing.T) {
	a := assert.New(t)
	p := newPacketIDLimiter(10)
	ids := p.pollPacketIds(20)
	a.Len(ids, 10)
	a.Equal([]packet.Id{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, ids)

	p.batchRelease([]packet.Id{7, 8, 9})

	ids = p.pollPacketIds(4)
	a.Len(ids, 3)
	a.Equal([]packet.Id{11, 12, 13}, ids)

	c := make(chan struct{})
	go func() {
		p.pollPacketIds(1)
		c <- struct{}{}
	}()
	select {
	case <-c:
		t.Fatal("pollPacketIds should be blocked")
	case <-time.After(1 * time.Second):
	}
	p.close()
	a.Nil(p.pollPacketIds(10))
}

func Test_packetIDLimiterMax(t *testing.T) {
	a := assert.New(t)
	p := newPacketIDLimiter(packet.MaxPacketID)
	ids := p.pollPacketIds(packet.MaxPacketID)
	a.Len(ids, int(packet.MaxPacketID))
	p.batchRelease([]packet.Id{1, 2, 3, packet.MaxPacketID})
	a.Equal([]packet.Id{1, 2, 3}, p.pollPacketIds(3))
	a.Equal([]packet.Id{packet.MaxPacketID}, p.pollPacketIds(3))

}
