package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type OperationResponse int

const (
	EtchingOperationResponse OperationResponse = iota
	MintOperationResponse
	BurnOperationResponse
	SendOperationResponse
	ReceiveOperationResponse
)

func (o OperationResponse) String() string {
	return [...]string{"etching", "mint", "burn", "send", "receive"}[o]
}

type RuneActivitiesResponse struct {
	Limit   int `json:"limit"`
	Offset  int `json:"offset"`
	Total   int `json:"total"`
	Results []struct {
		Rune struct {
			ID         string `json:"id"`
			Number     int    `json:"number"`
			Name       string `json:"name"`
			SpacedName string `json:"spaced_name"`
		} `json:"rune"`
		Operation string `json:"operation"`
		Location  struct {
			BlockHash   string `json:"block_hash"`
			BlockHeight int    `json:"block_height"`
			TxID        string `json:"tx_id"`
			TxIndex     int    `json:"tx_index"`
			Timestamp   int    `json:"timestamp"`
			Vout        int    `json:"vout"`
			Output      string `json:"output"`
		} `json:"location"`
		Address         string `json:"address"`
		ReceiverAddress string `json:"receiver_address"`
		Amount          string `json:"amount"`
	} `json:"results"`
}

type AddressBalancesResponse struct {
	Limit   int `json:"limit"`
	Offset  int `json:"offset"`
	Total   int `json:"total"`
	Results []struct {
		Rune struct {
			ID         string `json:"id"`
			Number     int    `json:"number"`
			Name       string `json:"name"`
			SpacedName string `json:"spaced_name"`
		} `json:"rune"`
		Balance string `json:"balance"`
		Address string `json:"address"`
	} `json:"results"`
}

type EtchingResponse struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	SpacedName   string                   `json:"spaced_name"`
	Number       int                      `json:"number"`
	Divisibility int                      `json:"divisibility"`
	Symbol       string                   `json:"symbol"`
	Turbo        bool                     `json:"turbo"`
	MintTerms    EtchingMintTermsResponse `json:"mint_terms"`
	Supply       EtchingSupplyResponse    `json:"supply"`
}

type EtchingMintTermsResponse struct {
	Amount      *string `json:"amount,omitempty"`
	Cap         *string `json:"cap,omitempty"`
	HeightStart *int    `json:"height_start"`
	HeightEnd   *int    `json:"height_end"`
	OffsetStart *int    `json:"offset_start"`
	OffsetEnd   *int    `json:"offset_end"`
}

type EtchingSupplyResponse struct {
	Current        string `json:"current"`
	Minted         string `json:"minted"`
	TotalMints     string `json:"total_mints"`
	MintPercentage string `json:"mint_percentage"`
	Mintable       bool   `json:"mintable"`
	Burned         string `json:"burned"`
	TotalBurns     string `json:"total_burns"`
	Premine        string `json:"premine"`
}

type Operation int

const (
	EtchingOperation Operation = iota
	MintOperation
	BurnOperation
	EdictOperation
	ReceiveOperation
)

func (o Operation) String() string {
	return [...]string{"etching", "mint", "burn", "edict", "receive"}[o]
}

type RuneActivity struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Op          string  `json:"op"`
	Txid        string  `json:"txid"`
	Vout        *int    `json:"vout"`
	FromAddress *string `json:"from_address"`
	ToAddress   *string `json:"to_address"`
	Amount      *string `json:"amount"`
}

type AddressBalance struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Balance string `json:"balance"`
}

type Etching struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Supply          string  `json:"supply"`
	TotalMints      string  `json:"total_mints"`
	TotalBurns      string  `json:"total_burns"`
	Divisibility    int     `json:"divisibility"`
	Premine         string  `json:"premine"`
	Symbol          string  `json:"symbol"`
	Turbo           bool    `json:"turbo"`
	TermAmount      *string `json:"term_amount"`
	TermCap         *string `json:"term_cap"`
	TermHeightStart *int    `json:"term_height_start"`
	TermHeightEnd   *int    `json:"term_height_end"`
	TermOffsetStart *int    `json:"term_offset_start"`
	TermOffsetEnd   *int    `json:"term_offset_end"`
}

type RuneExternal struct{}

func NewRuneExternal() *RuneExternal {
	return &RuneExternal{}
}

func (r *RuneExternal) Get(url string) ([]byte, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, os.Getenv("RUNE_EXTERNAL_API")+url, nil)

	if err != nil {
		return []byte{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

func (r *RuneExternal) GetEtchingById(id string) (*Etching, error) {
	etchingResponseRaw, err := r.Get(fmt.Sprintf("/runes/v1/etchings/%s", id))
	if err != nil {
		return nil, err
	}

	etchingResponse := &EtchingResponse{}
	err = json.Unmarshal(etchingResponseRaw, etchingResponse)
	if err != nil {
		return nil, err
	}

	etching := &Etching{
		ID:   etchingResponse.ID,
		Name: etchingResponse.Name,

		Supply:     etchingResponse.Supply.Current,
		TotalMints: etchingResponse.Supply.TotalMints,
		TotalBurns: etchingResponse.Supply.TotalBurns,

		Divisibility: etchingResponse.Divisibility,
		Premine:      etchingResponse.Supply.Premine,
		Symbol:       etchingResponse.Symbol,
		Turbo:        etchingResponse.Turbo,

		TermAmount:      etchingResponse.MintTerms.Amount,
		TermCap:         etchingResponse.MintTerms.Cap,
		TermHeightStart: etchingResponse.MintTerms.HeightStart,
		TermHeightEnd:   etchingResponse.MintTerms.HeightEnd,
		TermOffsetStart: etchingResponse.MintTerms.OffsetStart,
		TermOffsetEnd:   etchingResponse.MintTerms.OffsetEnd,
	}

	return etching, nil
}

func (r *RuneExternal) GetAddressBalancesByAddress(address string) (balances []AddressBalance, err error) {
	addressBalancesResponseRaw, err := r.Get(fmt.Sprintf("/runes/v1/addresses/%s/balances?limit=9007199254740991", address))
	if err != nil {
		return balances, err
	}

	addressBalancesResponse := &AddressBalancesResponse{}
	err = json.Unmarshal(addressBalancesResponseRaw, addressBalancesResponse)
	if err != nil {
		return balances, err
	}

	for _, balance := range addressBalancesResponse.Results {
		addressBalance := AddressBalance{
			ID:      balance.Rune.ID,
			Name:    balance.Rune.SpacedName,
			Balance: balance.Balance,
		}
		balances = append(balances, addressBalance)
	}

	return balances, nil
}

func (r *RuneExternal) GetActivitiesByAddress(address string) (acttivites []RuneActivity, err error) {
	activites := []RuneActivity{}
	activitiesResponseRaw, err := r.Get(fmt.Sprintf("/runes/v1/addresses/%s/activity?limit=9007199254740991", address))
	if err != nil {
		return activites, err
	}

	return r.ParseGetActivities(activitiesResponseRaw)
}

func (r *RuneExternal) GetActivitiesByTxid(txid string) (acttivites []RuneActivity, err error) {
	activites := []RuneActivity{}
	activitiesResponseRaw, err := r.Get(fmt.Sprintf("/runes/v1/transactions/%s/activity?limit=9007199254740991", txid))
	if err != nil {
		return activites, err
	}

	return r.ParseGetActivities(activitiesResponseRaw)
}

func (r *RuneExternal) ParseGetActivities(raw []byte) (acttivites []RuneActivity, err error) {
	activites := []RuneActivity{}
	activitesResponse := &RuneActivitiesResponse{}
	err = json.Unmarshal(raw, activitesResponse)
	if err != nil {
		return activites, err
	}

	for _, activity := range activitesResponse.Results {

		switch activity.Operation {
		case EtchingOperationResponse.String():
			runeActivity := RuneActivity{
				ID:   activity.Rune.ID,
				Name: activity.Rune.SpacedName,
				Op:   EtchingOperation.String(),
				Txid: activity.Location.TxID,
			}
			activites = append(activites, runeActivity)

		case MintOperationResponse.String():
			runeActivity := RuneActivity{
				ID:     activity.Rune.ID,
				Name:   activity.Rune.SpacedName,
				Op:     MintOperation.String(),
				Txid:   activity.Location.TxID,
				Amount: &activity.Amount,
			}
			activites = append(activites, runeActivity)

		case BurnOperationResponse.String():
			runeActivity := RuneActivity{
				ID:          activity.Rune.ID,
				Name:        activity.Rune.SpacedName,
				Op:          BurnOperation.String(),
				Txid:        activity.Location.TxID,
				Amount:      &activity.Amount,
				FromAddress: &activity.Address,
			}
			activites = append(activites, runeActivity)

		case SendOperationResponse.String():
			runeActivity := RuneActivity{
				ID:          activity.Rune.ID,
				Name:        activity.Rune.SpacedName,
				Op:          EdictOperation.String(),
				Txid:        activity.Location.TxID,
				Vout:        &activity.Location.Vout,
				FromAddress: &activity.Address,
				ToAddress:   &activity.ReceiverAddress,
				Amount:      &activity.Amount,
			}
			activites = append(activites, runeActivity)

		case ReceiveOperationResponse.String():
			runeActivity := RuneActivity{
				ID:        activity.Rune.ID,
				Name:      activity.Rune.SpacedName,
				Op:        ReceiveOperation.String(),
				Txid:      activity.Location.TxID,
				Vout:      &activity.Location.Vout,
				ToAddress: &activity.Address,
				Amount:    &activity.Amount,
			}
			activites = append(activites, runeActivity)

		default:
			panic("Invalid operation")
		}
	}

	return activites, err
}
