Raft implementation in golang
=============================

Disclaimer: This codebase is written for educational purposes and has yet to see production usage. You have been
warned.

Implementation of the raft algorithm as described in Diego Ongaro's PhD dissertation [1]. Comments in the codebase
will directly point back to the PhD dissertation for reference, and implementation has been intentionally kept as close
as practically possible to the algorithm described in [1].

[1] https://github.com/ongardie/dissertation#readme

- Use something else, such as etcd/raft
- This is my playground, TODOs I want to work on
  - Experiment with godoc and generate proper documentation for package
  - Writing an actual storage backend, and handle issues around ensuring data is actually persisted (flush to disk)
  - Raft optionals proposed by [1] (membership changes, compaction)
  - Automatic redirection - allow clients to talk to any node
  - Catch and replay of events for debugging/crash analysis
