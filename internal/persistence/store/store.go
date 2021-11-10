package store

import (
	"container/list"
	"github.com/yunqi/lighthouse/internal/persistence/queue"
)

type Store interface {
	Add(elem *queue.Element)
	Remove(elem *queue.Element)
	Len() int
	Front() (elem *queue.Element)
	Replace(elem *queue.Element) (replaced bool, err error)
	Iterator() Iterator
	Reset()
}

type Iterator interface {
	HasNext() bool
	Next() (*queue.Element, error)
	Remove() error
}

var _ Store = (*Memory)(nil)

type Memory struct {
	l *list.List
}

func NewMemory() *Memory {
	return &Memory{l: list.New()}
}

func (m *Memory) Add(elem *queue.Element) {
	m.l.PushFront(elem)
}

func (m *Memory) Remove(elem *queue.Element) {
	for element := m.l.Front(); element != nil; element = element.Next() {
		if element.Value.(*queue.Element) == elem {
			m.l.Remove(element)
		}
	}
}

func (m *Memory) Len() int {
	return m.l.Len()
}

func (m *Memory) Front() (elem *queue.Element) {
	return m.l.Front().Value.(*queue.Element)
}

func (m *Memory) Replace(elem *queue.Element) (replaced bool, err error) {
	for e := m.l.Front(); e != nil; e = e.Next() {
		if e.Value.(*queue.Element).Id() == elem.Id() {
			e.Value = elem
			return true, nil
		}
	}
	return false, nil
}

func (m *Memory) Iterator() Iterator {
	return &iterator{
		l:       m.l,
		current: m.l.Front(),
	}
}

func (m *Memory) Reset() {
	m.l = list.New()
}

var _ Iterator = (*iterator)(nil)

type iterator struct {
	l       *list.List
	current *list.Element
}

func (itr *iterator) HasNext() bool {
	return itr.current.Next() != nil
}

func (itr *iterator) Next() (*queue.Element, error) {
	element := itr.current
	itr.current = element.Next()
	return element.Value.(*queue.Element), nil
}

func (itr *iterator) Remove() error {
	itr.l.Remove(itr.current)
	return nil
}
