package indexer

import (
	"log"
	"time"
)

type Scheduler struct {
	indexer *Indexer
}

const MAX_BLOCK_FLUSH = 100
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
			err := s.SyncBlocks(indexerLastHeight+1, bitcoinLastHeight)
			if err != nil {
				panic(err)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *Scheduler) SyncBlocks(startHeight int32, endHeight int32) error {
	if endHeight-startHeight > MAX_BLOCK_FLUSH {
		// Many blocks to sync, log the sync process
		log.Printf("Syncing %d blocks from height %d to %d", endHeight-startHeight, startHeight, endHeight)

		for i := startHeight; i <= endHeight; i += MAX_BLOCK_FLUSH {
			endBlock := i + MAX_BLOCK_FLUSH
			if endBlock > endHeight {
				endBlock = endHeight
			}

			err := s.indexer.IndexBlocks(i, endBlock)
			if err != nil {
				return err
			}
			// Increment i after the block is indexed
			i++
		}
	} else {
		// Just a few blocks to add, sync one by one
		for i := startHeight; i <= endHeight; i++ {
			log.Printf("Indexing block %d", i)

			// Check for reorg before indexing each block
			reorgHeight, err := s.indexer.Reorg(i, REORG_DEPTH_CHECK)
			if err != nil {
				return err
			}

			if reorgHeight > 0 {
				i = reorgHeight
			}

			err = s.indexer.IndexBlocks(i, i)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
