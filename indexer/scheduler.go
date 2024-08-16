package indexer

import (
	"time"
)

type Scheduler struct {
	indexer *Indexer
}

const REORG_DEPTH_CHECK = 6

func NewScheduler(indexer *Indexer) *Scheduler {
	return &Scheduler{indexer}
}

func (s *Scheduler) Start() {
	for {
		indexerLastHeight, err := s.indexer.DbRepo.GetLastHeight()
		if err != nil {
			panic(err)
		}
		bitcoinLastHeight, err := s.indexer.bitcoinRepo.GetBlockCount()
		if err != nil {
			panic(err)
		}

		if indexerLastHeight == bitcoinLastHeight {
			// if fully synced, continue
			time.Sleep(15 * time.Second)
			continue
		} else {
			// if not fully synced, sync blocks
			err := s.indexer.SyncBlocks(indexerLastHeight+1, bitcoinLastHeight)
			if err != nil {
				panic(err)
			}
		}

		time.Sleep(1 * time.Second)
	}
}
