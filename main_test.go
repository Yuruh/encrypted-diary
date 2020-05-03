package main

import (
	asserthelper "github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestEnsureEnvSet(t *testing.T) {
	assert := asserthelper.New(t)

	os.Setenv("2FA_TOKEN_SECRET", "")
	err := EnsureEnvSet()
	assert.NotNil(err)
	os.Setenv("2FA_TOKEN_SECRET", "Secret value")
	err = EnsureEnvSet()
	assert.Nil(err)
}

func TestInitSentry(t *testing.T) {
	assert := asserthelper.New(t)

	assert.NotPanics(func() {
		InitSentry()
	})
}