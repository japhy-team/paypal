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
	NVP_VERSION             = "84"
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

type PayPalRequest struct {
	Action                string `json:"action"`
	BillingMethodID       string `json:"billingMethodID"`
	Amount                string `json:"paymentAmount"`
	BillingAgreementToken string `json:"billingAgreementToken"`
	TransactionID         string `json:"transactionID"`
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
	Ack                       string `json:"ACK"`
	Amount                    string `json:"AMT"`
	BillingAgreementID        string `json:"BILLINGAGREEMENTID"`
	Build                     string `json:"BUILD"`
	CorrelationID             string `json:"CORRELATIONID"`
	CurrencyCode              string `json:"CURRENCYCODE"`
	DateOrdered               string `json:"ORDERTIME"`
	PaymentStatus             string `json:"PAYMENTSTATUS"`
	PaymentType               string `json:"PAYMENTTYPE"`
	PendingReason             string `json:"PENDINGREASON"`
	ProtectionEligibility     string `json:"PROTECTIONELIGIBLITY"`
	ProtectionEligibilityType string `json:"PROTECTIONELIGIBILITYTYPE"`
	ReasonCode                string `json:"REASONCODE"`
	TaxedAmount               string `json:"TAXAMT"`
	Timestamp                 string `json:"TIMESTAMP"`
	TransactionID             string `json:"TRANSACTIONID"`
	TransactionType           string `json:"TRANSACTIONTYPE"`
	Version                   string `json:"VERSION"`
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
		response.CorrelationId = responseValues.Get("CORRELATIONID")
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

func (pClient *PayPalClient) RefundTransaction(transactionID string, refundType string) (*PayPalResponse, error) {
	values := url.Values{}

	values.Set("METHOD", "RefundTransaction")
	values.Set("TRANSACTIONID", transactionID)
	values.Set("REFUNDTYPE", "FULL")

	return pClient.PerformRequest(values)
}
