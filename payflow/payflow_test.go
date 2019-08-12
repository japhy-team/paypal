package payflow_test

import (
	"log"
	"os"
	"testing"

	"paypal/payflow"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// PayPal Credit Card Numbers for Testing
// For the PayPal processor, use the following credit card numbers for testing. Any other card number produces a general failure. The credit card numbers are applicable only for Payflow users. Express Checkout users cannot use these test numbers.

const AmericanExpress1 = "378282246310005"
const AmericanExpress2 = "371449635398431"
const AmericanExpressCorporate = "378734493671000"
const DinersClub = "30569309025904"
const Discover1 = "6011111111111117"
const Discover2 = "6011000990139424"
const JCB1 = "3530111333300000"
const JCB2 = "3566002020360505"
const MasterCard1 = "5555555555554444"
const MasterCard2 = "5105105105105100"
const Visa1 = "4111111111111111"
const Visa2 = "4012888888881881"
const Visa3 = "4222222222222"

// Oh man I wonder whats gonna go in here, no way its gonna be credit cards that paypal supports for testing. No way its that.
var CreditCardsPayPalSupports map[string]string

// Processors Other Than PayPal
// For processors other than the PayPal processor, use the guidelines below.

// Credit Card Numbers for Testing
// For processors other than PayPal, use the following credit card numbers for testing. For PayPal test credit card numbers see PayPal credit card numbers for testing.

// American Express	378282246310005
// American Express	371449635398431
// American Express Corporate	378734493671000
// Diners Club	30569309025904
// Discover	6011111111111117
// Discover	6011000990139424
// JCB	3530111333300000
// JCB	3566002020360505
// MasterCard	2221000000000009
// MasterCard	2223000048400011
// MasterCard	2223016768739313
// MasterCard	5555555555554444
// MasterCard	5105105105105100
// Visa	4111111111111111
// Visa	4012888888881881
// Visa	4222222222222

func TestMain(m *testing.M) {
	// ListOfCreditCardsPayPalSupports = []string{AmericanExpress1, AmericanExpress2, AmericanExpressCorporate, DinersClub, Discover1, Discover2, JCB1, JCB2, MasterCard1, MasterCard2, Visa1, Visa2, Visa3}

	username, password, partner, vendor := fetchEnvVars()
	client = payflow.NewClient(username, password, partner, vendor, true)

	code := m.Run()
	// call flag.Parse() here if TestMain uses flags
	os.Exit(code)
}

var client *payflow.PayPalClient

func TestDoSaleWithVisa(t *testing.T) {
	sampleVisa := payflow.PayPalCreditCard{
		PAN:     Visa1,
		Amount:  "3.50",
		ExpDate: "1220",
	}

	response, err := client.DoSale(sampleVisa)
	if err != nil {
		t.Errorf("Error while conducting sale. CreditCard information used: %#v | Error: %#v", sampleVisa, err)
	}

	t.Logf("%+v", response)
}

func TestDoSaleWithAllPayPalCreditCards(t *testing.T) {
	CreditCardsPayPalSupports := map[string]string{
		"AmericanExpress1":         AmericanExpress1,
		"AmericanExpress2":         AmericanExpress2,
		"AmericanExpressCorporate": AmericanExpressCorporate,
		"DinersClub":               DinersClub,
		"Discover1":                Discover1,
		"Discover2":                Discover2,
		"JCB1":                     JCB1,
		"JCB2":                     JCB2,
		"MasterCard1":              MasterCard1,
		"MasterCard2":              MasterCard2,
		"Visa1":                    Visa1,
		"Visa2":                    Visa2,
		"Visa3":                    Visa3,
	}
	t.Log("CreditCardsPayPalSupports: ", CreditCardsPayPalSupports)
	for key, value := range CreditCardsPayPalSupports {
		t.Log(key, value)
		cc := payflow.PayPalCreditCard{
			PAN:     value,
			Amount:  "3.50",
			ExpDate: "1220",
		}

		response, err := client.DoSale(cc)
		if err != nil {
			t.Errorf("Error while conducting sale. %s CreditCard information used: %#v | Error: %#v", key, cc, err)
		}

		t.Logf("%+v", response)
	}

}

func TestDoSaleMissingPAN(t *testing.T) {
	sampleVisa := payflow.PayPalCreditCard{
		PAN:     "",
		Amount:  "3.50",
		ExpDate: "1220",
	}

	response, err := client.DoSale(sampleVisa)
	assert.Error(t, err, "Payflow API Called failed. Response Code: 23 Response Message: Invalid account number")
	t.Logf("%+v", response)
}

func TestDoSaleMissingAmount(t *testing.T) {
	sampleVisa := payflow.PayPalCreditCard{
		PAN:     "4111111111111111",
		Amount:  "",
		ExpDate: "1220",
	}

	response, err := client.DoSale(sampleVisa)
	assert.Error(t, err, "Payflow API Called failed. Response Code: 4 Response Message: Invalid amount")
	t.Logf("%+v", response)
}

func TestDoSaleMissingExpDate(t *testing.T) {
	sampleVisa := payflow.PayPalCreditCard{
		PAN:     "4111111111111111",
		Amount:  "3.50",
		ExpDate: "",
	}

	response, err := client.DoSale(sampleVisa)
	assert.Error(t, err, "Payflow API Called failed. Response Code: 4 Response Message: Invalid amount")
	t.Logf("%+v", response)
}

func TestDoAuthorizeWithVisa(t *testing.T) {
	sampleVisa := payflow.PayPalCreditCard{
		PAN:     Visa1,
		Amount:  "3.50",
		ExpDate: "1220",
	}

	response, err := client.DoAuth(sampleVisa, false)
	if err != nil {
		t.Errorf("Error while conducting authorize. CreditCard information used: %#v | Error: %#v", sampleVisa, err)
	}
	t.Logf("%+v", response)
}

func TestDoAuthorizeMissingPAN(t *testing.T) {}

func TestDoAuthorizeMissingExpDate(t *testing.T) {}

func TestDoAuthorizeMissingAmount(t *testing.T) {}

func fetchEnvVars() (username, password, partner, vendor string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error occurred while loading .env file", err)
	}

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
