package go_raft

import (
	"log"
)

type actor interface {
	Process(message Message) *MessageResult

	Mode() actorMode
	ModeName() string
}

type actorImpl struct {
	mode         actorMode
	modeStrategy map[actorMode]actorModeStrategy

	persistentStorage PersistentStorage
}

func newActor(serverID ServerID, storage PersistentStorage, gateway ServerGateway, discovery ServerDiscovery) actor {

	// TODO determine if we need this
	vstorage := &VolatileStorage{
		ServerID: serverID,
	}

	subInterpreters := map[actorMode]actorModeStrategy{
		follower:  newFollowerMode(storage, vstorage, gateway, discovery),
		candidate: newCandidateMode(storage, vstorage, gateway, discovery),
		leader:    newLeaderMode(storage, vstorage, gateway, discovery),
	}

	return &actorImpl{
		persistentStorage: storage,
		modeStrategy:      subInterpreters,
	}
}

func (i *actorImpl) Mode() actorMode {
	return i.mode
}

func (i *actorImpl) ModeName() string {
	return i.currentModeStrategy().Name()
}

func (i *actorImpl) Process(message Message) *MessageResult {
	log.Printf("Process message %s", message.Kind)

	// Messages originating from previous terms are discarded
	if i.messageHasExpired(message) {
		log.Printf("Discard message as it has expired")
		return newMessageResult()
	}

	// Messages belonging to newer terms change server into a follower
	if i.messageFromNewTerm(message) {
		log.Printf("New term observed, change into follower")
		i.changeMode(follower, message.Term)
	}

	// Specific messages may trigger a mode change
	if newMode, newTerm := i.currentModeStrategy().PreExecuteModeChange(message); newMode != existing {
		i.changeMode(newMode, newTerm)
	}

	messageResult := i.currentModeStrategy().Process(message)
	if messageResult.NewMode != existing {
		i.changeMode(messageResult.NewMode, messageResult.NewTerm)
	}

	return messageResult
}

func (i *actorImpl) messageFromNewTerm(message Message) bool {
	if message.Kind == msgTick { // TODO generalize
		return false // ticks are exempt
	}

	return message.Term > i.persistentStorage.CurrentTerm()
}

func (i *actorImpl) messageHasExpired(message Message) bool {
	if message.Kind == msgTick { // TODO generalize
		return false // ticks are exempt
	}

	return message.Term < i.persistentStorage.CurrentTerm()
}

func (i *actorImpl) changeMode(newMode actorMode, newTerm Term) {
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
