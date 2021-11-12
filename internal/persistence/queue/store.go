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

package queue

import (
	"container/list"
)

//go:generate mockgen -destination ./store_mock.go -package store -source store.go

type Store interface {
	Add(elem *Element)
	Remove(elem *Element)
	Len() int
	Front() (elem *Element)
	Replace(elem *Element) (replaced bool, err error)
	Iterator() Iterator
	Reset()
}

type Iterator interface {
	HasNext() bool
	Next() (*Element, error)
	Remove() error
}

var _ Store = (*MemoryStore)(nil)

type MemoryStore struct {
	l *list.List
}

func NewMemory() *MemoryStore {
	return &MemoryStore{l: list.New()}
}

func (m *MemoryStore) Add(elem *Element) {
	m.l.PushFront(elem)
}

func (m *MemoryStore) Remove(elem *Element) {
	for element := m.l.Front(); element != nil; element = element.Next() {
		if element.Value.(*Element) == elem {
			m.l.Remove(element)
		}
	}
}

func (m *MemoryStore) Len() int {
	return m.l.Len()
}

func (m *MemoryStore) Front() (elem *Element) {
	return m.l.Front().Value.(*Element)
}

func (m *MemoryStore) Replace(elem *Element) (replaced bool, err error) {
	for e := m.l.Front(); e != nil; e = e.Next() {
		if e.Value.(*Element).Id() == elem.Id() {
			e.Value = elem
			return true, nil
		}
	}
	return false, nil
}

func (m *MemoryStore) Iterator() Iterator {
	return &iterator{
		l:    m.l,
		next: m.l.Front(),
	}
}

func (m *MemoryStore) Reset() {
	m.l = list.New()
}

var _ Iterator = (*iterator)(nil)

type iterator struct {
	l       *list.List
	current *list.Element
	next    *list.Element
}

func (itr *iterator) HasNext() bool {
	return itr.next != nil
}

func (itr *iterator) Next() (*Element, error) {
	itr.current = itr.next
	itr.next = itr.current.Next()
	return itr.current.Value.(*Element), nil
}

func (itr *iterator) Remove() error {
	itr.l.Remove(itr.current)
	return nil
}
