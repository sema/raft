Raft implementation in golang
=============================

Status: Not production ready, incompletely, and certainly wrong.


Server - Implements timing and term triggered state changes
->
ServerState - Implements per-state logic for each event. Assumes (invariant) matching terms.

// TODO
// - outgoing RPCs
// - heartbeats
// - timing adjustments / configuration