package go_raft

type ServerDiscovery interface {
	Servers() []ServerID
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
