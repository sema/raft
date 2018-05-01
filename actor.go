package raft

import (
	"log"
)

type Actor interface {
	Process(message Message)

	Mode() ActorMode
	ModeName() string
}

type actorImpl struct {
	mode         ActorMode
	modeStrategy map[ActorMode]actorModeStrategy

	persistentStorage PersistentStorage
}

func NewActor(serverID ServerID, storage PersistentStorage, gateway ServerGateway, config Config) Actor {

	vstorage := &VolatileStorage{
		ServerID: serverID,
	}

	subInterpreters := map[ActorMode]actorModeStrategy{
		FollowerMode:  NewFollowerMode(storage, vstorage, gateway, config),
		CandidateMode: newCandidateMode(storage, vstorage, gateway, config),
		LeaderMode:    newLeaderMode(storage, vstorage, gateway, config),
	}

	actor := &actorImpl{
		persistentStorage: storage,
		modeStrategy:      subInterpreters,
	}
	actor.currentModeStrategy().Enter()

	return actor
}

func (i *actorImpl) Mode() ActorMode {
	return i.mode
}

func (i *actorImpl) ModeName() string {
	return i.currentModeStrategy().Name()
}

func (i *actorImpl) Process(message Message) {
	log.Printf("Process message %s", message.Kind)

	// Messages originating from previous terms are discarded
	if i.messageHasExpired(message) {
		log.Printf("Discard message as it has expired")
		return
	}

	// Messages belonging to newer terms change server into a FollowerMode
	if i.messageFromNewTerm(message) {
		log.Printf("New term observed, change into FollowerMode")
		i.changeMode(FollowerMode, message.Term)
	}

	// Specific messages may trigger a mode change
	if newMode, newTerm := i.currentModeStrategy().PreExecuteModeChange(message); newMode != ExistingMode {
		i.changeMode(newMode, newTerm)
	}

	messageResult := i.currentModeStrategy().Process(message)
	if messageResult.NewMode != ExistingMode {
		i.changeMode(messageResult.NewMode, messageResult.NewTerm)
	}
}

func (i *actorImpl) messageFromNewTerm(message Message) bool {
	if !message.HasTerm() {
		return false
	}

	return message.Term > i.persistentStorage.CurrentTerm()
}

func (i *actorImpl) messageHasExpired(message Message) bool {
	if !message.HasTerm() {
		return false
	}

	return message.Term < i.persistentStorage.CurrentTerm()
}

func (i *actorImpl) changeMode(newMode ActorMode, newTerm Term) {
	log.Printf("Change mode %d(%d) -> %d(%d)", i.mode, i.persistentStorage.CurrentTerm(), newMode, newTerm)

	if newTerm != i.persistentStorage.CurrentTerm() {
		i.persistentStorage.SetCurrentTerm(newTerm)
		i.persistentStorage.ClearVotedFor()
	}

	i.currentModeStrategy().Exit()
	i.mode = newMode
	i.currentModeStrategy().Enter()
}

func (i *actorImpl) currentModeStrategy() actorModeStrategy {
	return i.modeStrategy[i.mode]
}
