package paypal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	NVP_SANDBOX_URL         = "https://api-3t.sandbox.paypal.com/nvp"
	NVP_PRODUCTION_URL      = "https://api-3t.paypal.com/nvp"
	CHECKOUT_SANDBOX_URL    = "https://www.sandbox.paypal.com/cgi-bin/webscr"
	CHECKOUT_PRODUCTION_URL = "https://www.paypal.com/cgi-bin/webscr"
	NVP_VERSION             = "204"
)

type PayPalClient struct {
	username    string
	password    string
	signature   string
	endpoint    string
	usesSandbox bool
	client      *http.Client
}

type PayPalDigitalGood struct {
	Name     string
	Amount   float64
	Quantity int16
}

type PayPalResponse struct {
	Ack           string     `json:"Ack"`
	Build         string     `json:"Build"`
	CorrelationID string     `json:"CorrelationId"`
	Timestamp     string     `json:"Timestamp"`
	Version       string     `json:"Version"`
	Values        url.Values `json:"Values"`
	usedSandbox   bool
}

type PayPalValues struct {
	Ack                       string `json:"ack,omitempty"`
	Amount                    string `json:"amt,omitempty"`
	BillingAgreementID        string `json:"billingagreementid,omitempty"`
	Build                     string `json:"build,omitempty"`
	CorrelationID             string `json:"correlationid,omitempty"`
	CurrencyCode              string `json:"currencycode,omitempty"`
	ErrorCode                 string `json:"errorcode0,omitempty"`
	ErrorMessage              string `json:"l_shortmessage0,omitempty"`
	ErrorMessageExtended      string `json:"l_longmessage0,omitempty"`
	DateOrdered               string `json:"ordertime,omitempty"`
	PaymentStatus             string `json:"paymentstatus,omitempty"`
	PaymentType               string `json:"paymenttype,omitempty"`
	PendingReason             string `json:"pendingreason,omitempty"`
	ProtectionEligibility     string `json:"protectioneligiblity,omitempty"`
	ProtectionEligibilityType string `json:"protectioneligibilitytype,omitempty"`
	ReasonCode                string `json:"reasoncode,omitempty"`
	SeverityCode              string `json:"l_severitycode0,omitempty"`
	TaxedAmount               string `json:"taxamt,omitempty"`
	Timestamp                 string `json:"timestamp,omitempty"`
	TransactionID             string `json:"transactionid,omitempty"`
	TransactionType           string `json:"transactiontype,omitempty"`
	Version                   string `json:"version,omitempty"`
}

type PayPalError struct {
	Ack          string
	ErrorCode    string
	ShortMessage string
	LongMessage  string
	SeverityCode string
}

func (e *PayPalError) Error() string {
	var message string
	if len(e.ErrorCode) != 0 && len(e.ShortMessage) != 0 {
		message = "PayPal Error " + e.ErrorCode + ": " + e.ShortMessage
	} else if len(e.Ack) != 0 {
		message = e.Ack
	} else {
		message = "PayPal is undergoing maintenance.\nPlease try again later."
	}

	return message
}

func (r *PayPalResponse) CheckoutUrl() string {
	query := url.Values{}
	query.Set("cmd", "_express-checkout")
	query.Add("token", r.Values["TOKEN"][0])
	checkoutUrl := CHECKOUT_PRODUCTION_URL
	if r.usedSandbox {
		checkoutUrl = CHECKOUT_SANDBOX_URL
	}
	return fmt.Sprintf("%s?%s", checkoutUrl, query.Encode())
}

func SumPayPalDigitalGoodAmounts(goods *[]PayPalDigitalGood) (sum float64) {
	for _, dg := range *goods {
		sum += dg.Amount * float64(dg.Quantity)
	}
	return
}

func NewDefaultClientEndpoint(username, password, signature, endpoint string, usesSandbox bool) *PayPalClient {
	return &PayPalClient{
		username:    username,
		password:    password,
		signature:   signature,
		endpoint:    endpoint,
		usesSandbox: usesSandbox,
		client:      new(http.Client),
	}
}

func NewDefaultClient(username, password, signature string, usesSandbox bool) *PayPalClient {
	var endpoint = NVP_PRODUCTION_URL
	if usesSandbox {
		endpoint = NVP_SANDBOX_URL
	}

	return &PayPalClient{
		username:    username,
		password:    password,
		signature:   signature,
		endpoint:    endpoint,
		usesSandbox: usesSandbox,
		client:      new(http.Client),
	}
}

func NewClient(username, password, signature string, usesSandbox bool, client *http.Client) *PayPalClient {
	var endpoint = NVP_PRODUCTION_URL
	if usesSandbox {
		endpoint = NVP_SANDBOX_URL
	}

	return &PayPalClient{
		username:    username,
		password:    password,
		signature:   signature,
		endpoint:    endpoint,
		usesSandbox: usesSandbox,
		client:      client,
	}
}

func (pClient *PayPalClient) PerformRequest(values url.Values) (*PayPalResponse, error) {
	values.Add("USER", pClient.username)
	values.Add("PWD", pClient.password)
	values.Add("SIGNATURE", pClient.signature)
	values.Add("VERSION", NVP_VERSION)

	formResponse, err := pClient.client.PostForm(pClient.endpoint, values)
	if err != nil {
		return nil, err
	}
	defer formResponse.Body.Close()

	body, err := ioutil.ReadAll(formResponse.Body)
	if err != nil {
		return nil, err
	}

	responseValues, err := url.ParseQuery(string(body))
	response := &PayPalResponse{usedSandbox: pClient.usesSandbox}
	if err == nil {
		response.Ack = responseValues.Get("ACK")
		response.CorrelationID = responseValues.Get("CORRELATIONID")
		response.Timestamp = responseValues.Get("TIMESTAMP")
		response.Version = responseValues.Get("VERSION")
		response.Build = responseValues.Get("2975009")
		response.Values = responseValues

		errorCode := responseValues.Get("L_ERRORCODE0")
		if len(errorCode) != 0 || strings.ToLower(response.Ack) == "failure" || strings.ToLower(response.Ack) == "failurewithwarning" {
			pError := new(PayPalError)
			pError.Ack = response.Ack
			pError.ErrorCode = errorCode
			pError.ShortMessage = responseValues.Get("L_SHORTMESSAGE0")
			pError.LongMessage = responseValues.Get("L_LONGMESSAGE0")
			pError.SeverityCode = responseValues.Get("L_SEVERITYCODE0")

			err = pError
		}
	}

	return response, err
}

func (pClient *PayPalClient) SetExpressCheckoutDigitalGoods(paymentAmount float64, currencyCode string, returnURL, cancelURL string, goods []PayPalDigitalGood) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "SetExpressCheckout")
	values.Add("PAYMENTREQUEST_0_AMT", fmt.Sprintf("%.2f", paymentAmount))
	values.Add("PAYMENTREQUEST_0_PAYMENTACTION", "Sale")
	values.Add("PAYMENTREQUEST_0_CURRENCYCODE", currencyCode)
	values.Add("RETURNURL", returnURL)
	values.Add("CANCELURL", cancelURL)
	values.Add("REQCONFIRMSHIPPING", "0")
	values.Add("NOSHIPPING", "1")
	values.Add("SOLUTIONTYPE", "Sole")

	for i := 0; i < len(goods); i++ {
		good := goods[i]

		values.Add(fmt.Sprintf("%s%d", "L_PAYMENTREQUEST_0_NAME", i), good.Name)
		values.Add(fmt.Sprintf("%s%d", "L_PAYMENTREQUEST_0_AMT", i), fmt.Sprintf("%.2f", good.Amount))
		values.Add(fmt.Sprintf("%s%d", "L_PAYMENTREQUEST_0_QTY", i), fmt.Sprintf("%d", good.Quantity))
		values.Add(fmt.Sprintf("%s%d", "L_PAYMENTREQUEST_0_ITEMCATEGORY", i), "Digital")
	}

	return pClient.PerformRequest(values)
}

// Convenience function for Sale (Charge)
func (pClient *PayPalClient) DoExpressCheckoutSale(token, payerId, currencyCode string, finalPaymentAmount float64) (*PayPalResponse, error) {
	return pClient.DoExpressCheckoutPayment(token, payerId, "Sale", currencyCode, finalPaymentAmount)
}

// paymentType can be "Sale" or "Authorization" or "Order" (ship later)
func (pClient *PayPalClient) DoExpressCheckoutPayment(token, payerId, paymentType, currencyCode string, finalPaymentAmount float64) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "DoExpressCheckoutPayment")
	values.Add("TOKEN", token)
	values.Add("PAYERID", payerId)
	values.Add("PAYMENTREQUEST_0_PAYMENTACTION", paymentType)
	values.Add("PAYMENTREQUEST_0_CURRENCYCODE", currencyCode)
	values.Add("PAYMENTREQUEST_0_AMT", fmt.Sprintf("%.2f", finalPaymentAmount))

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) GetExpressCheckoutDetails(token string) (*PayPalResponse, error) {
	values := url.Values{}
	values.Add("TOKEN", token)
	values.Set("METHOD", "GetExpressCheckoutDetails")
	return pClient.PerformRequest(values)
}

//----------------------------------------------------------
// Forked
//----------------------------------------------------------

func (pClient *PayPalClient) CreateRecurringPaymentsProfile(token string, params map[string]string) (*PayPalResponse, error) {
	values := url.Values{}
	values.Add("TOKEN", token)
	values.Set("METHOD", "CreateRecurringPaymentsProfile")

	if params != nil {
		for key, value := range params {
			values.Add(key, value)
		}
	}

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) BillOutstandingAmount(profileId string) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "BillOutstandingAmount")
	values.Set("PROFILEID", profileId)

	return pClient.PerformRequest(values)
}

func NewDigitalGood(name string, amount float64) *PayPalDigitalGood {
	return &PayPalDigitalGood{
		Name:     name,
		Amount:   amount,
		Quantity: 1,
	}
}

type ExpressCheckoutSingleArgs struct {
	Amount                             float64
	CurrencyCode, ReturnURL, CancelURL string
	Recurring                          bool
	Item                               *PayPalDigitalGood
}

func NewExpressCheckoutSingleArgs() *ExpressCheckoutSingleArgs {
	return &ExpressCheckoutSingleArgs{
		Amount:       0,
		CurrencyCode: "USD",
		Recurring:    true,
	}
}

type ExpressCheckoutArgs struct {
	Amount                             float64
	CurrencyCode, ReturnURL, CancelURL string
	BillingAgreementDescription        string
	Brandname                          string
	LogoImg                            string
	Items                              []PayPalDigitalGood
}

// DoReferenceTransaction Completes a transaction through Billing Agreements
// see (https://developer.paypal.com/docs/classic/api/merchant/DoReferenceTransaction-API-Operation-NVP/ for more information
func (pClient *PayPalClient) DoReferenceTransaction(paymentAmount string, referenceID string, paymentMethod string, merchantSessionID string) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "DoReferenceTransaction")
	values.Add("AMT", paymentAmount)
	values.Add("PAYMENTACTION", paymentMethod)
	values.Add("REFERENCEID", referenceID)

	return pClient.PerformRequest(values)
}

// DoCapture captures an authorized payment, for our purposes it captures payments that are authorized by doReferenceTransaction which can be confusing because it returns a transactionID not an authorizationID
// however it can be used to capture any authorized payment
// See https://developer.paypal.com/docs/classic/api/merchant/DoCapture-API-Operation-NVP/ for details
func (pClient *PayPalClient) DoCapture(paymentAmount string, authorizationID string, isComplete bool, invoiceID string) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "DoCapture")
	values.Add("AMT", paymentAmount)
	values.Add("INVNUM", invoiceID)
	values.Add("AUTHORIZATIONID", authorizationID)
	if isComplete {
		values.Add("COMPLETETYPE", "Complete")
	} else {
		values.Add("COMPLETETYPE", "NotComplete")
	}

	return pClient.PerformRequest(values)
}

// DoVoid voids an authorized payment
// See https://developer.paypal.com/docs/classic/api/merchant/DoCapture-API-Operation-NVP/ for details
func (pClient *PayPalClient) DoVoid(authorizationID, note, messageID string) (*PayPalResponse, error) {
	err := new(PayPalError)
	if len(authorizationID) > 19 {
		err.Ack = "failure"
		err.ErrorCode = "0"
		err.ShortMessage = "authorizationID is longer than 19 characters"
		err.SeverityCode = "0"
		return nil, err
	}
	if len(note) > 255 {
		err.Ack = "failure"
		err.ErrorCode = "0"
		err.ShortMessage = "authorizationID is longer than 255 characters"
		err.SeverityCode = "0"
		return nil, err
	}
	if len(messageID) > 38 {
		err.Ack = "failure"
		err.ErrorCode = "0"
		err.ShortMessage = "authorizationID is longer than 38 characters"
		err.SeverityCode = "0"
		return nil, err
	}
	values := url.Values{}
	values.Set("METHOD", "DoVoid")
	values.Add("AUTHORIZATIONID", authorizationID)
	values.Add("NOTE", note)
	values.Add("MSGSUBID", messageID)

	return pClient.PerformRequest(values)
}

// ConvertResponse takes the url.Values from the PayPal Response and places them into struct
// According to their docs: Ack, CorrelationID, Timestamp, Version, and Build should be in every response
// from PayPal so we return the values placed in the root of the PayPalResponse struct instead
// of checking the url.Values array. Since some values may not be provided in the PayPalResponse struct
// We have to check the response
func (pClient *PayPalClient) ConvertResponse(paypalResponse PayPalResponse) *PayPalValues {
	return &PayPalValues{
		Ack:                  paypalResponse.Ack,
		Amount:               pClient.parseResponse(paypalResponse.Values["AMT"]),
		BillingAgreementID:   pClient.parseResponse(paypalResponse.Values["BILLINGAGREEMENTID"]),
		Build:                pClient.parseResponse(paypalResponse.Values["BUILD"]),
		CorrelationID:        paypalResponse.CorrelationID,
		CurrencyCode:         pClient.parseResponse(paypalResponse.Values["CURRENCYCODE"]),
		DateOrdered:          pClient.parseResponse(paypalResponse.Values["ORDERTIME"]),
		ErrorCode:            pClient.parseResponse(paypalResponse.Values["ERRORCODE0"]),
		ErrorMessage:         pClient.parseResponse(paypalResponse.Values["L_SHORTMESSAGE0"]),
		ErrorMessageExtended: pClient.parseResponse(paypalResponse.Values["L_LONGMESSAGE0"]),
		SeverityCode:         pClient.parseResponse(paypalResponse.Values["L_SEVERITYCODE0"]),
		PaymentStatus:        pClient.parseResponse(paypalResponse.Values["PAYMENTSTATUS"]),
		PendingReason:        pClient.parseResponse(paypalResponse.Values["PENDINGREASON"]),
		ReasonCode:           pClient.parseResponse(paypalResponse.Values["REASONCODE"]),
		Timestamp:            paypalResponse.Timestamp,
		TransactionID:        pClient.parseResponse(paypalResponse.Values["TRANSACTIONID"]),
		TransactionType:      pClient.parseResponse(paypalResponse.Values["TRANSACTIONTYPE"]),
		Version:              pClient.parseResponse(paypalResponse.Values["VERSION"]),
	}
}

// parseResponse is a helper function for convert response. this simple functionality
// was pulled out to reduce duplicate code
func (pClient *PayPalClient) parseResponse(s []string) string {
	if s != nil {
		return s[0]
	}
	return ""
}

// SetExpressCheckoutInitiateBilling is the first step to create a billing agreement. It returns a token that should be used to redirect the user so they can agree to recurring billing of varying quantities
// the token returned is not a billing agreement, however, it must be created once the user has approved
// See https://developer.paypal.com/docs/classic/express-checkout/ec-set-up-reference-transactions/# for details
func (pClient *PayPalClient) SetExpressCheckoutInitiateBilling(cancelURL string, returnURL string, currencyCode string, billingAgreementDescription string) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "SetExpressCheckout")
	values.Add("PAYMENTREQUEST_0_PAYMENTACTION", "AUTHORIZATION")
	values.Add("PAYMENTREQUEST_0_AMT", "00.00")
	values.Add("PAYMENTREQUEST_0_CURRENCYCODE", currencyCode)

	values.Add("L_BILLINGTYPE0", "MerchantInitiatedBilling")
	values.Add("L_BILLINGAGREEMENTDESCRIPTION0", billingAgreementDescription)

	values.Add("RETURNURL", returnURL)
	values.Add("CANCELURL", cancelURL)

	return pClient.PerformRequest(values)
}

// CreateBillingAgreement will create a billing agreement with the provided token. Once a billing agreement id has been obtained, your backend can conduct transactions against the corresponding user
// for as long as the billing agreement, without explicit user approval through the doReferenceTransaction call
// See https://developer.paypal.com/docs/classic/express-checkout/ec-set-up-reference-transactions/# for details
func (pClient *PayPalClient) CreateBillingAgreement(token string) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "CreateBillingAgreement")
	values.Add("TOKEN", token)

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) SetExpressCheckoutSingle(args *ExpressCheckoutSingleArgs) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "SetExpressCheckout")
	values.Add("PAYMENTREQUEST_0_AMT", fmt.Sprintf("%.2f", args.Amount))
	values.Add("NOSHIPPING", "1")

	values.Add("L_PAYMENTREQUEST_0_NAME0", args.Item.Name)
	values.Add("L_BILLINGTYPE0", "RecurringPayments")
	values.Add("L_BILLINGAGREEMENTDESCRIPTION0", args.Item.Name)
	values.Add("L_PAYMENTTYPE0", "InstantOnly")

	values.Add("RETURNURL", args.ReturnURL)
	values.Add("CANCELURL", args.CancelURL)

	return pClient.PerformRequest(values)
}

type Action string

const (
	Cancel     Action = "Cancel"
	Suspend    Action = "Suspend"
	Reactivate Action = "Reactivate"
)

func (pClient *PayPalClient) ManageRecurringPaymentsProfileStatus(profileId string, action Action) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "ManageRecurringPaymentsProfileStatus")
	values.Set("PROFILEID", profileId)
	values.Set("ACTION", string(action))

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) UpdateRecurringPaymentsProfile(profileId string, params map[string]string) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "UpdateRecurringPaymentsProfile")
	values.Set("PROFILEID", profileId)

	if params != nil {
		for key, value := range params {
			values.Add(key, value)
		}
	}

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) GetRecurringPaymentsProfileDetails(profileId string) (*PayPalResponse, error) {
	values := url.Values{}

	values.Set("METHOD", "GetRecurringPaymentsProfileDetails")
	values.Set("PROFILEID", profileId)

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) ProfileTransactionSearch(profileId string, startDate time.Time) (*PayPalResponse, error) {
	values := url.Values{}

	values.Set("PROFILEID", profileId)
	values.Set("METHOD", "TransactionSearch")
	values.Set("STARTDATE", startDate.Format(time.RFC3339))

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) RefundFullTransaction(transactionID string) (*PayPalResponse, error) {
	values := url.Values{}

	values.Set("METHOD", "RefundTransaction")
	values.Set("TRANSACTIONID", transactionID)
	values.Set("REFUNDTYPE", "FULL")

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) RefundPartialTransaction(transactionID string, amount string) (*PayPalResponse, error) {
	values := url.Values{}

	values.Set("METHOD", "RefundTransaction")
	values.Set("TRANSACTIONID", transactionID)
	values.Set("REFUNDTYPE", "PARTIAL")
	values.Set("AMT", amount)

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) SetExpressCheckoutPaymentAndInitiateBilling(args *ExpressCheckoutArgs) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "SetExpressCheckout")

	values.Add("PAYMENTREQUEST_0_AMT", fmt.Sprintf("%.2f", args.Amount))
	values.Add("PAYMENTREQUEST_0_PAYMENTACTION", "Sale")
	values.Add("PAYMENTREQUEST_0_CURRENCYCODE", args.CurrencyCode)
	values.Add("PAYMENTREQUEST_0_DESC", args.Items[0].Name)
	values.Add("PAYMENTREQUEST_0_ALLOWEDPAYMENTMETHOD", "InstantPaymentOnly")

	for i := 0; i < len(args.Items); i++ {
		item := args.Items[i]
		values.Add(fmt.Sprintf("%s%d", "L_PAYMENTREQUEST_0_QTY", i), fmt.Sprintf("%d", item.Quantity))
	}

	values.Add("L_BILLINGTYPE0", "MerchantInitiatedBillingSingleAgreement")
	values.Add("L_BILLINGAGREEMENTDESCRIPTION0", args.BillingAgreementDescription)
	values.Add("L_PAYMENTTYPE0", "Any")

	values.Add("RETURNURL", args.ReturnURL)
	values.Add("CANCELURL", args.CancelURL)
	values.Add("REQCONFIRMSHIPPING", "0")
	values.Add("NOSHIPPING", "1")
	values.Add("SOLUTIONTYPE", "Mark")
	values.Add("LANDINPAGE", "Login")
	values.Add("CHANNELTYPE", "Merchant")
	values.Add("BRANDNAME", args.Brandname)
	values.Add("LOGOIMG", args.LogoImg)

	return pClient.PerformRequest(values)
}
