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