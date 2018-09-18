package grpcserver

import (
	"github.com/sema/raft/pkg/actor"
	pb "github.com/sema/raft/pkg/grpcserver/pb"
)

func AsGRPCMessage(message actor.Message) *pb.Message {
	var logEntries []*pb.LogEntry
	for _, logEntry := range message.LogEntries {
		logEntries = append(logEntries, &pb.LogEntry{
			Term:    uint64(logEntry.Term),
			Index:   uint64(logEntry.Index),
			Payload: logEntry.Payload,
		})
	}

	return &pb.Message{
		Kind:             string(message.Kind),
		From:             string(message.From),
		To:               string(message.To),
		Term:             uint64(message.Term),
		LastLogTerm:      uint64(message.LastLogTerm),
		LastLogIndex:     uint64(message.LastLogIndex),
		VoteGrated:       message.VoteGranted,
		PreviousLogTerm:  uint64(message.PreviousLogTerm),
		PreviousLogIndex: uint64(message.PreviousLogIndex),
		LeaderCommit:     uint64(message.LeaderCommit),
		LogEntries:       logEntries,
		Success:          message.Success,
		MatchIndex:       uint64(message.MatchIndex),
		ProposalPayload:  message.ProposalPayload,
	}
}

func AsNativeMessage(message *pb.Message) actor.Message {
	var logEntries []actor.LogEntry
	for _, entry := range message.LogEntries {
		logEntries = append(logEntries, actor.LogEntry{
			Term:    actor.Term(entry.Term),
			Index:   actor.LogIndex(entry.Index),
			Payload: entry.Payload,
		})
	}

	return actor.Message{
		Kind:             actor.MessageKind(message.Kind),
		From:             actor.ServerID(message.From),
		To:               actor.ServerID(message.To),
		Term:             actor.Term(message.Term),
		LastLogTerm:      actor.Term(message.LastLogTerm),
		LastLogIndex:     actor.LogIndex(message.LastLogIndex),
		VoteGranted:      message.VoteGrated,
		PreviousLogTerm:  actor.Term(message.PreviousLogTerm),
		PreviousLogIndex: actor.LogIndex(message.PreviousLogIndex),
		LeaderCommit:     actor.LogIndex(message.LeaderCommit),
		LogEntries:       logEntries,
		Success:          message.Success,
		MatchIndex:       actor.LogIndex(message.MatchIndex),
		ProposalPayload:  message.ProposalPayload,
	}
}
