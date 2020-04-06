package helpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainsString(t *testing.T) {
	array := [3]string{"str", "plop", "plip"}

	ret := ContainsString(array[:], "plap")
	assert.Equal(t, false, ret)
	ret = ContainsString(array[:], "plop")
	assert.Equal(t, true, ret)
}
