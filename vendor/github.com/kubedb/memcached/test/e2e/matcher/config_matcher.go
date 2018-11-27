package matcher

import (
	"fmt"

	"github.com/kubedb/memcached/test/e2e/framework"
	"github.com/onsi/gomega/types"
)

func UseCustomConfig(config framework.MemcdConfig) types.GomegaMatcher {
	return &configMatcher{
		expected: config,
	}
}

type configMatcher struct {
	expected framework.MemcdConfig
}

func (matcher *configMatcher) Match(actual interface{}) (success bool, err error) {
	// TODO
	return false, nil
}

func (matcher *configMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to be equivalent to %v", actual, matcher.expected)
}

func (matcher *configMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to be equivalent to %v", actual, matcher.expected)
}
