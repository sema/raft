package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxLogIndex(t *testing.T) {
	assert.Equal(t, LogIndex(3), MaxLogIndex(2, 3))
	assert.Equal(t, LogIndex(3), MaxLogIndex(3, 2))
}

func TestMinLogIndex(t *testing.T) {
	assert.Equal(t, LogIndex(2), MinLogIndex(2, 3))
	assert.Equal(t, LogIndex(2), MinLogIndex(3, 2))
}
