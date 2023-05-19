package service

import (
	"sync"
	"time"
)

type SequenceService interface {
	Generate() int64
}

func Sequence() SequenceService {
	return &sequenceService{}
}

type sequenceService struct {
	mutex  sync.Mutex
	lastID int64
}

func (seq *sequenceService) Generate() int64 {
	seq.mutex.Lock()
	defer seq.mutex.Unlock()

	for {
		if nano := time.Now().UnixNano(); nano != seq.lastID {
			seq.lastID = nano
			return nano
		}
	}
}
