package matcher

import (
	"fmt"
	"strings"

	"github.com/onsi/gomega/types"
)

func UseCustomConfig(config string) types.GomegaMatcher {
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

	var variableName, variableValue []byte
	for _, rs := range results {
		val, ok := rs["Variable_name"]
		if ok {
			variableName = val
		}
		val, ok = rs["Value"]
		if ok {
			variableValue = val
		}
	}

	if string(variableName) == configPair[0] && string(variableValue) == configPair[1] {
		return true, nil
	}
	return false, nil
}

func (matcher *configMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v to be equivalent to %v", actual, matcher.expected)
}

func (matcher *configMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %v not to be equivalent to %v", actual, matcher.expected)
}
