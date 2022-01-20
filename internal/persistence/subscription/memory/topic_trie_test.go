package memory

import (
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/subscription"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testTopicMatch = []struct {
	subTopic string //subscribe topic
	topic    string //publish topic
	isMatch  bool
}{
	{subTopic: "#", topic: "/abc/def", isMatch: true},
	{subTopic: "/a", topic: "a", isMatch: false},
	{subTopic: "a/#", topic: "a", isMatch: true},
	{subTopic: "+", topic: "/a", isMatch: false},

	{subTopic: "a/", topic: "a", isMatch: false},
	{subTopic: "a/+", topic: "a/123/4", isMatch: false},
	{subTopic: "a/#", topic: "a/123/4", isMatch: true},

	{subTopic: "/a/+/+/abcd", topic: "/a/dfdf/3434/abcd", isMatch: true},
	{subTopic: "/a/+/+/abcd", topic: "/a/dfdf/3434/abcdd", isMatch: false},
	{subTopic: "/a/+/abc/", topic: "/a/dfdf/abc/", isMatch: true},
	{subTopic: "/a/+/abc/", topic: "/a/dfdf/abc", isMatch: false},
	{subTopic: "/a/+/+/", topic: "/a/dfdf/", isMatch: false},
	{subTopic: "/a/+/+", topic: "/a/dfdf/", isMatch: true},
	{subTopic: "/a/+/+/#", topic: "/a/dfdf/", isMatch: true},
}

var topicMatchQosTest = []struct {
	topics     []packet.Topic
	matchTopic struct {
		name string // matched topic name
		qos  uint8  // matched qos
	}
}{
	{
		topics: []packet.Topic{
			{
				SubOptions: packet.SubOptions{
					QoS: packet.QoS1,
				},
				Name: "a/b",
			},
			{
				Name: "a/#",
				SubOptions: packet.SubOptions{
					QoS: packet.QoS2,
				},
			},
			{
				Name: "a/+",
				SubOptions: packet.SubOptions{
					QoS: packet.QoS0,
				},
			},
		},
		matchTopic: struct {
			name string
			qos  uint8
		}{
			name: "a/b",
			qos:  packet.QoS2,
		},
	},
}

var testSubscribeAndFind = struct {
	subTopics  map[string][]packet.Topic // subscription
	findTopics map[string][]struct {     //key by clientID
		exist     bool
		topicName string
		wantQos   uint8
	}
}{
	subTopics: map[string][]packet.Topic{
		"cid1": {
			{
				SubOptions: packet.SubOptions{
					QoS: packet.QoS1,
				}, Name: "t1/t2/+"},
			{SubOptions: packet.SubOptions{
				QoS: packet.QoS2,
			}, Name: "t1/t2/"},
			{SubOptions: packet.SubOptions{
				QoS: packet.QoS0,
			}, Name: "t1/t2/cid1"},
		},
		"cid2": {
			{SubOptions: packet.SubOptions{
				QoS: packet.QoS2,
			}, Name: "t1/t2/+"},
			{SubOptions: packet.SubOptions{
				QoS: packet.QoS1,
			}, Name: "t1/t2/"},
			{SubOptions: packet.SubOptions{
				QoS: packet.QoS0,
			}, Name: "t1/t2/cid2"},
		},
	},
	findTopics: map[string][]struct { //key by clientID
		exist     bool
		topicName string
		wantQos   uint8
	}{
		"cid1": {
			{exist: true, topicName: "t1/t2/+", wantQos: packet.QoS1},
			{exist: true, topicName: "t1/t2/", wantQos: packet.QoS2},
			{exist: false, topicName: "t1/t2/cid2"},
			{exist: false, topicName: "t1/t2/cid3"},
		},
		"cid2": {
			{exist: true, topicName: "t1/t2/+", wantQos: packet.QoS2},
			{exist: true, topicName: "t1/t2/", wantQos: packet.QoS1},
			{exist: false, topicName: "t1/t2/cid1"},
		},
	},
}

var testUnsubscribe = struct {
	subTopics   map[string][]packet.Topic //key by clientID
	unsubscribe map[string][]string       // clientID => topic name
	afterUnsub  map[string][]struct {     // test after unsubscribe, key by clientID
		exist     bool
		topicName string
		wantQos   uint8
	}
}{
	subTopics: map[string][]packet.Topic{
		"cid1": {
			{SubOptions: packet.SubOptions{
				QoS: packet.QoS1,
			}, Name: "t1/t2/t3"},
			{SubOptions: packet.SubOptions{
				QoS: packet.QoS2,
			}, Name: "t1/t2"},
		},
		"cid2": {
			{
				SubOptions: packet.SubOptions{
					QoS: packet.QoS2,
				},
				Name: "t1/t2/t3"},
			{
				SubOptions: packet.SubOptions{
					QoS: packet.QoS1,
				}, Name: "t1/t2"},
		},
	},
	unsubscribe: map[string][]string{
		"cid1": {"t1/t2/t3", "t4/t5"},
		"cid2": {"t1/t2/t3"},
	},
	afterUnsub: map[string][]struct { // test after unsubscribe
		exist     bool
		topicName string
		wantQos   uint8
	}{
		"cid1": {
			{exist: false, topicName: "t1/t2/t3"},
			{exist: true, topicName: "t1/t2", wantQos: packet.QoS2},
		},
		"cid2": {
			{exist: false, topicName: "t1/t2/+"},
			{exist: true, topicName: "t1/t2", wantQos: packet.QoS1},
		},
	},
}

var testPreOrderTraverse = struct {
	topics   []packet.Topic
	clientID string
}{
	topics: []packet.Topic{
		{
			SubOptions: packet.SubOptions{
				QoS: packet.QoS0,
			},
			Name: "a/b/c",
		},
		{
			SubOptions: packet.SubOptions{
				QoS: packet.QoS1,
			},
			Name: "/a/b/c",
		},
		{
			SubOptions: packet.SubOptions{
				QoS: packet.QoS2,
			},
			Name: "b/c/d",
		},
	},
	clientID: "abc",
}

func TestTopicTrie_matchedClients(t *testing.T) {
	a := assert.New(t)
	for _, v := range testTopicMatch {
		trie := newTopicTrie()
		trie.subscribe("cid", &subscription.Subscription{
			TopicFilter: v.subTopic,
		})
		qos := trie.getMatchedTopicFilter(v.topic)
		if v.isMatch {
			a.EqualValues(qos["cid"][0].QoS, 0, v.subTopic)
		} else {
			_, ok := qos["cid"]
			a.False(ok, v.subTopic)
		}
	}
}

func TestTopicTrie_matchedClients_Qos(t *testing.T) {
	a := assert.New(t)
	for _, v := range topicMatchQosTest {
		trie := newTopicTrie()
		for _, tt := range v.topics {
			trie.subscribe("cid", &subscription.Subscription{
				TopicFilter: tt.Name,
				QoS:         tt.QoS,
			})
		}
		rs := trie.getMatchedTopicFilter(v.matchTopic.name)
		a.EqualValues(v.matchTopic.qos, rs["cid"][0].QoS)
	}
}

func TestTopicTrie_subscribeAndFind(t *testing.T) {
	a := assert.New(t)
	trie := newTopicTrie()
	for cid, v := range testSubscribeAndFind.subTopics {
		for _, topic := range v {
			trie.subscribe(cid, &subscription.Subscription{
				TopicFilter: topic.Name,
				QoS:         topic.QoS,
			})
		}
	}
	for cid, v := range testSubscribeAndFind.findTopics {
		for _, tt := range v {
			node := trie.find(tt.topicName)
			if tt.exist {
				a.Equal(tt.wantQos, node.clients[cid].QoS)
			} else {
				if node != nil {
					_, ok := node.clients[cid]
					a.False(ok)
				}
			}
		}
	}
}

func TestTopicTrie_unsubscribe(t *testing.T) {
	a := assert.New(t)
	trie := newTopicTrie()
	for cid, v := range testUnsubscribe.subTopics {
		for _, topic := range v {
			trie.subscribe(cid, &subscription.Subscription{
				TopicFilter: topic.Name,
				QoS:         topic.QoS,
			})
		}
	}
	for cid, v := range testUnsubscribe.unsubscribe {
		for _, tt := range v {
			trie.unsubscribe(cid, tt, "")
		}
	}
	for cid, v := range testUnsubscribe.afterUnsub {
		for _, tt := range v {
			matched := trie.getMatchedTopicFilter(tt.topicName)
			if tt.exist {
				a.EqualValues(matched[cid][0].QoS, tt.wantQos)
			} else {
				a.Equal(0, len(matched))
			}
		}
	}
}

func TestTopicTrie_preOrderTraverse(t *testing.T) {
	a := assert.New(t)
	trie := newTopicTrie()
	for _, v := range testPreOrderTraverse.topics {
		trie.subscribe(testPreOrderTraverse.clientID, &subscription.Subscription{
			TopicFilter: v.Name,
			QoS:         v.QoS,
		})
	}
	var rs []packet.Topic
	trie.preOrderTraverse(func(clientID string, subscription *subscription.Subscription) bool {
		a.Equal(testPreOrderTraverse.clientID, clientID)
		rs = append(rs, packet.Topic{
			SubOptions: packet.SubOptions{
				QoS: subscription.QoS,
			},
			Name: subscription.TopicFilter,
		})
		return true
	})
	a.ElementsMatch(testPreOrderTraverse.topics, rs)
}
