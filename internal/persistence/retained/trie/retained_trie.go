package trie

import (
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"github.com/yunqi/lighthouse/internal/persistence/retained"
	"strings"
)

const delimiter = "/"
const poundSign = "#"
const plus = "+"

// children
type children map[string]*topicNode

// topicNode
type topicNode struct {
	children  children
	msg       *message.Message
	parent    *topicNode // pointer of parent node
	topicName string
}

// newTopicTrie create a new trie tree
func newTopicTrie() *topicNode {
	return newNode()
}

// newNode create a new trie node
func newNode() *topicNode {
	return &topicNode{
		children: children{},
	}
}

// newChild create a child node of t
func (t *topicNode) newChild() *topicNode {
	return &topicNode{
		children: children{},
		parent:   t,
	}
}

// find walk through the tire and return the node that represent the topicName
// return nil if not found
func (t *topicNode) find(topicName string) *topicNode {
	topicSlice := splitTopicName(topicName)
	var pNode = t
	for _, lv := range topicSlice {
		if _, ok := pNode.children[lv]; ok {
			pNode = pNode.children[lv]
		} else {
			return nil
		}
	}
	if pNode.msg != nil {
		return pNode
	}
	return nil
}

// matchTopic walk through the tire and call the fn callback for each message witch match the topic filter.
func (t *topicNode) matchTopic(topicSlice []string, fn retained.IterateFn) {
	endFlag := len(topicSlice) == 1
	switch topicSlice[0] {
	case poundSign:
		t.preOrderTraverse(fn)
	case plus:
		// 当前层的所有
		for _, v := range t.children {
			if endFlag {
				if v.msg != nil {
					fn(v.msg)
				}
			} else {
				v.matchTopic(topicSlice[1:], fn)
			}
		}
	default:
		if n := t.children[topicSlice[0]]; n != nil {
			if endFlag {
				if n.msg != nil {
					fn(n.msg)
				}
			} else {
				n.matchTopic(topicSlice[1:], fn)
			}
		}
	}
}

func (t *topicNode) getMatchedMessages(topicFilter string) []*message.Message {
	topicLv := splitTopicName(topicFilter)
	var rs []*message.Message
	t.matchTopic(topicLv, func(message *message.Message) bool {
		rs = append(rs, message.Copy())
		return true
	})
	return rs
}

func isSystemTopic(topicName string) bool {
	return len(topicName) >= 1 && topicName[0] == '$'
}

// addRetainMsg add a retain message
func (t *topicNode) addRetainMsg(topicName string, message *message.Message) {
	topicSlice := splitTopicName(topicName)
	var pNode = t
	for _, lv := range topicSlice {
		if _, ok := pNode.children[lv]; !ok {
			pNode.children[lv] = pNode.newChild()
		}
		pNode = pNode.children[lv]
	}
	pNode.msg = message
	pNode.topicName = topicName
}

func (t *topicNode) remove(topicName string) {
	topicSlice := splitTopicName(topicName)
	l := len(topicSlice)
	var pNode = t
	for _, lv := range topicSlice {
		if _, ok := pNode.children[lv]; ok {
			pNode = pNode.children[lv]
		} else {
			return
		}
	}
	pNode.msg = nil
	if len(pNode.children) == 0 {
		delete(pNode.parent.children, topicSlice[l-1])
	}
}

func (t *topicNode) preOrderTraverse(fn retained.IterateFn) bool {

	if t.msg != nil {
		if !fn(t.msg) {
			return false
		}
	}
	for _, c := range t.children {
		if !c.preOrderTraverse(fn) {
			return false
		}
	}
	return true
}

func splitTopicName(topicName string) []string {
	return strings.Split(topicName, delimiter)
}
