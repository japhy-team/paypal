paypal
======

`paypal` is a Go package that allows you to access PayPal APIs, with optional Google AppEngine support, using the  ["PayPal NVP"](https://cms.paypal.com/us/cgi-bin/?cmd=_render-content&content_ID=developer/e_howto_api_nvp_NVPAPIOverview#id09C2F0G0C7U) format.

Included is a method for using the [Digital Goods for Express Checkout](https://cms.paypal.com/us/cgi-bin/?cmd=_render-content&content_ID=developer/e_howto_api_IntegratingExpressCheckoutDG) payment option.

This fork lets you choose to use [AppEngine's urlfetch package](https://developers.google.com/appengine/docs/go/urlfetch/overview) to create the HTTP Client

Quick Start
---
	import (
		"fmt"
		"paypal"
		"appengine"
		"appengine/urlfetch"
	)
	
	func paypalExpressCheckoutHandler(w http.ResponseWriter, r *http.Request) {
		// An example to setup paypal express checkout for digital goods
	
		// Create an appengine context for this request
		appengineContext := appengine.NewContext(r)
		
		// Create a urelfetch based HTTP Client (*http.Client) for sending PayPal a request
		httpClient := urlfetch.Client(appengineContext)
		
		isSandbox := true
		
		// Create the paypal Client
		client := paypal.NewClient("Your_Uername", "Your_Password", "Your_Signature", httpClient, isSandbox)
		
		// Make an array of your digital-goods
		goods := make([]paypal.PayPalDigitalGood, 1)
		good := new(paypal.PayPalDigitalGood)
		good.Name, good.Amount, good.Quantity = "Test Good", paymentAmount, 1
		goods[0] = *good
		
		// Setup your checkout options
		paymentAmount := 200
		currencyCode := "USD"
		returnURL := "http://example.com/returnURL"
		cancelURL := "http://example.com/cancelURL"
		
		// Setup Express checkout
		response, err := client.SetExpressCheckoutDigitalGoods(paymentAmount, currencyCode, returnURL, cancelURL, goods)
		
		// Print token, etc from paypal
		fmt.Fprint(w, response.Values)
		
	}
