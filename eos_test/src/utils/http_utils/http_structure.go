package http_utils

// 通用合约
type Contract struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Address   string      `json:"address"`
	Timestamp int64       `json:"timestamp"`
	Signature string      `json:"signature"`
	Pubkey    string      `json:"pubkey"`
}


// 转账结构
type TransferRequest struct {
	SignedContract TransferContract `json:"signedContract"`
}

type TransferContract struct {
	ID        string                  `json:"id"`
	Type      string                  `json:"type"`
	Payload   TransferContractPayload `json:"payload"`
	Address   string                  `json:"address"`
	Timestamp int64                   `json:"timestamp"`
	Signature string                  `json:"signature"`
	Pubkey    string                  `json:"pubkey"`
}

type TransferContractPayload struct {
	AccountFrom string  `json:"accountFrom"`
	AccountTo   string  `json:"accountTo"`
	CoinNum     float64 `json:"coinNum"`
}

type ContractConfirmationResponseData struct {
	ConfirmedContractID string `json:"confirmedContractID"`
}

type ContractConfirmationResponse struct {
	Code int                              `json:"code"`
	Data ContractConfirmationResponseData `json:"data"`
	Msg  string                           `json:"msg"`
}

