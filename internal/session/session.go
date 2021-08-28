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

package session

import (
	"github.com/yunqi/lighthouse/internal/store"
	"time"
)

type (
	Session struct {
		// ClientId represents the client id.
		ClientId string
		// Will is the will message of the client, can be nil if there is no will message.
		Will *store.Message
		// WillDelayInterval represents the Will Delay Interval in seconds
		WillDelayInterval uint32
		// ConnectedAt is the session create time.
		ConnectedAt time.Time
		// ExpiryInterval represents the Session Expiry Interval in seconds
		ExpiryInterval uint32
	}
)

// IsExpired return whether the session is expired
func (s *Session) IsExpired(now time.Time) bool {
	return s.ConnectedAt.Add(time.Duration(s.ExpiryInterval) * time.Second).Before(now)
}
