package paypal_test

import (
  "../paypal"
	"testing"
	"os"
)

func TestSandboxRedirect(t *testing.T) {
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
  _ = client 
  
  t.Errorf("This test hasn't been fleshed out yet.")
}