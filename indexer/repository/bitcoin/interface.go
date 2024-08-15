package bitcoin

type BitcoinRepositoryInterface interface {
	GetBlockHash(height int32) (string, error)
	GetBlock(blockHash string) (*GetBlock, error)
	GetBlockCount() (int32, error)
}
