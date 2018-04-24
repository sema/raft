package go_raft

import "fmt"

type candidateState struct {
	persistentStorage PersistentStorage
	volatileStorage   *VolatileStorage
	gateway           ServerGateway
	discovery         ServerDiscovery

	votes map[ServerID]bool
}

func newCandidateState(persistentStorage PersistentStorage, volatileStorage *VolatileStorage, gateway ServerGateway, discovery ServerDiscovery) serverState {
	return &candidateState{
			persistentStorage: persistentStorage,
			volatileStorage:   volatileStorage,
			gateway:           gateway,
			discovery:         discovery,
		}
}

func (s *candidateState) PreExecuteModeChange(command Command) (newMode interpreterMode, newTerm Term) {
	if command.Kind == cmdAppendEntries {
		return follower, command.Term
	}

	return existing, 0
}

func (s *candidateState) Execute(command Command) (result *CommandResult) {
	switch command.Kind {
	case cmdVoteFor:
		return s.handleRequestVote(command)
	case cmdVoteForResponse:
		return s.handleRequestVoteResponse(command)
	case cmdStartLeaderElection:
		return s.handleStartLeaderElection(command)
	default:
		panic(fmt.Sprintf("Unexpected Command %s passed to candidate", command.Kind))
	}
}

func (s *candidateState) Name() (name string) {
	return "candidate"
}

func (s *candidateState) handleStartLeaderElection(command Command) *CommandResult {
	result := newCommandResult(true, s.persistentStorage.CurrentTerm())
	result.ChangeMode(candidate, s.persistentStorage.CurrentTerm() + 1)
	return result
}

func (s *candidateState) handleRequestVote(command Command) (result *CommandResult) {
	// Multiple candidates in same Term. stateContext always votes for itself when entering candidate state, so
	// skip voting process.

	s.gateway.SendRequestVoteResponseRPC(command.From, s.volatileStorage.ServerID, s.persistentStorage.CurrentTerm(), false)
	return newCommandResult(false, s.persistentStorage.CurrentTerm())
}

func (s *candidateState) handleRequestVoteResponse(command Command) (result *CommandResult) {
	s.votes[command.From] = command.VoteGranted

	votesPositive := 0
	votesNegative := 0
	for serverID := range s.votes {
		if s.votes[serverID] {
			votesPositive += 1
		} else {
			votesNegative += 1
		}
	}

	quorum := len(s.discovery.Servers())/2

	// TODO fail fast if majority rejects leader?
	if votesPositive > quorum {
		result = newCommandResult(true, s.persistentStorage.CurrentTerm())
		result.ChangeMode(leader, s.persistentStorage.CurrentTerm())
		return result
	}

	return newCommandResult(false, s.persistentStorage.CurrentTerm())
}


func (s *candidateState) Enter() {
	s.votes = make(map[ServerID]bool)

	// vote for ourselves
	s.votes[s.volatileStorage.ServerID] = true
	s.persistentStorage.SetVotedForIfUnset(s.volatileStorage.ServerID) // TODO ensure this happens

	// Send RPCs
	for _, serverID := range s.discovery.Servers() {
		if serverID == s.volatileStorage.ServerID {
			continue // skip self
		}

		logEntry := s.persistentStorage.LatestLogEntry()

		// TODO retries?
		s.gateway.SendRequestVoteRPC(serverID, s.volatileStorage.ServerID, s.persistentStorage.CurrentTerm(), logEntry.Index, logEntry.Term)
	}
}

func (s *candidateState) Exit() {
	// TODO we need to to tear down the request votes thing
}

// TODO send request vote RPCs to all other leaders
// If majority of hosts send votes then transition to leader

// TODO, remember, this may be triggered while already a candidate, should trigger new election
