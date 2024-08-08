package bitcoin

import "encoding/json"

type Request struct {
	Jsonrpc string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

type GetBlockHashRPC struct {
	Result string `json:"result"`
}

type GetBlockRPC struct {
	Result GetBlock `json:"result"`
}

type GetBlock struct {
	Hash              string  `json:"hash"`
	Confirmations     int     `json:"confirmations"`
	Height            int     `json:"height"`
	Version           int     `json:"version"`
	VersionHex        string  `json:"versionHex"`
	Merkleroot        string  `json:"merkleroot"`
	Time              int     `json:"time"`
	Mediantime        int     `json:"mediantime"`
	Nonce             int     `json:"nonce"`
	Bits              string  `json:"bits"`
	Difficulty        float64 `json:"difficulty"`
	Chainwork         string  `json:"chainwork"`
	NTx               int     `json:"nTx"`
	Previousblockhash string  `json:"previousblockhash"`
	Nextblockhash     string  `json:"nextblockhash"`
	Strippedsize      int     `json:"strippedsize"`
	Size              int     `json:"size"`
	Weight            int     `json:"weight"`
	Tx                []struct {
		Txid     string `json:"txid"`
		Hash     string `json:"hash"`
		Version  int    `json:"version"`
		Size     int    `json:"size"`
		Vsize    int    `json:"vsize"`
		Weight   int    `json:"weight"`
		Locktime int    `json:"locktime"`
		Vin      []struct {
			Coinbase    string   `json:"coinbase"`
			Txid        string   `json:"txid"`
			Vout        int64    `json:"vout"`
			Txinwitness []string `json:"txinwitness"`
			Sequence    int64    `json:"sequence"`
			ScriptSig   struct {
				Asm string `json:"asm"`
				Hex string `json:"hex"`
			} `json:"scriptSig"`
			PrevOutput struct {
				Value        float64 `json:"value"`
				ScriptPubKey struct {
					Asm     string `json:"asm"`
					Desc    string `json:"desc"`
					Hex     string `json:"hex"`
					Address string `json:"address"`
					Type    string `json:"type"`
				} `json:"scriptPubKey"`
			} `json:"prevout"`
		} `json:"vin"`
		Vout []struct {
			Value        float64 `json:"value"`
			N            int     `json:"n"`
			ScriptPubKey struct {
				Asm     string `json:"asm"`
				Desc    string `json:"desc"`
				Hex     string `json:"hex"`
				Address string `json:"address"`
				Type    string `json:"type"`
			} `json:"scriptPubKey"`
		} `json:"vout"`
		Hex string `json:"hex"`
	} `json:"tx"`
}

type GetBlockCountRPC struct {
	Result int32 `json:"result"`
}
