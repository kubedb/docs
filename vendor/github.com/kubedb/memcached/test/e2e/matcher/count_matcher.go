package matcher

import (
	"fmt"

	"github.com/onsi/gomega/types"
)

func MoreThan(expected int) types.GomegaMatcher {
	return &countMatcher{
		expected: expected,
	}
}

type countMatcher struct {
	expected int
}

func (matcher *countMatcher) Match(actual interface{}) (success bool, err error) {
	total := actual.(int)
	return total >= matcher.expected, nil
}

func (matcher *countMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected to have snapshot more than %v", matcher.expected)
}

func (matcher *countMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected to have snapshot more than %v", matcher.expected)
}
