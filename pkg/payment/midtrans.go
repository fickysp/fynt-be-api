package payment

import (
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type midtransClient struct {
	client snap.Client
}

func NewMidtransClient(serverKey string, env string) *midtransClient {
	var cl snap.Client

	enviro := midtrans.Sandbox
	if env == "production" {
		enviro = midtrans.Production
	}

	cl.New(serverKey, enviro)
	return &midtransClient{
		client: cl,
	}
}

func (m *midtransClient) CreatePaymentURL(transactionID string, amount int64, customerName string) (string, string, error) {
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  transactionID,
			GrossAmt: amount,
		}, CustomerDetail: &midtrans.CustomerDetails{
			FName: customerName,
		},
	}

	resp, err := m.client.CreateTransaction(req)
	if err != nil {
		return "", "", err
	}

	return resp.Token, resp.RedirectURL, nil
}
