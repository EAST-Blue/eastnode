package jsonrpc

import (
	"eastnode/indexer/repository/bitcoin"
	"encoding/json"
	"net/http"
)

type BitcoinServer struct {
	BitcoinRepo *bitcoin.BitcoinRepository
}

func (s *BitcoinServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request struct {
		JsonRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      interface{}     `json:"id"`
	}

	// Decode the incoming JSON-RPC request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Unmarshal params
	var params []interface{}
	if err := json.Unmarshal(request.Params, &params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Forward the RPC call to the Bitcoin node
	jsonRes, err := s.BitcoinRepo.ForwardRPC(request.Method, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := struct {
		JsonRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}{
		JsonRPC: "2.0",
		Result:  json.RawMessage(jsonRes),
		ID:      request.ID,
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
