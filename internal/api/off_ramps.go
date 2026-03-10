package api

import (
	"fmt"
	"net/url"
)

type OffRamp struct {
	ID        string        `json:"id"`
	Status    string        `json:"status"`
	CreatedAt string        `json:"createdAt"`
	Chain     string        `json:"chain"`
	AccountID string        `json:"accountId"`
	Payout    AmountField   `json:"payout"`
	Fees      AmountField   `json:"fees"`
	Payment   PaymentAsset  `json:"paymentAsset"`
	Tx        OffRampTx     `json:"transaction"`
}

type AmountField struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type PaymentAsset struct {
	Amount string `json:"amount"`
	Token  string `json:"token"`
}

type OffRampTx struct {
	Hash        *string `json:"hash"`
	ExplorerURL *string `json:"explorerUrl"`
}

type OffRampList struct {
	Data       []OffRamp `json:"data"`
	HasMore    bool      `json:"hasMore"`
	NextCursor string    `json:"nextCursor,omitempty"`
}

type ListOffRampsParams struct {
	Limit     int
	Cursor    string
	Status    string
	Chain     string
	AccountID string
	Sort      string
}

func (p *ListOffRampsParams) queryString() string {
	q := url.Values{}
	if p.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", p.Limit))
	}
	if p.Cursor != "" {
		q.Set("cursor", p.Cursor)
	}
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	if p.Chain != "" {
		q.Set("chain", p.Chain)
	}
	if p.AccountID != "" {
		q.Set("accountId", p.AccountID)
	}
	if p.Sort != "" {
		q.Set("sort", p.Sort)
	}
	if encoded := q.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

func (c *Client) ListOffRamps(params *ListOffRampsParams) (*OffRampList, error) {
	path := "/v1/off-ramps/" + params.queryString()
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return decodeJSONPtr[OffRampList](resp)
}
