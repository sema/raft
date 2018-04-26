Raft implementation in golang
=============================

Disclaimer: This codebase is written for educational purposes and has yet to see production usage. You have been
warned.

Implementation of the raft algorithm as described in Diego Ongaro's PhD dissertation [1]. Comments in the codebase
will directly point back to the PhD dissertation for reference, and implementation has been intentionally kept as close
as practically possible to the algorithm described in [1].

[1] https://github.com/ongardie/dissertation#readme


----

- Interpreter executing a sequence of events
- Internal state allowing events to have different meaning
- Monitors (async processes)
  - Heartbeat timeout (event source)
  - Election coordinator (event source)
  - Leader heartbeat
  - [other leader related activities]

Question? Monitors internal to states or external:
- Internal: Requires circular dependency back to interpreter, to post back events
  - Alternative: API for states to post arbitrary events, which interpreter picks up and processes
- External: Requires introspection into interpreter state, and generic notification/observer system to react to state changes

Integration testing of interpreter and states (no timing) == event sequences - easier with external
Integration testing of monitors == verify event output with different states
Integration testing everything - put everything together == external events generation, fault injection, etc

Insight: Getting timing events into the event queue solves some details - minimize logic outside of event queue.
Insight: Monitors react to state changes, and trigger events against other servers. Should be deterministic (modulo faults)
Insight: Each server has its own event sequence, and the set of all event sequences, interleaved, describes all
  interaction and state changes to servers.

- Global event sequences can be used to test servers + monitors (the sequence defines input and output to monitors).
  - Faults are not easy to model, as they may not always generate events. Fake "faulty" events may be necessary
- Random fuzz testing must rely on user generated events, and fault injection + quick test style bi-simulation to validate the outcome or general assertions
  - Timing is again an issue here, may require a "ticker" based timing system


server -> stateContext -> serverState
monitor

[external]     / [internal]
(thin wrapper)  (event queue processing and state swap)  (implements handling of events)
server ->       interpreter  ->                          interpreterMode
                   ^- monitor                                <-



	// Heartbeat monitor
    //   Trigger leader election if no heartbeat has been observed within deadline for a term. Reset timer
    //   on term change and on heartbeat (for given term). If deadline is exceeded then trigger change for term.
    //   Input:
    //     - term change
    //     - heartbeat (+ term it belongs to)
    //     - (optional) mode change, to disable/enable events (alternative, have modes ignore the event)

	// Leader election
    //   Start election at specific term, and return back with result (and known term) at some point
    //   Input:
    //     - mode change (+ term at mode change)
    //       + needs access to log data

	// Heartbeat
	//   Send heartbeat data with precise term for leader, possibly other values should also be exact
	//   relative to term? Need to eagerly trigger heartbeats when new logs are ingested.
	//   Input:
	//     - mode change (+ term at mode change)
	//       + needs access to log data
	//     - new logs available
	// Index updates? Based on heartbeats the index will be incremented. We need to either model this as an event or
	// allow the leader to update value directly (lock?)

	// Summary
	// - mode change + term change event
	// - heartbeat + term
	// - Log changes

	// Embedded
	// + Tighter coupling between closely related logic
	// - Need to create cyclic dependency to pass back events

	// External
	// - Easier to test and change out
	// - Need to create a lot of notification features

	// Events
	//
	// - AppendEntries
	// - VoteFor
	//
	// - Heartbeat timeout
	// - Elected as leader
	// - Heartbeat?
	//
	// - Log
