package paypal_test

import (
  "../paypal"
	"testing"
	"os"
)

const (
  TEST_RETURN_URL = "http://localhost/RETURN-URL"
  TEST_CANCEL_URL = "http://localhost/CANCEL-URL"
)

func TestSandboxRedirect(t *testing.T) {
  currencyCode :=  "USD"
  
  username := os.Getenv("PAYPAL_TEST_USERNAME")
	if len(username) <= 0 {
		t.Fatalf("Test cannot run because cannot get environment variable PAYPAL_TEST_USERNAME")
	}
  password := os.Getenv("PAYPAL_TEST_PASSWORD")
	if len(password) <= 0 {
		t.Fatalf("Test cannot run because cannot get environment variable PAYPAL_TEST_PASSWORD")
	}
  signature := os.Getenv("PAYPAL_TEST_SIGNATURE")
	if len(signature) <= 0 {
		t.Fatalf("Test cannot run because cannot get environment variable PAYPAL_TEST_SIGNATURE")
	}
  
  client := paypal.NewDefaultClient(username, password, signature, true)  
	
	// Make an array of your digital-goods
	testGoods := []paypal.PayPalDigitalGood{paypal.PayPalDigitalGood{
    Name: "Test Good", 
    Amount: 200.000,
    Quantity: 5,
  }}
  
	response, err := client.SetExpressCheckoutDigitalGoods(paypal.SumPayPalDigitalGoodAmounts(&testGoods), 
    currencyCode, 
    TEST_RETURN_URL, 
    TEST_CANCEL_URL, 
    testGoods,
  )
  
  if err != nil {
    t.Errorf("Error returned in SetExpressCheckoutDigitalGoods: %#v.", err)
  }
	
  if len(response.Values["TOKEN"][0]) <= 0 {
    t.Errorf("Didn't get token back from PayPal. Response was: %#v", response.Values)
  }
	
  if response.Values["ACK"][0] != "Success" {
    t.Errorf("Didn't get ACK=Success back from PayPal. Response was: %#v", response.Values)
  }
}