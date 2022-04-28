package function

import (
	"bfc/src/test"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"bfc/src/block"
	"bfc/src/sharecc"
	"bfc/src/utils"
)

func RespondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}

	// utils.Log.Infof("Response: %s", string(response))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func createTx(contract []byte, txType int) *block.Transaction {
	bTxHash, _ := utils.GetHashFromBytes(contract)
	txHash := hex.EncodeToString(bTxHash)

	tx := block.Transaction{
		TransactionID:   txHash,
		TimeStamp:       time.Now().UTC().UnixNano(),
		TransactionType: txType,
		TransactionBody: nil,
	}
	switch txType {
	case block.TX_JOIN:
		var txBody sharecc.JoinRequest
		_ = json.Unmarshal(contract, &txBody)
		tx.TransactionBody = txBody
	case block.TX_TEST:
		var txBody test.TestTx
		_ = json.Unmarshal(contract, &txBody)
		tx.TransactionBody = txBody
	case block.TX_REQUEST:
		var txBody test.TXRequestData
		_ = json.Unmarshal(contract, &txBody)
		tx.TransactionBody = txBody
	case block.TX_RESPONSE:
		var txBody test.TXResponseData
		_ = json.Unmarshal(contract, &txBody)
		tx.TransactionBody = txBody
	}


	return &tx
}