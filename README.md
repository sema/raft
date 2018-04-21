Raft implementation in golang
=============================

Disclaimer: This codebase is written for educational purposes and has yet to see production usage. You have been
warned.

Implementation of the raft algorithm as described in Diego Ongaro's PhD dissertation [1]. Comments in the codebase
will directly point back to the PhD dissertation for reference, and implementation has been intentionally kept as close
as practically possible to the algorithm described in [1].

[1] https://github.com/ongardie/dissertation#readme

----

Server - Implements timing and term triggered state changes
->
ServerState - Implements per-state logic for each event. Assumes (invariant) matching terms.

// TODO
// - outgoing RPCs
// - heartbeats
// - timing adjustments / configuration


- Everything in one pool, if-s to handle differences -- NO
- Differences in states, more complex server with knowledge about states, async manipulates server
  - Easy locking at server level
  - Easier to coordinate state changes (done in server)
  - Unintuitive separation of logic - monitors need to know about states to adjust their runtime, or they need to
    blindly call no-op methods implemented in states. Logic for a state spread out between two distinct structures.
- Everything in different states OR Layered states, simple/generic server
  - Async processes need a feedback loop to trigger a state change in server
  - Locking must be done at lower level (not server)


States
- Follower
- Candidate
- Leader

Sync processes
- AppendEntries
- Vote

Async processes
- Heartbeat timeout
- Heartbeat / Log replication (trigger eager heartbeat?)

Problems:
- Locking
