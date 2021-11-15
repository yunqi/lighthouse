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

package memory

import (
	"container/list"
	"github.com/yunqi/lighthouse/internal/persistence/queue"
)

var _ queue.Store = (*Store)(nil)

type Store struct {
	l *list.List
}

func NewMemory() *Store {
	return &Store{l: list.New()}
}

func (m *Store) Add(elem *queue.Element) {
	m.l.PushFront(elem)
}

func (m *Store) Remove(elem *queue.Element) {
	for element := m.l.Front(); element != nil; element = element.Next() {
		if element.Value.(*queue.Element) == elem {
			m.l.Remove(element)
		}
	}
}

func (m *Store) Len() int {
	return m.l.Len()
}

func (m *Store) Front() (elem *queue.Element) {
	return m.l.Front().Value.(*queue.Element)
}

func (m *Store) Replace(elem *queue.Element) (replaced bool, err error) {
	for e := m.l.Front(); e != nil; e = e.Next() {
		if e.Value.(*queue.Element).Id() == elem.Id() {
			e.Value = elem
			return true, nil
		}
	}
	return false, nil
}

func (m *Store) Iterator() queue.Iterator {
	return &iterator{
		l:    m.l,
		next: m.l.Front(),
	}
}

func (m *Store) Reset() {
	m.l = list.New()
}

var _ queue.Iterator = (*iterator)(nil)

type iterator struct {
	l       *list.List
	current *list.Element
	next    *list.Element
}

func (itr *iterator) HasNext() bool {
	return itr.next != nil
}

func (itr *iterator) Next() (*queue.Element, error) {
	itr.current = itr.next
	itr.next = itr.current.Next()
	return itr.current.Value.(*queue.Element), nil
}

func (itr *iterator) Remove() error {
	itr.l.Remove(itr.current)
	return nil
}
