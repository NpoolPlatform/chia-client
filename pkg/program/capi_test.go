package native

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	result := Add(10, 20)
	assert.Equal(t, result, int32(30))
}
