package server

import "github.com/yunqi/lighthouse/internal/persistence/subscription"

func defaultIterateOptions(topicName string) subscription.IterationOptions {
	return subscription.IterationOptions{
		Type:      subscription.TypeAll,
		TopicName: topicName,
		MatchType: subscription.MatchFilter,
	}
}
