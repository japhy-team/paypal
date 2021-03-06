package payflow

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// These constants specify the URL that the library hits
const (
	PayflowSandboxURL    = "https://pilot-payflowpro.paypal.com"
	PayflowProductionURL = "https://payflowpro.paypal.com"
)

// PayPalClient is the type you should use for your Payflow API Requests
type PayPalClient struct {
	Username    string
	Password    string
	Vendor      string
	Partner     string
	Endpoint    string
	UsesSandbox bool
	Client      *http.Client
}

// PayPalCreditCard is composed of the data required to conduct a transaction against the payflow API with a credit card.
// ExpirationDate is of the format MMYY
type PayPalCreditCard struct {
	PAN     string `json:"pan"`
	Amount  string `json:"amount"`
	ExpDate string `json:"expirationDate"`
}

// PayPalResponse encompases a generic response from PayFlow
type PayPalResponse struct {
	Result          string     `json:"Result"`
	ResponseMessage string     `json:"ResponseMessage"`
	Values          url.Values `json:"Values"`
	UsedSandbox     bool
}

// PayPalValues encapsulates all the possible return values that could come back from Payflow. See below docs:
// https://developer.paypal.com/docs/classic/payflow/integration-guide/#transaction-responses
type PayPalValues struct {
	AdditionalMessages    string `json:"ADDLMSGS,omitempty"`
	Amount                string `json:"AMT,omitempty"`
	AmexID                string `json:"AMEXID,omitempty"`    // VERBOSITY=HIGH
	AmexPOSID             string `json:"AMEXPOSID,omitempty"` //VERBOSITY=HIGH
	AuthCode              string `json:"AUTHCODE,omitempty"`
	AVSAddress            string `json:"AVSADDR,omitempty"`
	AVSZipcode            string `json:"AVSZIP,omitempty"`
	AVSInternational      string `json:"IAVS,omitempty"`
	CardType              string `json:"CARDTYPE,omitempty"` //VERBOSITY=HIGH
	CorrelationID         string `json:"CORRELATIONID,omitempty"`
	CCTransID             string `json:"CCTRANSID,omitempty"`
	CCTransPOSData        string `json:"CCTRANS_POSDATA,omitempty"`
	CVV2Match             rune   `json:"CVV2MATCH,omitempty"`
	DateToSettle          string `json:"DATE_TO_SETTLE,omitempty"` //This parameter is returned in the response for inquiry transactions only (TRXTYPE=I)
	Duplicate             string `json:"DUPLICATE,omitempty"`      // - DUPLICATE=2 — ORDERID has already been submitted in a previous request with the same ORDERID.  - DUPLICATE=1 — The request ID has already been submitted for a previous request.  - DUPLICATE=-1 — The Gateway database is not available. PayPal cannot determine whether this is a duplicate order or request.
	EmailMatch            rune   `json:"EMAILMATCH,omitempty"`
	ExtraProcessorMessage string `json:"EXTRAPMSG,omitempty"`
	HostCode              string `json:"HOSTCODE,omitempty"` //VERBOSITY=HIGH
	OriginalAmount        string `json:"ORIGAMT,omitempty"`
	PaymentAdviceCode     string `json:"PAYMENTADVICECODE,omitempty"` // A value of 03 or 21 indicates it is the merchant's responsibility to stop this recurring transaction. These two codes indicate that either the account was closed, fraud was involved, or the cardholder has asked the bank to stop this payment for another reason. Even if a re-attempted transaction is successful, it will likely result in a chargeback.
	PaymentType           string `json:"PAYMENTTYPE,omitempty"`
	PhoneMatch            rune   `json:"PHONEMATCH,omitempty"`
	PNREF                 string `json:"PNREF,omitempty"`
	PPREF                 string `json:"PPREF,omitempty"`
	ProCardSecure         rune   `json:"PROCCARDSECURE,omitempty"` //VERBOSITY=HIGH
	ProcessorAVS          rune   `json:"PROCAVS,omitempty"`        //VERBOSITY=HIGH
	ProcessorCVV2         rune   `json:"PROCCVV2,omitempty"`       //VERBOSITY=HIGH
	Result                int    `json:"RESULT,omitempty"`
	ResponseMessage       string `json:"RESPMSG,omitempty"`
	ResponseText          string `json:"RESPTEXT,omitempty"` //VERBOSITY=HIGH
	TimeOfTransaction     string `json:"TRANSTIME,omitempty"`
	TransactionState      int    `json:"TRANSSTATE,omitempty"` // State of the transaction sent in an Inquiry response or with errors associated with Fraud Protection Service (FPS) transactions
}

// PayPalError is used when RESP is anything but 0.
// It exists only to indicate to the caller that something went wrong and does not add functionality / information
type PayPalError struct {
	ErrorCode    string
	ErrorMessage string
}

func (e *PayPalError) Error() string {
	return "Payflow API Call failed. Response Code: " + e.ErrorCode + " Response Message: " + e.ErrorMessage
}

// NewClient is a required method call before any API calls are made. Username, Password, Partner, Vendor are all values from paypal's merchant website.
// Sandbox environment variables usually have the vendor and username as the same value.
func NewClient(username, password, partner, vendor string, usesSandbox bool) *PayPalClient {
	endpoint := PayflowProductionURL
	if usesSandbox {
		endpoint = PayflowSandboxURL
	}

	return &PayPalClient{
		Username:    username,
		Password:    password,
		Partner:     partner,
		Vendor:      vendor,
		Endpoint:    endpoint,
		UsesSandbox: usesSandbox,
		Client:      new(http.Client),
	}
}

func (pClient *PayPalClient) performRequest(values url.Values) (*PayPalResponse, error) {
	values.Add("USER", pClient.Username)
	values.Add("PWD", pClient.Password)
	values.Add("PARTNER", pClient.Partner)
	values.Add("VENDOR", pClient.Vendor)

	formResponse, err := pClient.Client.PostForm(pClient.Endpoint, values)
	if err != nil {
		return nil, err
	}
	defer formResponse.Body.Close()

	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(formResponse.Body)
	if err != nil {
		return nil, err
	}

	responseValues, err := url.ParseQuery(string(body))
	response := &PayPalResponse{UsedSandbox: pClient.UsesSandbox}
	if err == nil {
		response.Result = responseValues.Get("RESULT")
		response.ResponseMessage = responseValues.Get("RESPMSG")
		response.Values = responseValues

		if response.Result != "0" {
			pError := PayPalError{}
			pError.ErrorCode = response.Result
			pError.ErrorMessage = response.ResponseMessage

			err = &pError
		}
	}

	return response, err
}

//parseString(paypalResponse.Values["AMT"])
func convertResponse(paypalResponse *PayPalResponse) *PayPalValues {
	result, _ := strconv.Atoi(parseString(paypalResponse.Values["RESULT"]))
	transactionState, _ := strconv.Atoi(parseString(paypalResponse.Values["TRANSSATE"]))
	cvv2Match := parseRune(parseString(paypalResponse.Values["CVV2MATCH"]))
	emailMatch := parseRune(parseString(paypalResponse.Values["EMAILMATCH"]))
	phoneMatch := parseRune(parseString(paypalResponse.Values["PHONEMATCH"]))
	proCardSecure := parseRune(parseString(paypalResponse.Values["PROCCARDSECURE"]))
	processorAVS := parseRune(parseString(paypalResponse.Values["PROCAVS"]))
	processorCVV2 := parseRune(parseString(paypalResponse.Values["PROCCVV2"]))
	return &PayPalValues{
		AdditionalMessages:    parseString(paypalResponse.Values["ADDLMSGS"]),
		Amount:                parseString(paypalResponse.Values["AMT"]),
		AmexID:                parseString(paypalResponse.Values["AMEXID"]),
		AmexPOSID:             parseString(paypalResponse.Values["AMEXPOSID"]),
		AuthCode:              parseString(paypalResponse.Values["AUTHCODE"]),
		AVSAddress:            parseString(paypalResponse.Values["AVSADDR"]),
		AVSZipcode:            parseString(paypalResponse.Values["AVSZIP"]),
		AVSInternational:      parseString(paypalResponse.Values["IAVS"]),
		CardType:              parseString(paypalResponse.Values["CARDTYPE"]),
		CorrelationID:         parseString(paypalResponse.Values["CORRELATIONID"]),
		CCTransID:             parseString(paypalResponse.Values["CCTRANSID"]),
		CCTransPOSData:        parseString(paypalResponse.Values["CCTRANS_POSDATA"]),
		CVV2Match:             cvv2Match,
		DateToSettle:          parseString(paypalResponse.Values["DATE_TO_SETTLE"]), //This parameter is returned in the response for inquiry transactions only (TRXTYPE=I)
		Duplicate:             parseString(paypalResponse.Values["DUPLICATE"]),      // - DUPLICATE=2 — ORDERID has already been submitted in a previous request with the same ORDERID.  - DUPLICATE=1 — The request ID has already been submitted for a previous request.  - DUPLICATE=-1 — The Gateway database is not available. PayPal cannot determine whether this is a duplicate order or request.
		EmailMatch:            emailMatch,
		ExtraProcessorMessage: parseString(paypalResponse.Values["EXTRAPMSG"]),
		HostCode:              parseString(paypalResponse.Values["HOSTCODE"]),
		OriginalAmount:        parseString(paypalResponse.Values["ORIGAMT"]),
		PaymentAdviceCode:     parseString(paypalResponse.Values["PAYMENTADVICECODE"]), // A value of 03 or 21 indicates it is the merchant's responsibility to stop this recurring transaction. These two codes indicate that either the account was closed, fraud was involved, or the cardholder has asked the bank to stop this payment for another reason. Even if a re-attempted transaction is successful, it will likely result in a chargeback.
		PaymentType:           parseString(paypalResponse.Values["PAYMENTTYPE"]),
		PhoneMatch:            phoneMatch,
		PNREF:                 parseString(paypalResponse.Values["PNREF"]),
		PPREF:                 parseString(paypalResponse.Values["PPREF"]),
		ProCardSecure:         proCardSecure,
		ProcessorAVS:          processorAVS,
		ProcessorCVV2:         processorCVV2, //VERBOSITY=HIGH
		Result:                result,
		ResponseMessage:       parseString(paypalResponse.Values["RESPMSG"]),
		ResponseText:          parseString(paypalResponse.Values["RESPTEXT"]),
		TimeOfTransaction:     parseString(paypalResponse.Values["TRANSTIME"]),
		TransactionState:      transactionState,
	}
}

// parseString is a helper function for convert response. this simple functionality
// was pulled out to reduce duplicate code
func parseString(s []string) string {
	if s != nil {
		return s[0]
	}
	return ""
}

func parseRune(s string) rune {
	r := []rune(s)
	if len(r) != 0 {
		return r[0]
	}
	return rune(0)
}

// DoSale conducts a sale operation against payflow
// PayPalCreditCard have a Card Number (PAN), Amount specified, and an expiration data in the format of MMYY
func (pClient *PayPalClient) DoSale(c PayPalCreditCard) (*PayPalValues, error) {
	values := url.Values{}
	values.Set("TRXTYPE", "S")
	values.Set("TENDER", "C")
	values.Set("ACCT", c.PAN)
	values.Set("AMT", c.Amount)
	values.Set("EXPDATE", c.ExpDate)

	res, err := pClient.performRequest(values)
	// log.Printf("%v", res.Values)
	return convertResponse(res), err
}

// DoAuth conducts an authorization against payflow
// PayPalCreditCard have a Card Number (PAN), Amount specified, and an expiration data in the format of MMYY
// isPartialAuthorization specifies if a partial authorization is acceptable. Read Below notes about authorizations for more information
func (pClient *PayPalClient) DoAuth(c PayPalCreditCard, isPartialAuthorization bool) (*PayPalValues, error) {
	values := url.Values{}
	values.Set("TRXTYPE", "A")
	values.Set("TENDER", "C")
	values.Set("ACCT", c.PAN)
	values.Set("AMT", c.Amount)
	values.Set("EXPDATE", c.ExpDate)
	if isPartialAuthorization {
		values.Set("PARTIALAUTH", "Y")
		values.Set("VERBOSITY", "HIGH")
	}

	res, err := pClient.performRequest(values)
	return convertResponse(res), err
}

// Submitting Partial Authorizations

// A partial authorization is a partial approval of an authorization (TRXTYPE=A) transaction.
// A partial authorization approves a transaction when the balance available is less than the amount of the transaction.
// The transaction response returns the amount of the original transaction and the amount approved.

// When To Use Partial Authorizations
// Use partial authorizations to reduce the number of declines resulting from buyers spending more than their balance on prepaid cards.

// Say, for example, that you sell sportswear on your website. Joe purchases a pair of running shoes in the amount of $100.00. At checkout, Joe uses a gift card with a balance of $80.00 to pay. You request partial authorization of $100.00. The transaction response returns the original amount of $100.00 and the approved amount of $80.00.

// You can take either of the following actions:

// Accept the $80.00 and ask the buyer to provide an alternate payment for the additional $20.00.
// Reject the partial authorization and submit to the card issuer an authorization reversal (Void) for $80.00.
