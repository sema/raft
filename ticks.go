package go_raft

import "math/rand"

// getTicksWithSplay returns a Tick value between [base, base+splay]
func getTicksWithSplay(base Tick, splay Tick) Tick {
	result := int(base) + rand.Intn(int(splay)+1)
	return Tick(result)
}
