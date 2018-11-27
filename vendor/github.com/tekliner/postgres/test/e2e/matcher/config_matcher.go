package matcher

import (
	"fmt"
	"strings"

	"github.com/onsi/gomega/types"
)

func Use(config string) types.GomegaMatcher {
	return &configMatcher{
		expected: config,
	}
}

type configMatcher struct {
	expected string
}

func (matcher *configMatcher) Match(actual interface{}) (success bool, err error) {
	results := actual.([]map[string][]byte)
	configPair := strings.Split(matcher.expected, "=")

	for _, rs := range results {
		value, ok := rs[configPair[0]]
		if !ok {
			return false, nil
		}
		if string(value) == configPair[1] {
			return true, nil
		}
	}
	return false, nil
}

func (matcher *configMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to be equivalent to %v", actual, matcher.expected)
}

func (matcher *configMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to be equivalent to %v", actual, matcher.expected)
}
