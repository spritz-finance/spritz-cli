package api

import "encoding/json"

type CreateQuoteRequest struct {
	AccountID    string `json:"accountId"`
	Amount       string `json:"amount"`
	Chain        string `json:"chain"`
	AmountMode   string `json:"amountMode,omitempty"`
	TokenAddress string `json:"tokenAddress,omitempty"`
	Rail         string `json:"rail,omitempty"`
	Memo         string `json:"memo,omitempty"`
}

type CreateTransactionRequest struct {
	SenderAddress string `json:"senderAddress,omitempty"`
	FeePayer      string `json:"feePayer,omitempty"`
}

func (c *Client) CreateQuote(req *CreateQuoteRequest) (json.RawMessage, error) {
	resp, err := c.do("POST", "/v1/off-ramp-quotes/", req)
	if err != nil {
		return nil, err
	}
	return decodeJSONRaw(resp)
}

func (c *Client) GetQuote(quoteID string) (json.RawMessage, error) {
	resp, err := c.do("GET", "/v1/off-ramp-quotes/"+quoteID, nil)
	if err != nil {
		return nil, err
	}
	return decodeJSONRaw(resp)
}

func (c *Client) CreateTransaction(quoteID string, req *CreateTransactionRequest) (json.RawMessage, error) {
	resp, err := c.do("POST", "/v1/off-ramp-quotes/"+quoteID+"/transaction", req)
	if err != nil {
		return nil, err
	}
	return decodeJSONRaw(resp)
}
