package bitcoin

import (
	"bytes"
	"log"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type BitcoinRepository struct {
	url      string
	username string
	password string
}

func NewBitcoinRepo(url, username, password string) *BitcoinRepository {
	return &BitcoinRepository{url, username, password}
}

func (b *BitcoinRepository) authorization() string {
	str := b.username + ":" + b.password
	encoded := base64.StdEncoding.EncodeToString([]byte(str))
	return "Basic " + encoded
}

func (b *BitcoinRepository) rpc(method string, params []json.RawMessage) ([]byte, error) {
	request := &Request{
		Jsonrpc: "1.0",
		Method:  method,
		Params:  params,
	}
	requestMarshalled, _ := json.Marshal(request)

	r, err := http.NewRequest("POST", b.url, bytes.NewBuffer(requestMarshalled))
	if err != nil {
		return nil, err
	}

	if b.username != "" && b.password != "" {
		r.Header.Add("Authorization", b.authorization())
	}
	r.Header.Add("Content-Type", "application/json")

	c := &http.Client{}
	res, err := c.Do(r)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		log.Println(string(body))
		return nil, fmt.Errorf("errors.request status code: %d", res.StatusCode)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (b *BitcoinRepository) GetBlockHash(height int32) (string, error) {
	params, _ := json.Marshal(height)
	paramsJson := []json.RawMessage{json.RawMessage(params)}
	resBytes, err := b.rpc("getblockhash", paramsJson)
	if err != nil {
		return "", err
	}

	getBlockHash := GetBlockHashRPC{}
	_ = json.Unmarshal(resBytes, &getBlockHash)

	return getBlockHash.Result, nil
}

func (b *BitcoinRepository) GetBlock(blockHash string) (*GetBlock, error) {
	blockHashParam, _ := json.Marshal(blockHash)
	verbosity, _ := json.Marshal(3)
	paramsJson := []json.RawMessage{json.RawMessage(blockHashParam), json.RawMessage(verbosity)}

	resBytes, err := b.rpc("getblock", paramsJson)
	if err != nil {
		return nil, err
	}

	getBlock := GetBlockRPC{}
	_ = json.Unmarshal(resBytes, &getBlock)

	return &getBlock.Result, nil
}

func (b *BitcoinRepository) GetBlockCount() (int32, error) {
	paramsJson := []json.RawMessage{}

	resBytes, err := b.rpc("getblockcount", paramsJson)
	if err != nil {
		return 0, err
	}

	getBlockCount := GetBlockCountRPC{}
	_ = json.Unmarshal(resBytes, &getBlockCount)

	return getBlockCount.Result, nil
}

func (b *BitcoinRepository) ForwardRPC(method string, params []interface{}) (json.RawMessage, error) {
	// Convert params to json.RawMessage
	paramsJson := make([]json.RawMessage, len(params))
	for i, param := range params {
		jsonParam, err := json.Marshal(param)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal parameter: %v", err)
		}
		paramsJson[i] = jsonParam
	}

	// Call the existing rpc method
	resBytes, err := b.rpc(method, paramsJson)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var response struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Check for RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("RPC error (code %d): %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}
