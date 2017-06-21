package types

type IndexInfo struct {
	IdCount map[string]int64 `json:"idCount"`
	Mapping interface{}      `json:"mapping"`
	Setting struct {
		Analysis struct {
			Analyzer struct {
				Trigrams struct {
					Filter    []string `json:"filter"`
					Tokenizer string   `json:"tokenizer"`
					Type      string   `json:"type"`
				} `json:"trigrams"`
			} `json:"analyzer"`
			Filter struct {
				TrigramsFilter map[string]string `json:"trigrams_filter"`
			} `json:"filter"`
		} `json:"analysis"`
		NumberOfReplicas string `json:"number_of_replicas"`
		NumberOfShards   string `json:"number_of_shards"`
	} `json:"setting"`
}
