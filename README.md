paypal
======

`paypal` is a Go package that allows you to access PayPal APIs, with optional Google AppEngine support, using the  ["PayPal NVP"](https://cms.paypal.com/us/cgi-bin/?cmd=_render-content&content_ID=developer/e_howto_api_nvp_NVPAPIOverview#id09C2F0G0C7U) format.

Included is a method for using the [Digital Goods for Express Checkout](https://cms.paypal.com/us/cgi-bin/?cmd=_render-content&content_ID=developer/e_howto_api_IntegratingExpressCheckoutDG) payment option.

This fork lets you choose to use [AppEngine's urlfetch package](https://developers.google.com/appengine/docs/go/urlfetch/overview) to create the HTTP Client

Quick Start
---

####### Standard Go Usage
```go
import (
  "fmt"
  "github.com/crowdmob/paypal"
  "appengine"
  "appengine/urlfetch"
)

func paypalExpressCheckoutHandler(w http.ResponseWriter, r *http.Request) {
  // An example to setup paypal express checkout for digital goods
  currencyCode := "USD"
  isSandbox    := true
  returnURL    := "http://example.com/returnURL"
  cancelURL    := "http://example.com/cancelURL"
  
  // Create the paypal Client with urlfetch
  client := paypal.NewDefaultClient("Your_Uername", "Your_Password", "Your_Signature", isSandbox)
  
  // Make a array of your digital-goods
  testGoods := []paypal.PayPalDigitalGood{paypal.PayPalDigitalGood{
    Name: "Test Good", 
    Amount: 200.000,
    Quantity: 5,
  }}
  
  // Sum amounts and get the token!
  response, err := client.SetExpressCheckoutDigitalGoods(paypal.SumPayPalDigitalGoodAmounts(&testGoods), 
    currencyCode, 
    returnURL, 
    cancelURL, 
    testGoods,
  )
  
  if err != nil {
    // ... gracefully handle error
  } else { // redirect to paypal
    http.Redirect(w, r, fmt.Sprintf("http://...%s", response.Values["TOKEN"][0]), 301)
  }
}
```

####### App Engine Usage
```go
import (
	"fmt"
	"github.com/crowdmob/paypal"
	"appengine"
	"appengine/urlfetch"
)

func paypalExpressCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	// An example to setup paypal express checkout for digital goods
	currencyCode := "USD"
	isSandbox    := true
	returnURL    := "http://example.com/returnURL"
	cancelURL    := "http://example.com/cancelURL"
	
	// Create the paypal Client with urlfetch
	client := paypal.NewClient("Your_Uername", "Your_Password", "Your_Signature", urlfetch.Client(appengine.NewContext(r)), isSandbox)

  // Make a array of your digital-goods
  testGoods := []paypal.PayPalDigitalGood{paypal.PayPalDigitalGood{
    Name: "Test Good", 
    Amount: 200.000,
    Quantity: 5,
  }}
  
  // Sum amounts and get the token!
  response, err := client.SetExpressCheckoutDigitalGoods(paypal.SumPayPalDigitalGoodAmounts(&testGoods), 
    currencyCode, 
    returnURL, 
    cancelURL, 
    testGoods,
  )
  
  if err != nil {
  // ... gracefully handle error
  } else { // redirect to paypal
    http.Redirect(w, r, fmt.Sprintf("http://...%s", response.Values["TOKEN"][0]), 301)
  }
}
```

Running Tests
---
There's a test suite included.  To run it, simply run:

    go test paypal_test.go

You'll have to have set the following environment variables to run the tests:

    export PAYPAL_TEST_USERNAME=XXX
    export PAYPAL_TEST_PASSWORD=XXX
    export PAYPAL_TEST_SIGNATURE=XXX

Tests currently run in sandbox.


