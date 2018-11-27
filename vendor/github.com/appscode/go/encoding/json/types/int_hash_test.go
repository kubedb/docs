package types_test

import (
	"encoding/json"
	"testing"

	. "github.com/appscode/go/encoding/json/types"
	"github.com/stretchr/testify/assert"
)

func TestIntHash_Empty(t *testing.T) {
	assert := assert.New(t)

	var x IntHash
	err := x.UnmarshalJSON([]byte(`""`))
	assert.Nil(err)
}

func TestIntHash(t *testing.T) {
	assert := assert.New(t)

	type Example struct {
		A IntHash
		B IntHash
		C IntHash
		D IntHash
		E *IntHash
		F *IntHash `json:",omitempty"`
		G IntHash
	}
	s := `{
		"A": "0$str\\",
		"B": 1,
		"C": "8$xyz",
		"E": null
	}`

	var e Example
	err := json.Unmarshal([]byte(s), &e)
	assert.Nil(err)

	b, err := json.Marshal(&e)
	assert.Nil(err)
	assert.Equal(`{"A":"0$str\\","B":1,"C":"8$xyz","D":0,"E":null,"G":0}`, string(b))
}
