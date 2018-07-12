package suretax

import (
	"net/http"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"sync"
	"time"
	)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var httpClientOverride HttpClient = nil

// Sets the package's http client.
func SetHttpClient(client HttpClient) {
	httpClientOverride = client
}

type SuretaxClient struct {
	// SureTax post request url.
	Url string

	// SureTax cancel post request url.
	CancelUrl string

	mu         sync.Mutex
	httpClient HttpClient
}

func (c *SuretaxClient) Send(req *Request) (*Response, error) {

	cli := c.getClient()

	r, err := c.buildRequest(req)
	if err != nil {
		return nil, err
	}

	resp, err := cli.Do(r)
	if err != nil {
		return nil, err
	}

	logger.Debug("Response Code:", resp.StatusCode, "Status:", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SureTax returned " + resp.Status)
	}

	res, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *SuretaxClient) Cancel(req *CancelRequest) (*CancelResponse, error) {

	cli := c.getClient()

	r, err := c.buildCancelRequest(req)
	if err != nil {
		return nil, err
	}

	resp, err := cli.Do(r)
	if err != nil {
		return nil, err
	}

	logger.Debug("Response Code:", resp.StatusCode, "Status:", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SureTax returned " + resp.Status)
	}

	res, err := c.parseCancelResponse(resp)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *SuretaxClient) getClient() HttpClient {

	if httpClientOverride != nil {
		return httpClientOverride
	}

	if c.httpClient != nil {
		return c.httpClient
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.httpClient == nil {
		tr := &http.Transport{
			IdleConnTimeout: time.Second * 10,
		}
		c.httpClient = &http.Client{Transport: tr, Timeout: time.Minute * 5}
	}

	return c.httpClient
}

func (c *SuretaxClient) buildRequest(req *Request) (*http.Request, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	rw := requestWrapper{string(reqBytes)}
	reqWrapperBytes, err := json.Marshal(rw)
	if err != nil {
		return nil, err
	}

	logger.Debug("Request Data: ", string(reqWrapperBytes))

	reader := bytes.NewReader(reqWrapperBytes)

	r, err := http.NewRequest("POST", c.Url, reader)
	if err != nil {
		return nil, err
	}

	r.Header.Add("Content-Type", "application/json")

	return r, nil
}

func (c *SuretaxClient) buildCancelRequest(req *CancelRequest) (*http.Request, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	rw := cancelRequestWrapper{string(reqBytes)}
	reqWrapperBytes, err := json.Marshal(rw)
	if err != nil {
		return nil, err
	}

	logger.Debug("Request Data: ", string(reqWrapperBytes))

	reader := bytes.NewReader(reqWrapperBytes)

	r, err := http.NewRequest("POST", c.CancelUrl, reader)
	if err != nil {
		return nil, err
	}

	r.Header.Add("Content-Type", "application/json")

	return r, nil
}

func (c *SuretaxClient) parseResponse(resp *http.Response) (*Response, error) {

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	logger.Debug("Response Data: ", string(bodyBytes))

	respw := ResponseWrapper{}
	if err := json.Unmarshal(bodyBytes, &respw); err != nil {
		return nil, fmt.Errorf("Response Wrapper Unmarshal Failed. Error: %v", err)
	}

	res := &Response{}
	if err := json.Unmarshal([]byte(respw.D), res); err != nil {
		return nil, fmt.Errorf("Response Unmarshal Failed. Error: %v", err)
	}

	return res, nil
}

func (c *SuretaxClient) parseCancelResponse(resp *http.Response) (*CancelResponse, error) {

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	logger.Debug("Response Data: ", string(bodyBytes))

	res := &CancelResponse{}
	if err := json.Unmarshal(bodyBytes, res); err != nil {
		return nil, fmt.Errorf("Response Unmarshal Failed. Error: %v", err)
	}

	return res, nil
}

type requestWrapper struct {
	Request string `json:"request"`
}

type cancelRequestWrapper struct {
	Request string `json:"requestCancel"`
}

type Request struct {
	// Client ID Number – provided by CCH SureTax. Required. Max Len: 10
	ClientNumber string

	// Client’s Business Unit. Value for this field is not required. Max Len: 20
	BusinessUnit string

	// Validation Key provided by CCH SureTax. Required for client access to API function. Max Len: 36
	ValidationKey string

	// Required. YYYY – Year to use for tax calculation purposes
	DataYear string

	// Required. MM – Month to use for tax calculation purposes. Leading zero is preferred.
	DataMonth string

	// Required. YYYY – Year to use for recording the tax calculations for tax remittance purposes.
	CmplDataYear string

	// Required. MM – Month to use for recording the tax calculations for tax remittance purposes. Leading zero is preferred.
	CmplDataMonth string

	// Required. Format: $$$$$$$$$.CCCC
	// For Negative charges, the first position should have a minus (-) indicator.
	TotalRevenue string

	// Required.
	// 0 – Default
	// Q – Quote purposes – taxes are computed and returned in the response message for generating quotes.
	// No detailed tax information is saved in the CCH SureTax tables for reporting.
	ReturnFileCode string

	// Field for client transaction tracking. This value will be provided in the response data. Value for this field is not required, but preferred.
	// Max Len: 100
	ClientTracking string

	// Required. Determines how taxes are grouped for the response. Values:
	// 00 – Tax grouped by Line Item
	ResponseType string

	// Required. Determines the granularity of taxes and (optionally) the decimal precision for the tax calculations and amounts in the response.
	// First Position Values:
	// D – Detailed. Tax values are returned by tax type for all levels of tax (Federal, State, and Local).
	// (optional) Second Position Values:
	// An integer (1-9) determines the decimal places used for all tax responses.
	// If no value is provided, all taxes are returned with a default number decimal places. This default precision value varies by “engine”.
	// For example, to return tax amounts to four decimals places, the Response Type would be: D4
	ResponseGroup string

	// Optional value. A unique value provided by client for transaction audit purposes.
	// Max Len: 16
	STAN string

	ItemList []RequestItem
}

type RequestItem struct {
	// Used to identify an item within the request. If no value is provided, requests are numbered sequentially. Max Len: 40
	LineNumber string

	// Used for tax aggregation by Invoice. Must be alphanumeric. Max Len: 40
	InvoiceNumber string

	// Used for tax aggregation by Customer. Must be alphanumeric. Max Len: 40
	CustomerNumber string

	// Required when using Tax Situs Rule 01 or 03. Format: NPANXXNNNN
	OrigNumber string

	// Required when using Tax Situs Rule 01. Format: NPANXXNNNN
	TermNumber string

	// Required when using Tax Situs Rule 01 or 02. Format: NPANXXNNNN
	BillToNumber string

	// Required. Date of transaction. Valid date formats include: MM/DD/YYYY, MM-DD-YYYY, YYYY-MM-DDTHH:MM:SS
	TransDate string

	// Optional. Billing Period Start Date
	BillingPeriodStartDate string

	// Optional. Billing Period End Date
	BillingPeriodEndDate string

	// Required. Format: $$$$$$$$$.CCCC
	// For Negative charges, the first position should have a minus (-) indicator.
	Revenue string

	// Required. Values:
	// 0 – Default (No Tax Included) 1 – Tax Included in Revenue
	TaxIncludedCode string

	// Required.
	// Units representing number of “lines” or unique charges contained within the revenue.
	// This value is essentially a multiplier on unit-based fees (e.g. E911 fees).
	// Format: 99999. Default should be 1 (one unit).
	Units string

	// Required.
	// 00 – Default / Number of unique access lines. *See Appendix F for additional values.
	UnitType string

	// Required. Values:
	// 01 – Two-out-of-Three test using NPA-NXX
	// 02 – Billed to number
	// 03 – Origination number
	// 04 – Zip code
	// 05 – Zip code + 4
	// 07 – Point to Point Zip codes (private line transactions) *
	// 09 – Two-out-of-three test using Zip+4 as tax situs jurisdiction 11 – Used when Zipcode/Plus4 field is assigned as the Billing location and P2PZipcode/P2PPlus4 assigned as Service location*
	// 14 – Use Zip code field for international country code (VAT calculations)
	// 17 - Point to Point Zip codes (private line transactions) with both A and Z endpoints calculated*
	// 27 – Use only Billing Address / Zip+4
	TaxSitusRule string

	// Required. Transaction Type Indicator.
	TransTypeCode string

	// Required. Values:
	// R – Residential customer (default) B – Business customer
	// I – Industrial customer
	// L – Lifeline customer
	SalesTypeCode string

	// Required. Provider Type.
	RegulatoryCode string

	// Required. Tax Exemption to be applied to this item only.
	TaxExemptionCodeList []string

	// Required. Tax Exemption reason value.
	ExemptReasonCode string

	// Optional. Field for client use at the item level. This value is not returned in the response, but is available for use in reports and extracts.
	// Max Len: 100
	UDF string

	// Optional. Field for client use at the item level. This value is not returned in the response, but is available for use in reports and extracts.
	// Max Len: 100
	UDF2 string

	// Optional. Available for use in the rules engine.
	CostCenter string

	// Optional. Available for use in the rules engine. Max Len: 25 Alphanumeric
	GLAccount string

	// Optional. Available for use in the rules engine. Max Len: 25 Alphanumeric
	MaterialGroup string

	// Optional. Billing Days in Period.
	BillingDaysInPeriod string

	// Optional. Origin Country Code
	OriginCountryCode string

	// Optional. Destination Country Code
	DestCountryCode string

	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter1 string
	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter2 string
	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter3 string
	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter4 string
	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter5 string
	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter6 string
	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter7 string
	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter8 string
	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter9 string
	// Optional, user defined field. Available for use in the rules engine.
	// Max Len: 25 Alphanumeric
	Parameter10 string

	// Optional. Currency code based on ISO standard. As reference - http://www.xe.com/iso4217.php
	CurrencyCode string

	// Required. Duration of call in seconds. Format 99999. Default should be 1.
	Seconds string

	// Billing address for transaction
	Address Address

	// P2P address for transaction
	P2PAddress P2PAddress
}

type Address struct {
	// Address Line 1
	PrimaryAddressLine string

	// Address Line 2
	SecondaryAddressLine string

	// County
	County string

	// City
	City string

	// State – full state name or two-character abbreviation accepted
	State string

	// Zip code or Canadian postal code
	PostalCode string

	// Zip+4
	Plus4 string

	// International country ISO code value in format: XX
	Country string

	// Optional value. If provided, the CCH Geocode will take precedence and will be used instead of the address / zip+4 information.
	Geocode string

	// Required. If selected, SureAddress will validate address and update the zip+4.
	// 0 – No (default)
	// 1 – Yes
	VerifyAddress string
}

type P2PAddress struct {
	// Address Line 1
	PrimaryAddressLine string

	// Address Line 2
	SecondaryAddressLine string

	// County
	County string

	// City
	City string

	// State – full state name or two-character abbreviation accepted
	State string

	// Zip code or Canadian postal code
	PostalCode string

	// Zip+4
	Plus4 string

	// International country ISO code value in format: XX
	Country string

	// Optional value. If provided, the CCH Geocode will take precedence and will be used instead of the address / zip+4 information.
	Geocode string

	// Required. If selected, SureAddress will validate address and update the zip+4.
	// 0 – No (default)
	// 1 – Yes
	VerifyAddress string
}

type ResponseWrapper struct {
	D string `json:"d"`
}

type Response struct {
	// Client transaction tracking provided in web request.
	ClientTracking string

	// Response message:
	// For ResponseCode 9999 – “Success”
	// For ResponseCode 9001 – “Success with Item errors”. See the ItemMessages field for a list of items / errors.
	// For ResponseCode 1100-1400 – Unsuccessful / declined web request. See Appendix I for a list of the response code and messages.
	HeaderMessage string

	// This field contains a list of items that were not able to be processed due to bad or invalid data (see Response Code of “9001”).
	// These invalid items will be listed by line number with the corresponding response code and message.
	// When an item error occurs, no tax processing will occur for that item record.
	ItemMessages []ItemMessage

	// ResponseCode:
	// 9999 – Request was successful.
	// 1101-1400 – Range of values for a failed request (no processing occurred)
	// 9001 – Request was successful, but items within the request have errors.
	// The specific items with errors are provided in the ItemMessages field.
	ResponseCode string

	STAN string

	// Response will be either ‘Y' or ‘N' :
	// Y = Success / Success with Item error N = Failure
	Successful string

	// Transaction ID (integer) – provided by CCH SureTax
	TransId int

	// Total Tax – a total of all taxes included in the TaxList
	TotalTax string

	GroupList []Group
}

type ItemMessage struct {
	// Value corresponding to the line number in the web request
	LineNumber string

	// The error message corresponding to the ResponseCode.
	Message string

	// Value in the range 9100-9400.
	ResponseCode string
}

type Group struct {
	// Customer number
	CustomerNumber string

	// Invoice Number
	InvoiceNumber string

	// Line Number from Request
	LineNumber string

	LocationCode string

	// Tax State
	StateCode string

	// See Tax Item(s) – contains one-to-many Tax Items
	TaxList []Tax
}

type Tax struct {
	// City Name of taxing jurisdiction
	CityName string

	// County Name of taxing jurisdiction
	CountyName string

	// The unit based fee for the tax type in format $.CCCC
	FeeRate   float64
	Juriscode string

	// Percentage of the tax that is taxable for the tax type in decimal format.
	// This can be less than 100% in certain circumstances such as application of a Safe Harbor rate or a Private Line that is allocated to two or more points.
	PercentTaxable float64

	// Source Revenue for Line Item
	Revenue string

	// The effective revenue for the tax provided in the TaxAmount field.
	// This amount can be different than the amount in the Revenue field when taxes are impacted by specific exemptions and/or tax on tax.
	RevenueBase string

	// Tax Amount (taxes returned with five decimal places)
	TaxAmount        string
	TaxAuthorityID   string
	TaxAuthorityName string

	// The amount of tax on tax attributed to the final amount of tax.
	// Please note this amount is included in the TaxAmount field total and is provided here separately only for reference purposes.
	TaxOnTax string

	// Tax rate for tax type applied in decimal format
	TaxRate float64

	// Tax Type Code
	TaxTypeCode string

	// Jurisdiction-specific Tax Type Description
	TaxTypeDesc string
}

type CancelRequest struct {
	// Client ID Number – provided by CCH SureTax.
	ClientNumber string

	// Optional. Field for client transaction tracking.
	ClientTracking string

	// Transaction ID of web request to be cancelled.
	TransId string

	// Validation Key provided by CCH SureTax. Required for client access to API function.
	ValidationKey string
}

type CancelResponse struct {
	// Response will be either ‘Y' or ‘N' : Y = Success / Success with errors N = Failure
	Successful string

	// ResponseCode:
	// 9999 – Cancel Request was successful.
	// 1101-1600 – Range of values for a failed request (no processing occurred). Values include: 1150 - Failure - Validation Key Required.
	// 1151 - Failure - Invalid Validation Key.
	// 1510 - Failure - Transaction is more than 60 days old.
	// 9410 - Failure - Transaction is already cancelled.
	ResponseCode string

	// Response message:
	// For ResponseCode 9999 – “Success”
	// For ResponseCode 9001 – “Success with Item errors”. See the ItemMessages field above for a list of items / errors.
	// For ResponseCode 1100-1600 – Unsuccessful / declined web request. See above or (Appendix I) for a list.
	HeaderMessage string

	// Field for client transaction tracking.
	ClientTracking string

	// Transaction ID (integer) – provided by CCH SureTax
	TransId int
}
