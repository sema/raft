package raft

//go:generate mockgen -destination=mocks/mock_gateway.go -source=gateway.go

type ServerGateway interface {
	Send(to ServerID, message Message)
}
