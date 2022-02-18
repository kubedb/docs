/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package configgenerator

import (
	"fmt"
	"strings"

	"github.com/iancoleman/orderedmap"
)

const CustomConfigBlockDivider = "#________******kubedb.com/inline-config******________#"

type CustomConfigGenerator struct {
	CurrentConfig      string   // current config string
	RequestedConfig    string   // the requested change that comes from inline custom config
	ConfigBlockDivider string   // This is the divider  String between default custom config and inline config
	KeyValueSeparators []string // KeyValueSeparators is the array with which character a key and value can be separated. for example: a=b. here key `a` and value `b` Separator is '='. This array must need to have one or more value
}

type ValueGenerator struct {
	Value     string
	Separator string
}

// GetMergedConfigString func return new config string where the current config string and requested config strings are merged
func (generator *CustomConfigGenerator) GetMergedConfigString() (string, error) {
	if len(generator.ConfigBlockDivider) == 0 || generator.KeyValueSeparators == nil || len(generator.KeyValueSeparators) == 0 {
		return "", fmt.Errorf("ConfigBlockDivider or KeyValueSeparators is empty or KeyValueSeparators  is null")
	}
	if strings.Count(generator.CurrentConfig, generator.ConfigBlockDivider) > 1 {
		return "", fmt.Errorf("for custom config ConfigBlockDivider string cann't appear multiple time")
	}
	requestedDataMap := ConvertStringInToMap(strings.TrimSpace(generator.RequestedConfig), generator.KeyValueSeparators)
	// there is no valid inline config provided. so just return the current config
	if len(requestedDataMap.Keys()) == 0 {
		return generator.CurrentConfig, nil
	}
	// there is already inline config exist. so the new configs are going to merge with the old one
	if strings.Contains(generator.CurrentConfig, generator.ConfigBlockDivider) {
		curConfig := strings.SplitAfterN(generator.CurrentConfig, generator.ConfigBlockDivider, 2)

		curInlineConfigMap := orderedmap.New()
		if len(curConfig) == 2 {
			curInlineConfigMap = ConvertStringInToMap(strings.TrimSpace(curConfig[1]), generator.KeyValueSeparators)
		}
		inlineConfigString := ConvertMapInToString(MergeAndOverWriteMap(curInlineConfigMap, requestedDataMap))
		return fmt.Sprintf("%s\n%s", curConfig[0], inlineConfigString), nil
	}
	// there is no inline config. so we are going to just add the new one with the provided current config
	return fmt.Sprintf("%s\n\n%s\n%s", generator.CurrentConfig, generator.ConfigBlockDivider, ConvertMapInToString(requestedDataMap)), nil

}
func ConvertStringInToMap(configString string, separators []string) (configData *orderedmap.OrderedMap) {
	configData = orderedmap.New()
	outputs := strings.Split(configString, "\n")
	for _, output := range outputs {
		output = strings.TrimSpace(output)
		// if inline configs any line starts with # is a commented line we are going to ignore this line
		if len(output) == 0 || output[:1] == "#" {
			continue
		}
		for _, separator := range separators {

			values := strings.SplitN(output, separator, 2)

			values[0] = strings.TrimSpace(values[0])
			if len(values) < 2 || strings.Contains(values[0], " ") {
				continue
			}
			values[1] = strings.TrimSpace(values[1])
			val := &ValueGenerator{
				Value:     strings.TrimSpace(values[1]),
				Separator: separator,
			}
			configData.Set(values[0], val)
			break
		}

	}
	return configData
}

func ConvertMapInToString(config *orderedmap.OrderedMap) string {
	configString := ""
	for _, key := range config.Keys() {
		obj, _ := config.Get(key)
		valueGen := obj.(*ValueGenerator)
		configString = configString + fmt.Sprintf("%s%s%s\n", key, valueGen.Separator, valueGen.Value)
	}
	return configString
}

func MergeAndOverWriteMap(config *orderedmap.OrderedMap, incomingConfig *orderedmap.OrderedMap) *orderedmap.OrderedMap {
	for _, key := range incomingConfig.Keys() {
		value, _ := incomingConfig.Get(key)
		config.Set(key, value)
	}
	return config
}
