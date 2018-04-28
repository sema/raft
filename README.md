Raft implementation in golang
=============================

Disclaimer: This codebase is written for educational purposes and has yet to see production usage. You have been
warned.

Implementation of the raft algorithm as described in Diego Ongaro's PhD dissertation [1]. Comments in the codebase
will directly point back to the PhD dissertation for reference, and implementation has been intentionally kept as close
as practically possible to the algorithm described in [1].

[1] https://github.com/ongardie/dissertation#readme


----

State
- CommitIndex
- Term (persistent)
- Log (persistent)

- MatchIndex (leader)
- NextIndex (leader)

Semantics
---

- AppendEntriesResponse (resp)
  - Success: MatchIndex = resp.MatchIndex;
             NextIndex = resp.MatchIndex + 1
  - Failure: NextIndex = max(1, NextIndex - 1)

- AppendEntries (req)
  - Failure (if prev* does not match): send resp with
       resp.success = false;
       resp.matchIndex = 0
  - Success


Process to agree on match/next index
-------------------------------------

