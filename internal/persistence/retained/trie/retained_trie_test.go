package trie

import (
	"github.com/stretchr/testify/assert"
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"testing"
)

func Test_topicNode_addRetainMsg(t *testing.T) {

	node := newNode()
	m1 := &message.Message{}
	m2 := &message.Message{}
	m3 := &message.Message{}
	m4 := &message.Message{}
	node.addRetainMsg("test", m1)
	node.addRetainMsg("test/1", m2)
	node.addRetainMsg("test/3", m3)
	node.addRetainMsg("test/3/1", m4)
	assert.EqualValues(t, m1, node.children["test"].msg)
	assert.EqualValues(t, m2, node.children["test"].children["1"].msg)
	assert.EqualValues(t, m3, node.children["test"].children["3"].msg)
	assert.EqualValues(t, m4, node.children["test"].children["3"].children["1"].msg)

}

func Test_topicNode_find(t *testing.T) {
	node := newNode()
	m1 := &message.Message{}
	m2 := &message.Message{}
	m3 := &message.Message{}
	m4 := &message.Message{}
	node.addRetainMsg("test", m1)
	node.addRetainMsg("test/1", m2)
	node.addRetainMsg("test/3", m3)
	node.addRetainMsg("test/3/1", m4)

	t.Run("1", func(t *testing.T) {
		n1 := node.find("test")
		assert.EqualValues(t, m1, n1.msg)
	})

	t.Run("2", func(t *testing.T) {
		n2 := node.find("test/1")
		assert.EqualValues(t, m2, n2.msg)
	})

	t.Run("3", func(t *testing.T) {
		n3 := node.find("test/3")
		assert.EqualValues(t, m3, n3.msg)
	})

	t.Run("4", func(t *testing.T) {
		n4 := node.find("test/3/1")
		assert.EqualValues(t, m4, n4.msg)
	})

	t.Run("5", func(t *testing.T) {
		n := node.find("")
		assert.EqualValues(t, (*topicNode)(nil), n)
	})

	t.Run("6", func(t *testing.T) {
		n := node.find("tt")
		assert.EqualValues(t, (*topicNode)(nil), n)
	})
}

func Test_topicNode_remove(t *testing.T) {

	node := newNode()
	m1 := &message.Message{}
	m2 := &message.Message{}
	m3 := &message.Message{}
	m4 := &message.Message{}
	node.addRetainMsg("test", m1)
	node.addRetainMsg("test/1", m2)
	node.addRetainMsg("test/3", m3)
	node.addRetainMsg("test/3/1", m4)

	t.Run("1", func(t *testing.T) {
		node.remove("test")
		n := node.find("test")
		assert.EqualValues(t, (*topicNode)(nil), n)

		n = node.find("test/1")
		assert.NotNil(t, n)

		n = node.find("test/3")
		assert.NotNil(t, n)

		n = node.find("test/3/1")
		assert.NotNil(t, n)

	})

	t.Run("2", func(t *testing.T) {
		node.remove("test/1")

		n := node.find("test/1")
		assert.Nil(t, n)

		n = node.find("test/3")
		assert.NotNil(t, n)

		n = node.find("test/3/1")
		assert.NotNil(t, n)
	})

	t.Run("3", func(t *testing.T) {
		node.remove("test/3")

		n := node.find("test/3")
		assert.Nil(t, n)

		n = node.find("test/3/1")
		assert.NotNil(t, n)
	})

	t.Run("4", func(t *testing.T) {
		node.remove("test/3/1")

		n := node.find("test/3/1")
		assert.Nil(t, n)
	})

}

func Test_topicNode_getMatchedMessages(t *testing.T) {

	node := newNode()
	m1 := &message.Message{Topic: "1"}
	m2 := &message.Message{Topic: "2"}
	m3 := &message.Message{Topic: "3"}
	m4 := &message.Message{Topic: "4"}
	node.addRetainMsg("test", m1)
	node.addRetainMsg("test/1", m2)
	node.addRetainMsg("test/3", m3)
	node.addRetainMsg("test/3/1", m4)

	t.Run("1", func(t *testing.T) {
		messages := node.getMatchedMessages("test/1")

		assert.EqualValues(t, m2, messages[0])
	})

	t.Run("2", func(t *testing.T) {
		messages := node.getMatchedMessages("test/3/1")
		assert.EqualValues(t, m4, messages[0])
	})
}
