package go_raft

type Discovery interface {
	Servers() []ServerID
}

type staticDiscovery struct {
	servers []ServerID
}

func NewStaticDiscovery(servers []ServerID) Discovery {
	return &staticDiscovery{
		servers: servers,
	}
}

func (d *staticDiscovery) Servers() []ServerID {
	return d.servers
}
