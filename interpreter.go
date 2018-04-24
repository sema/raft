package go_raft

import (
	"log"
)

type interpreter interface {
	Execute(command Command) *CommandResult

	Mode() interpreterMode
	ModeName() string
}

type commandKind string

const (
	cmdAppendEntries = "cmdAppendEntries"
	cmdVoteFor = "cmdVoteFor"
	cmdVoteForResponse = "cmdVoteForResponse"

	// cmdTick = "cmdTick"

	cmdStartLeaderElection = "cmdStartLeaderElection"
	cmdElectedAsLeader = "cmdElectedAsLeader"
)

type Command struct {
	Kind commandKind

	Term Term

	LastLogTerm Term
	LastLogIndex LogIndex

	PreviousLogTerm Term
	PreviousLogIndex LogIndex

	LeaderCommit LogIndex

	VoteGranted bool

	From ServerID
}

type interpreterImpl struct {
	mode interpreterMode
	subInterpreters map[interpreterMode]serverState  // TODO rename

	persistentStorage PersistentStorage
}

func newInterpreter(serverID ServerID, storage PersistentStorage, gateway ServerGateway, discovery ServerDiscovery) interpreter {

	// TODO determine if we need this
	vstorage := &VolatileStorage{
		ServerID: serverID,
	}

	subInterpreters := map[interpreterMode]serverState{
		follower: newFollowerState(storage, vstorage, gateway, discovery),
		candidate: newCandidateState(storage, vstorage, gateway, discovery),
		leader: newLeaderState(storage, vstorage, gateway, discovery),
	}

	return &interpreterImpl{
		persistentStorage: storage,
		subInterpreters: subInterpreters,
	}
}

func (i *interpreterImpl) Mode() interpreterMode {
	return i.mode
}

func (i *interpreterImpl) ModeName() string {
	return i.currentSubInterpreter().Name()
}

func (i *interpreterImpl) Execute(command Command) *CommandResult {
	log.Printf("Execute command %s", command.Kind)

	// Commands originating from previous terms are discarded
	if i.commandHasExpired(command) {
		log.Printf("Discard command as it has expired")
		return newCommandResult(false, i.persistentStorage.CurrentTerm())
	}

	// Commands belonging to newer terms change server into a follower
	if i.commandFromNewTerm(command) {
		log.Printf("New term observed, change into follower")
		i.changeMode(follower, command.Term)
	}

	// Specific commands may trigger a mode change
	if newMode, newTerm := i.currentSubInterpreter().PreExecuteModeChange(command); newMode != existing {
		i.changeMode(newMode, newTerm)
	}

	commandResult := i.currentSubInterpreter().Execute(command)
	if commandResult.NewMode != existing {
		i.changeMode(commandResult.NewMode, commandResult.NewTerm)
	}

	return commandResult
}

func (i *interpreterImpl) commandFromNewTerm(command Command) bool {
	return command.Term > i.persistentStorage.CurrentTerm()
}

func (i *interpreterImpl) commandHasExpired(command Command) bool {
	return command.Term < i.persistentStorage.CurrentTerm()
}

func (i *interpreterImpl) changeMode(newMode interpreterMode, newTerm Term) {
	log.Printf("Change mode %d(%d) -> %d(%d)", i.mode, i.persistentStorage.CurrentTerm(), newMode, newTerm)

	if newTerm != i.persistentStorage.CurrentTerm() {
		i.persistentStorage.SetCurrentTerm(newTerm)
		i.persistentStorage.ClearVotedFor()
	}

	i.currentSubInterpreter().Exit()
	i.mode = newMode
	i.currentSubInterpreter().Enter()
}

func (i * interpreterImpl) currentSubInterpreter() serverState {
	return i.subInterpreters[i.mode]
}
