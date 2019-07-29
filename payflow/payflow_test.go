package payflow_test

import (
	"os"
	"testing"

	"paypal/payflow"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	username, password, partner, vendor := fetchEnvVars()
	client = payflow.NewClient(username, password, partner, vendor, true)

	code := m.Run()
	// call flag.Parse() here if TestMain uses flags
	os.Exit(code)
}

var client *payflow.PayPalClient

func TestDoCaptureWithVisa(t *testing.T) {
	sampleVisa := payflow.PayPalCreditCard{
		PAN:     "4111111111111111",
		Amount:  "3.50",
		ExpDate: "1220",
	}

	response, err := client.DoSale(sampleVisa)
	if err != nil {
		t.Errorf("Error while conducting sale. CreditCard information used: %#v | Error: %#v", sampleVisa, err)
	}

	t.Logf("%+v", response)
}

func TestDoCaptureMissingPAN(t *testing.T) {
	sampleVisa := payflow.PayPalCreditCard{
		PAN:     "",
		Amount:  "3.50",
		ExpDate: "1220",
	}

	response, err := client.DoSale(sampleVisa)
	assert.Error(t, err, "Payflow API Called failed. Response Code: 23 Response Message: Invalid account number")
	t.Logf("%+v", response)
}

func TestDoCaptureMissingAmount(t *testing.T) {
	sampleVisa := payflow.PayPalCreditCard{
		PAN:     "4111111111111111",
		Amount:  "",
		ExpDate: "1220",
	}

	response, err := client.DoSale(sampleVisa)
	assert.Error(t, err, "Payflow API Called failed. Response Code: 4 Response Message: Invalid amount")
	t.Logf("%+v", response)
}

func TestDoCaptureMissingExpDate(t *testing.T) {
	sampleVisa := payflow.PayPalCreditCard{
		PAN:     "4111111111111111",
		Amount:  "3.50",
		ExpDate: "",
	}

	response, err := client.DoSale(sampleVisa)
	assert.Error(t, err, "Payflow API Called failed. Response Code: 4 Response Message: Invalid amount")
	t.Logf("%+v", response)
}

func fetchEnvVars() (username, password, partner, vendor string) {
	username = os.Getenv("PAYFLOW_TEST_USERNAME")
	if len(username) <= 0 {
		panic("Test cannot run because cannot get environment variable PAYPAL_TEST_USERNAME")
	}
	password = os.Getenv("PAYFLOW_TEST_PASSWORD")
	if len(password) <= 0 {
		panic("Test cannot run because cannot get environment variable PAYPAL_TEST_PASSWORD")
	}
	vendor = os.Getenv("PAYFLOW_TEST_VENDOR")
	if len(vendor) <= 0 {
		panic("Test cannot run because cannot get environment variable PAYPAL_TEST_VENDOR")
	}
	partner = os.Getenv("PAYFLOW_TEST_PARTNER")
	if len(partner) <= 0 {
		panic("Test cannot run because cannot get environment variable PAYPAL_TEST_PARTNER")
	}
	return
}
