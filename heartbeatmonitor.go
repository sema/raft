package go_raft

import (
	"math/rand"
	"time"
	"log"
)

// HeartbeatMonitor keeps track of leader heartbeats, and provides a signal if no leader has been observed within
// a randomized deadline.
type HeartbeatMonitor interface {
	Run()
	Signal() <-chan time.Time
	RecordHeartbeat()
}

type heartbeatMonitor struct {
	timer  *time.Timer
	signal chan time.Time

	timeout time.Duration
	splay   time.Duration
}

func NewHeartbeatMonitor(timeout time.Duration, splay time.Duration) HeartbeatMonitor {
	signal := make(chan time.Time, 1)
	timer := time.NewTimer(randomTimeout(timeout, splay))

	return &heartbeatMonitor{
		timer:   timer,
		signal:  signal,
		timeout: timeout,
		splay:   splay,
	}
}

func (s *heartbeatMonitor) Run() {
	log.Print("Starting heartbeat monitoring")

	// TODO figure out if this is necessary
	for {
		select {
		case t := <-s.timer.C:
			log.Print("Heartbeat not observed, signal listeners")
			s.signal <- t
			s.resetTimer()
		}
	}

	log.Print("Stopping hearbeat monitoring")
}

func (s *heartbeatMonitor) RecordHeartbeat() {
	s.resetTimer()
}

func (s *heartbeatMonitor) Signal() <-chan time.Time {
	return s.signal
}

func (s *heartbeatMonitor) resetTimer() {
	d := randomTimeout(s.timeout, s.splay)

	// TODO carefully read the documentation for reset and make sure we use this in a safe way
	s.timer.Reset(d)

	log.Printf("Heartbeat timeout reset for %s", d)
}

func randomTimeout(timeout, splay time.Duration) time.Duration {
	randomSplay := time.Duration(rand.Int63n(splay.Nanoseconds()))
	return timeout + randomSplay
}
