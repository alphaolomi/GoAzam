package GoAzam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

type (
	// Payload to send to the bank checkout endpoint
	BankCheckoutPayload struct {
		// This is amount that will be charged from the given account (required)
		Amount string `json:"amount"`

		// Code of currency (required)
		CurrencyCode string `json:"currencyCode"`

		// This is the account number/MSISDN that consumer will provide. The amount will be deducted from this account (required)
		MerchantAccountNumber string `json:"merchantAccountNumber"`

		// Mobile number (required)
		MerchantMobileNumber string `json:"merchantMobileNumber"`

		// The name of the customer (optional)
		MerchantName string `json:"merchantName"`

		// One time password (required)
		OTP string `json:"otp"`

		// Bank provider. Currently on CRDB and NMB are supported (required)
		Provider string `json:"provider"`

		// This id belongs to the calling application. Maximum Allowed length for this field is 128 ascii characters (Optional)
		ReferenceID string `json:"referenceId"`

		// This is additional data you can provide (Optional)
		AdditionalProperties AdditionalProperties `json:"additionalProperties"`
	}

	ReferenceID struct {
		// Reference ID of the transaction
		ReferenceID string `json:"ReferenceID"`
	}

	Properties struct {
		// List of properties
		Properties ReferenceID `json:"properties"`
	}

	// Data received from the server after a successful transaction
	BankCheckoutResponse struct {
		// will return true if successful
		Success bool `json:"success"`
		// message received from the server. Will be empty for sandbox
		Message string `json:"msg"`
		// data received from the server
		Data Properties `json:"data"`
	}
)

// Function to access the bank checkout endpoint. It accepts a parameter of
// type BankCheckoutPayload and returns a value of type BankCheckoutResponse.
// and an error if any.
func (api *APICONTEXT) BankCheckout(bankPayload BankCheckoutPayload) (*BankCheckoutResponse, error) {

	v := reflect.ValueOf(bankPayload)

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == "" {
			return nil, fmt.Errorf("(Bank Checkout) Error: Field '%v' is required.", v.Type().Field(i).Name)
		}
	}

	jsonParameters, err := json.Marshal(bankPayload)

	if err != nil {
		return nil, err
	}

	url := api.BaseURL + "/azampay/bank/checkout"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonParameters)))

	bearer := fmt.Sprintf("Bearer %v", api.Bearer)

	req.Header.Set("Authorization", bearer)
	req.Header.Set("X-API-KEY", api.token)
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var bankResponse *BankCheckoutResponse

	if resp.StatusCode == 200 {
		decodeErr := json.NewDecoder(bytes.NewReader(body)).Decode(&bankResponse)
		if decodeErr != nil {
			if decodeErr == io.EOF {
				return nil, fmt.Errorf("(Bank Checkout) Error: Server returned an empty body.")
			}
			return nil, decodeErr
		}
		return bankResponse, nil

	} else if resp.StatusCode == 400 {
		var badRequest *BadRequestError

		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&badRequest); err != nil {
			return nil, fmt.Errorf("(Bank Checkout) Error decoding badrequest: %w", err)
		}

		return nil, fmt.Errorf(badRequest.Error())
	} else if resp.StatusCode == 417 {
		var unauthorized *Unauthorized

		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&unauthorized); err != nil {
			return nil, fmt.Errorf("(Bank Checkout) Error decoding unauthorized err: %w", err)
		}

		return nil, fmt.Errorf(unauthorized.Error())
	} else if resp.StatusCode == 500 {
		return nil, fmt.Errorf("(Bank Checkout) Internal Server Error: status code 500")
	} else {
		return nil, fmt.Errorf("(Bank Checkout) Error: status code %d", resp.StatusCode)
	}

}
