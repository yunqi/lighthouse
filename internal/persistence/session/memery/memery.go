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

package memery

import (
	"context"
	"github.com/yunqi/lighthouse/internal/persistence/session"
	"sync"
)

type (
	Store struct {
		m *sync.Map
	}
)

func New() *Store {
	return &Store{
		m: &sync.Map{},
	}
}

func (s *Store) Set(ctx context.Context, session *session.Session) error {
	s.m.Store(session.ClientId, session)
	return nil
}

func (s *Store) Remove(ctx context.Context, clientId string) error {
	s.m.Delete(clientId)
	return nil
}

func (s *Store) Get(ctx context.Context, clientId string) (*session.Session, error) {
	if val, ok := s.m.Load(clientId); ok {
		return val.(*session.Session), nil
	}
	return nil, session.NotFoundErr
}

func (s *Store) Range(ctx context.Context, fn session.RangeFn) error {
	s.m.Range(func(_, value interface{}) bool {
		return fn(value.(*session.Session))
	})
	return nil
}

func (s *Store) SetSessionExpiry(ctx context.Context, clientId string, expiryInterval uint32) error {
	if value, ok := s.m.Load(clientId); ok {
		value.(*session.Session).ExpiryInterval = expiryInterval
	}
	return nil
}
