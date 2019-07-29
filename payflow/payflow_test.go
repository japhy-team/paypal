package payflow_test

import (
	"os"
	"testing"

	"payflow"
)

func fetchEnvVars(t *testing.T) (username, password, signature string) {
	username = os.Getenv("PAYFLOW_TEST_USERNAME")
	if len(username) <= 0 {
		t.Fatalf("Test cannot run because cannot get environment variable PAYPAL_TEST_USERNAME")
	}
	password = os.Getenv("PAYFLOW_TEST_PASSWORD")
	if len(password) <= 0 {
		t.Fatalf("Test cannot run because cannot get environment variable PAYPAL_TEST_PASSWORD")
	}
	signature = os.Getenv("PAYFLOW_TEST_SIGNATURE")
	if len(signature) <= 0 {
		t.Fatalf("Test cannot run because cannot get environment variable PAYPAL_TEST_SIGNATURE")
	}
	return
}

func TestDoCapture(t *testing.T) {
	username, password, signature := fetchEnvVars(t)
	client := payflow.NewClient()
}
