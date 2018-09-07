package actor

type LogIndex uint64

func MaxLogIndex(v1 LogIndex, v2 LogIndex) LogIndex {
	if v1 > v2 {
		return v1
	}

	return v2
}

func MinLogIndex(v1 LogIndex, v2 LogIndex) LogIndex {
	if v1 < v2 {
		return v1
	}

	return v2
}
