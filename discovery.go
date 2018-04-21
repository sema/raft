package go_raft

type Discovery interface {
	Servers() []ServerID
}

type discovery struct {

}

func (d* discovery) Servers() []ServerID {
	// TODO implement
	return nil
}