package actor

//go:generate mockgen -destination mocks/mock_gateway.go github.com/sema/raft/pkg/actor ServerGateway

type ServerGateway interface {
	Send(to ServerID, message Message)
}
