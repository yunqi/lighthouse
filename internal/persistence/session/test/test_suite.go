package test

import (
	"context"
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"github.com/yunqi/lighthouse/internal/persistence/session"
	session2 "github.com/yunqi/lighthouse/internal/session"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSuite(t *testing.T, store session.Store) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	var tt = []*session2.Session{
		{
			ClientId: "client",
			Will: &message.Message{
				Topic:   "topicA",
				Payload: []byte("abc"),
			},
			WillDelayInterval: 1,
			ConnectedAt:       time.Unix(1, 0),
			ExpiryInterval:    2,
		}, {
			ClientId:          "client2",
			Will:              nil,
			WillDelayInterval: 0,
			ConnectedAt:       time.Unix(2, 0),
			ExpiryInterval:    0,
		},
	}
	for _, v := range tt {
		a.Nil(store.Set(context.Background(), v))
	}
	for _, v := range tt {
		sess, err := store.Get(context.Background(), v.ClientId)
		a.Nil(err)
		a.EqualValues(v, sess)
	}
	var sess []*session2.Session
	err := store.Iterate(context.Background(), func(session *session2.Session) bool {
		sess = append(sess, session)
		return true
	})
	a.Nil(err)
	a.ElementsMatch(sess, tt)
}
