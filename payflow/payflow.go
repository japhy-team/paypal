package payflow

import (
	"io/ioutil"
	"net/http"
	"net/url"
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

// PayPalError is used when RESP is anything but 0.
// It exists only to indicate to the caller that something went wrong and does not add functionality / information
type PayPalError struct {
	ErrorCode    string
	ErrorMessage string
}

func (e *PayPalError) Error() string {
	return "Payflow API Called failed. Resposnse Code: " + e.ErrorCode + " Response Message: " + e.ErrorMessage
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

// DoSale conducts a sale operation against payflow with the
func (pClient *PayPalClient) DoSale(c PayPalCreditCard) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("TRXTYPE", "S")
	values.Set("TENDER", "C")
	values.Set("ACCT", c.PAN)
	values.Set("AMT", c.Amount)
	values.Set("EXPDATE", c.ExpDate)

	return pClient.performRequest(values)
}
