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
	"github.com/yunqi/lighthouse/internal/persistence/session"
	"github.com/yunqi/lighthouse/internal/persistence/subscription"
)

const (
	Memory = "memory"
	Redis  = "redis"
)

var (
	sessionStores      = map[string]session.NewStore{}
	subscriptionStores = map[string]subscription.NewStore{}
)

func RegisterSessionStore(name string, store session.NewStore) {
	sessionStores[name] = store
}
func GetSessionStore(name string) (store session.NewStore, ok bool) {
	s, ok := sessionStores[name]
	return s, ok
}

func RegisterSubscriptionStore(name string, store subscription.NewStore) {
	subscriptionStores[name] = store
}

func GetSubscriptionStore(name string) (store subscription.NewStore, ok bool) {
	s, ok := subscriptionStores[name]
	return s, ok
}
