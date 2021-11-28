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

package persistence

import (
	"github.com/yunqi/lighthouse/internal/persistence/queue"
	"github.com/yunqi/lighthouse/internal/persistence/session"
)

var (
	sessionStores = map[string]session.Store{}
	queueStores   = map[string]queue.Store{}
)

func RegisterSessionStore(name string, store session.Store) {
	sessionStores[name] = store
}
func GetSessionStore(name string) (store session.Store, ok bool) {
	s, ok := sessionStores[name]
	return s, ok
}

func RegisterQueueStore(name string, store queue.Store) {
	queueStores[name] = store
}
func GetQueueStore(name string) (store queue.Store, ok bool) {
	s, ok := queueStores[name]
	return s, ok
}
