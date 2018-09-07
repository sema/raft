package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTicksWithSplay__ZeroSplayReturnsBase(t *testing.T) {
	result := getTicksWithSplay(Tick(5), Tick(0))
	assert.Equal(t, Tick(5), result)
}
