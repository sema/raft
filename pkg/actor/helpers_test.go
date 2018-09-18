package actor_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sema/raft/pkg/actor"
	"github.com/stretchr/testify/assert"
)

const localServerID = actor.ServerID("server1.servers.local")
const peerServer1ID = actor.ServerID("server2.servers.local")
const peerServer2ID = actor.ServerID("server3.servers.local")

// Base test setup for an actor, with in-memory storage, and 2 peer actors.
func newActorTestSetup(t *testing.T) (
	actor.Actor, actor.Storage, func()) {

	mockCtrl := gomock.NewController(t)
	cleanup := mockCtrl.Finish

	storage := actor.NewMemoryStorage()

	config := actor.Config{
		Servers:                    []actor.ServerID{localServerID, peerServer1ID, peerServer2ID},
		LeaderElectionTimeout:      10,
		LeaderElectionTimeoutSplay: 0,
		LeaderHeartbeatFrequency:   5,
	}

	act := actor.NewActor(localServerID, storage, config)

	return act, storage, cleanup
}

func testTransitionFromFollowerToCandidate(t *testing.T, act actor.Actor, storage actor.Storage) {
	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	messagesExpected := []actor.Message{
		actor.NewMessageVoteFor(peerServer1ID, localServerID, actor.Term(1), prevIndex, prevTerm),
		actor.NewMessageVoteFor(peerServer2ID, localServerID, actor.Term(1), prevIndex, prevTerm),
	}
	messagesOut := testProgressTime(act, 11)
	assert.Equal(t, messagesExpected, messagesOut)
}

func testProgressTime(act actor.Actor, numTicks int) []actor.Message {
	var messages []actor.Message

	for i := 0; i < numTicks; i++ {
		messagesOut := act.Process(actor.NewMessageTick(localServerID, localServerID))
		messages = append(messages, messagesOut...)
	}

	return messages
}

func testTransitionFromFollowerToLeader(t *testing.T, act actor.Actor, storage actor.Storage) {
	testTransitionFromFollowerToCandidate(t, act, storage)

	prevIndex := storage.LatestLogEntry().Index
	prevTerm := storage.LatestLogEntry().Term

	messagesExpected := []actor.Message{
		actor.NewMessageAppendEntries(peerServer1ID, localServerID, actor.Term(1), 0, prevIndex, prevTerm, nil),
		actor.NewMessageAppendEntries(peerServer2ID, localServerID, actor.Term(1), 0, prevIndex, prevTerm, nil),
	}
	messagesOut := act.Process(actor.NewMessageVoteForResponse(
		localServerID, peerServer1ID, actor.Term(1), true))
	assert.Equal(t, messagesExpected, messagesOut)

}

func testFollowersReplicateUp(act actor.Actor, storage actor.Storage) {
	term := storage.CurrentTerm()
	lastIndex := storage.LatestLogEntry().Index

	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer1ID, term, true, lastIndex))
	act.Process(actor.NewMessageAppendEntriesResponse(
		localServerID, peerServer2ID, term, true, lastIndex))
}
