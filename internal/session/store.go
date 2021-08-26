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
	"errors"
)

var (
	NotFoundErr = errors.New("not found")
)

type (
	RangeFn func(session *Session) bool

	Store interface {
		Set(session *Session) error
		Remove(clientId string) error
		Get(clientId string) (*Session, error)
		Range(fn RangeFn) error
		SetSessionExpiry(clientId string, expiryInterval uint32) error
	}
)
