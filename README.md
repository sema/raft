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
