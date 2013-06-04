go-appengine-paypal
---

go-appengine-paypal is a Go package for Google AppEngine that allows you to access PayPal APIs using the ["NVP"](https://cms.paypal.com/us/cgi-bin/?cmd=_render-content&content_ID=developer/e_howto_api_nvp_NVPAPIOverview#id09C2F0G0C7U) format.

Included is a method for using the [Digital Goods for Express Checkout](https://cms.paypal.com/us/cgi-bin/?cmd=_render-content&content_ID=developer/e_howto_api_IntegratingExpressCheckoutDG) payment option.

This fork uses [AppEngine's urlfetch package](https://developers.google.com/appengine/docs/go/urlfetch/overview) to create the HTTP Client

Quick Start
---
	import (
		"fmt"
		"paypal"
		"appengine"
		"appengine/urlfetch"
	)
	
	func paypalExpressCheckoutHandler(w http.ResponseWriter, r *http.Request) {
		// An example to start a paypal express checkout for digital goods
	
		// Create an appengine context for this request
		appengineContext := appengine.NewContext(r)
		
		// Create a urelfetch based HTTP Client (*http.Client) for sending PayPal a request
		httpClient := urlfetch.Client(appengineContext)
		
		isSandbox := true
		
		// Create the paypal Client
		client := paypal.NewClient("Your_Uername", "Your_Password", "Your_Signature", httpClient, isSandbox)
		
		// Setup your checkout options
		paymentAmount := 200
		currencyCode := "USD"
		returnURL := "http://example.com/returnURL"
		cancelURL := "http://example.com/cancelURL"
		
		// Make an array of your digital-goods
		goods := make([]paypal.PayPalDigitalGood, 1)
		good := new(paypal.PayPalDigitalGood)
		good.Name, good.Amount, good.Quantity = "Test Good", paymentAmount, 1
		goods[0] = *good
		
		// Start Express checkout
		response, err := client.SetExpressCheckoutDigitalGoods(paymentAmount, currencyCode, returnURL, cancelURL, goods)
		
		// Print result from paypal
		fmt.Fprint(w, response.Values)
	}
