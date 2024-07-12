package indexer

import (
	"time"
)

type Scheduler struct {
	indexer *Indexer
}

func NewScheduler(indexer *Indexer) *Scheduler {
	return &Scheduler{indexer}
}

func (s *Scheduler) CheckBlock() {
	for {
		time.Sleep(1 * time.Second)

		indexerLastHeight, err := s.indexer.DbRepo.GetLastHeight()
		if err != nil {
			panic(err)
		}
		bitcoinLastHeight, err := s.indexer.bitcoinRepo.GetBlockCount()
		if err != nil {
			panic(err)
		}

		if indexerLastHeight == bitcoinLastHeight {
			continue
		}

		// index new block
		err = s.indexer.IndexBlocks(indexerLastHeight+1, bitcoinLastHeight)
		if err != nil {
			panic(err)
		}
	}
}
