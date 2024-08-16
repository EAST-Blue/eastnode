package bitcoin

type BitcoinRepositoryInterface interface {
	GetBlockHash(height int32) (string, error)
	GetBlock(blockHash string) (*GetBlock, error)
	GetBlockWithVerbosity(blockHash string, verbosity int32) (*GetBlock, error)
	GetBlockCount() (int32, error)
}
