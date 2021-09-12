package packet

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidTopicName(t *testing.T) {

	var topicNameTest = []struct {
		input string
		want  bool
	}{
		{input: "sport/tennis#", want: false},
		{input: "sport/tennis/#/rank", want: false},
		{input: "//1", want: true},
		{input: "/+1", want: false},
		{input: "+", want: false},
		{input: "#", want: false},
		{input: "sport/tennis/#", want: false},
		{input: "/1/+/#", want: false},
		{input: "/1/+/+/1234", want: false},
		{input: "/abc/def/gggggg/", want: true},
		{input: "/9 2", want: true},
		{input: "", want: false},
		{input: string(make([]byte, TopicMaxLen+1)), want: false},
	}
	for i, data := range topicNameTest {
		t.Run(fmt.Sprintf("[%d]", i), func(t *testing.T) {
			assert.Equal(t, data.want, ValidTopicName(true, []byte(data.input)))
		})
	}
}
func TestValidTopicFilter(t *testing.T) {

	var topicFilterTest = []struct {
		input string
		want  bool
	}{
		{input: "sport/tennis#", want: false},
		{input: "sport/tennis/#/rank", want: false},
		{input: "//1", want: true},
		{input: "/+1", want: false},
		{input: "+", want: true},
		{input: "#", want: true},
		{input: "sport/tennis/#", want: true},
		{input: "/1/+/#", want: true},
		{input: "/1/+/+/1234", want: true},
		{input: "##", want: false},
		{input: "#/", want: false},
		{input: "", want: false},
		{input: string(make([]byte, TopicMaxLen+1)), want: false},
	}
	for i, topic := range topicFilterTest {
		t.Run(fmt.Sprintf("[%d]", i), func(t *testing.T) {
			assert.Equal(t, topic.want, ValidTopicFilter(true, []byte(topic.input)))
		})
	}

}

func TestIsInternalTopic(t *testing.T) {
	var internalTopicTest = []struct {
		input string
		want  bool
	}{
		{input: "$sys", want: true},
		{input: "$", want: true},
		{input: "", want: false},
	}
	for _, topic := range internalTopicTest {
		t.Run(fmt.Sprintf("[%s]", topic.input), func(t *testing.T) {
			assert.Equal(t, topic.want, IsInternalTopic(topic.input))
		})
	}

}
