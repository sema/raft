Yet another Raft implementation in Go
=====================================

Disclaimer: This Raft implementation exists purely for my own educational purposes.
Please don't decide to use this code anywhere important. Just use etcd's raft implementation.

Implementation of the raft algorithm as described in Diego Ongaro's PhD dissertation [1].

Decided to implement the algorithm white room style, without looking at other implementations other than the
PhD dissertation and the proofs within. The result ended up fairly similar to etcd's implementation, with
much fewer features and less monolithic.

List of TODOs that I would like to get around to doing in this codebase:
 - [RAFT] Implement example state machine and drive it using the raft implementation
 - [RAFT] Writing an actual storage backend, and handle issues around ensuring data is actually persisted (flush to disk)
 - [RAFT] Raft optionals proposed by [1] (membership changes, compaction)
 - [RAFT] Automatic redirection - allow clients to talk to any node
 - [GO] Experiment with godoc and generate proper documentation for package
 - [TESTING] Catch and replay of events for debugging/crash analysis in some practical way
 - [PACKAGE] Add status/debug page

[1] https://github.com/ongardie/dissertation#readme

Development
-----------

The following dependencies are required when working on this codebase:

 go get -u github.com/golang/protobuf/protoc-gen-go
 brew install protobuf
 brew install dep
