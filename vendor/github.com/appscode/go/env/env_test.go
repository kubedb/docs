package env_test

import (
	"os"
	"testing"

	"github.com/appscode/go/env"
	"github.com/stretchr/testify/assert"
)

func TestDetectFromHostProd(t *testing.T) {
	os.Setenv(env.Key, env.Prod.String())
	assert.Equal(t, env.Prod, env.FromHost())
}

func TestDetectFromHostOnebox(t *testing.T) {
	os.Setenv(env.Key, env.Onebox.String())
	assert.Equal(t, env.Onebox, env.FromHost())
}

func TestDetectFromHostQA(t *testing.T) {
	os.Setenv(env.Key, env.QA.String())
	assert.Equal(t, env.QA, env.FromHost())
}

func TestDetectFromHostDev(t *testing.T) {
	os.Setenv(env.Key, env.Dev.String())
	assert.Equal(t, env.Dev, env.FromHost())
}

func TestEnvironment_MarshalJSON_Multiple(t *testing.T) {
	assert := assert.New(t)

	e := env.Prod
	data, err := e.MarshalJSON()
	assert.Nil(err)
	assert.Equal(`"prod"`, string(data))
}

func TestEnvironment_UnmarshalJSON_Empty(t *testing.T) {
	assert := assert.New(t)

	var a env.Environment
	err := a.UnmarshalJSON([]byte(`"prod"`))
	assert.Nil(err)
	assert.Equal(env.Prod, a)
}
