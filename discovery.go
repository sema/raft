package go_raft

import "math"

type ServerDiscovery interface {
	Servers() []ServerID
	Quorum() int
}

type staticDiscovery struct {
	servers []ServerID
}

func NewStaticDiscovery(servers []ServerID) ServerDiscovery {
	return &staticDiscovery{
		servers: servers,
	}
}

func (d *staticDiscovery) Servers() []ServerID {
	return d.servers
}

func (d *staticDiscovery) Quorum() int {
	return int(math.Floor(float64(len(d.servers))/float64(2))) + 1
}
