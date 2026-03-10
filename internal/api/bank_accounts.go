package api

import "strings"

type BankAccount struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	AccountHolderName  string   `json:"accountHolderName"`
	SupportedRails     []string `json:"supportedRails"`
	Label              string   `json:"label"`
	CreatedAt          string   `json:"createdAt"`
	Type               string   `json:"type"`
	Currency           string   `json:"currency"`
	AccountNumberLast4 string   `json:"accountNumberLast4"`
	RoutingNumberLast4 string   `json:"routingNumberLast4"`
}

func (a *BankAccount) SupportedRailsStr() string {
	return strings.Join(a.SupportedRails, ";")
}

func (c *Client) ListBankAccounts() ([]BankAccount, error) {
	resp, err := c.do("GET", "/v1/bank-accounts/", nil)
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]BankAccount](resp)
}

func (c *Client) CreateBankAccount(body map[string]interface{}) (*BankAccount, error) {
	resp, err := c.do("POST", "/v1/bank-accounts/", body)
	if err != nil {
		return nil, err
	}
	return decodeJSONPtr[BankAccount](resp)
}

func (c *Client) DeleteBankAccount(accountID string) error {
	resp, err := c.do("DELETE", "/v1/bank-accounts/"+accountID, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
