package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Quorum(t *testing.T) {
	config := Config{
		Servers: []ServerID{"server1"},
	}
	assert.Equal(t, 1, config.Quorum())

	config = Config{
		Servers: []ServerID{"server1", "server2"},
	}
	assert.Equal(t, 2, config.Quorum())

	config = Config{
		Servers: []ServerID{"server1", "server2", "server3"},
	}
	assert.Equal(t, 2, config.Quorum())

	config = Config{
		Servers: []ServerID{"server1", "server2", "server3", "server4"},
	}
	assert.Equal(t, 3, config.Quorum())
}
