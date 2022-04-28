package test

type TestTx struct {
	Timestamp1  int64	`json:"timestamp_1"`
	Timestamp2	int64	`json:"timestamp_2"`
	Timestamp3	int64	`json:"timestamp_3"`
}

type TXRequestData struct {
	Timestamp   int64	`json:"timestamp"`
	FromAddress string	`json:"fromAddress"`
	ToAddress	string	`json:"toAddress"`

	ReqInfo 	string	`json:"reqInfo"`
	Token 		string	`json:"token"`

	ReqHash		string	`json:"reqHash"`
}

type TXResponseData struct {
	Timestamp   int64	`json:"timestamp"`
	FromAddress string	`json:"fromAddress"`
	ToAddress	string	`json:"toAddress"`

	ExecResult 	bool	`json:"execResult"`
	InfoURL		string	`json:"infoUrl"`
	Code 		string	`json:"code"`

	ReqHash		string	`json:"reqHash"`
}