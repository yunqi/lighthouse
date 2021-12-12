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
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/xbitmap"
	"sync"
)

// packetIdLimiter limit the generation of packet id to keep the number of inflight messages
// always less or equal than receive maximum setting of the client.
type packetIdLimiter struct {
	cond              *sync.Cond
	used              uint16
	limit             uint16
	exit              bool
	lockedPacketIdMap *xbitmap.Bitmap // packet id in-use
	freePacketId      packet.PacketId // next available id
}

func newPacketIDLimiter(limit uint16) *packetIdLimiter {
	return &packetIdLimiter{
		cond:              sync.NewCond(&sync.Mutex{}),
		used:              0,
		limit:             limit,
		exit:              false,
		freePacketId:      packet.MinPacketID,
		lockedPacketIdMap: xbitmap.New(packet.MaxPacketID),
	}
}

func (p *packetIdLimiter) close() {
	p.lock()
	p.exit = true
	p.unlock()
	p.cond.Signal()
}

func (p *packetIdLimiter) lock() {
	p.cond.L.Lock()
}

func (p *packetIdLimiter) unlock() {
	p.cond.L.Unlock()
}

// markUsedLocked marks the given id as used.
func (p *packetIdLimiter) markUsedLocked(packetId packet.PacketId) {
	p.used++
	p.lockedPacketIdMap.Set(packetId, 1)
}

func (p *packetIdLimiter) unlockAndSignal() {
	p.cond.L.Unlock()
	p.cond.Signal()
}

func (p *packetIdLimiter) releaseLocked(packetId packet.PacketId) {
	if p.lockedPacketIdMap.Get(packetId) == 1 {
		p.lockedPacketIdMap.Set(packetId, 0)
		p.used--
	}
}

// release marks the given id list as unused
func (p *packetIdLimiter) release(id packet.PacketId) {
	p.lock()
	p.releaseLocked(id)
	p.unlock()
	p.cond.Signal()
}

func (p *packetIdLimiter) batchRelease(packetIds []packet.PacketId) {
	p.lock()
	for _, packetId := range packetIds {
		p.releaseLocked(packetId)
	}
	p.unlock()
	p.cond.Signal()

}

// pollPacketIds returns at most max number of unused packetID and marks them as used for a client.
// If there is no available id, the call will be blocked until at least one packet id is available or the limiter has been closed.
// return 0 means the limiter is closed.
// the return number = min(max, i.used).
func (p *packetIdLimiter) pollPacketIds(max uint16) (id []packet.PacketId) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	for p.used >= p.limit && !p.exit {
		p.cond.Wait()
	}
	if p.exit {
		return nil
	}
	n := max
	if remain := p.limit - p.used; remain < max {
		n = remain
	}
	for j := uint16(0); j < n; j++ {
		for p.lockedPacketIdMap.Get(p.freePacketId) == 1 {
			if p.freePacketId == packet.MaxPacketID {
				p.freePacketId = packet.MinPacketID
			} else {
				p.freePacketId++
			}
		}
		id = append(id, p.freePacketId)
		p.used++
		p.lockedPacketIdMap.Set(p.freePacketId, 1)
		if p.freePacketId == packet.MaxPacketID {
			p.freePacketId = packet.MinPacketID
		} else {
			p.freePacketId++
		}
	}
	return id
}
