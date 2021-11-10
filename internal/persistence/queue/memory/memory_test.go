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
	"github.com/golang/mock/gomock"
	"github.com/yunqi/lighthouse/internal/persistence/queue"
	"github.com/yunqi/lighthouse/internal/persistence/store"
	"testing"
	"time"
)

type testNotifier struct {
	dropElem    []*queue.Element
	dropErr     error
	inflightLen int
	msgQueueLen int
}

func (t *testNotifier) Dropped(elem *queue.Element, err error) {
	t.dropElem = append(t.dropElem, elem)
	t.dropErr = err
}

func (t *testNotifier) InflightAdded(delta int) {
	t.inflightLen += delta
	if t.inflightLen < 0 {
		t.inflightLen = 0
	}
}
func (t *testNotifier) MsgQueueAdded(delta int) {
	t.msgQueueLen += delta
	if t.msgQueueLen < 0 {
		t.msgQueueLen = 0
	}
}

func newStore(t *testing.T) store.Store {
	controller := gomock.NewController(t)
	mockStore := store.NewMockStore(controller)
	gomock.InOrder(
		mockStore.EXPECT().Len().Return(1),
		mockStore.EXPECT().Len().Return(2),
		mockStore.EXPECT().Len().Return(3),
	)
	mockStore.EXPECT().Add(gomock.Any()).AnyTimes()

	return mockStore
}

func TestQueue_Add(t *testing.T) {
	opt := &Options{
		MaxQueuedMsg:    10,
		InflightExpiry:  0,
		ClientId:        "",
		DefaultNotifier: &testNotifier{},
	}
	q := New(newStore(t), opt)
	for i := 0; i < 3; i++ {
		q.Add(&queue.Element{
			At:     time.Time{},
			Expiry: time.Time{},
		})
	}

}
