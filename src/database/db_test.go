package database

import (
	asserthelper "github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDB(t *testing.T) {
	assert := asserthelper.New(t)

	assert.Equal(GetDB(), GetDB())
}
