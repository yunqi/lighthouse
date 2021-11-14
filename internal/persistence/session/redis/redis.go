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

package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	session2 "github.com/yunqi/lighthouse/internal/persistence/session"
	"golang.org/x/sync/singleflight"
	"sync"
	"time"
)

const sessionPrefix = "lighthouse:session:"

type Redis struct {
	client       redis.Client
	singleflight singleflight.Group
	pool         sync.Pool
}

func (r *Redis) Set(ctx context.Context, session *session2.Session) error {
	jsonData, _ := json.Marshal(session)
	//r.client.HSet(ctx,)
	statusCmd := r.client.Set(ctx, session.ClientId, jsonData, time.Second*time.Duration(session.ExpiryInterval))
	return statusCmd.Err()
}
func getSessionKey(client string) string {
	return sessionPrefix + client
}
func (r *Redis) Remove(ctx context.Context, clientId string) error {
	return r.client.Del(ctx, clientId).Err()
}

func (r *Redis) Get(ctx context.Context, clientId string) (*session2.Session, error) {
	ValStr, err, _ := r.singleflight.Do(clientId, func() (interface{}, error) {
		stringCmd := r.client.Get(ctx, clientId)
		result, err := stringCmd.Bytes()
		return result, err
	})
	if err != nil {
		return nil, err
	}
	s := new(session2.Session)
	err = json.Unmarshal(ValStr.([]byte), s)
	if err != nil {
		return nil, err
	}
	return s, err

}

func (r *Redis) Range(ctx context.Context, fn session2.RangeFn) error {
	panic("implement me")
}

func (r *Redis) SetSessionExpiry(ctx context.Context, clientId string, expiryInterval uint32) error {
	panic("implement me")
}
