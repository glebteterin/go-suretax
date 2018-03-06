package suretax

import (
	"testing"
	"bytes"
	"os"
	"net/http"
	"io/ioutil"
	"sync"
)

var testCli = SuretaxClient{"", "", sync.Mutex{}, nil}

func TestMain(m *testing.M) {
	SetDebugLogger(nil)

	retCode := m.Run()
	os.Exit(retCode)
}

func Test_buildRequest(t *testing.T) {
	req, err := testCli.buildRequest(getTestRequest())
	if err != nil {
		t.Fatal(err)
	}

	requestBody, err := requestBodyToString(req)
	if err != nil {
		t.Fatal(err)
	}

	expectedBody := "{\"request\":\"{\\\"ClientNumber\\\":\\\"000000001\\\",\\\"BusinessUnit\\\":\\\"\\\",\\\"ValidationKey\\\":\\\"D4E909CF-76C1-4940-A00F-9B80FA363DE3\\\",\\\"DataYear\\\":\\\"2017\\\",\\\"DataMonth\\\":\\\"11\\\",\\\"CmplDataYear\\\":\\\"2016\\\",\\\"CmplDataMonth\\\":\\\"06\\\",\\\"TotalRevenue\\\":\\\"100\\\",\\\"ReturnFileCode\\\":\\\"0\\\",\\\"ClientTracking\\\":\\\"Certi\\\",\\\"ResponseType\\\":\\\"D2\\\",\\\"ResponseGroup\\\":\\\"00\\\",\\\"STAN\\\":\\\"\\\",\\\"ItemList\\\":[{\\\"LineNumber\\\":\\\"01\\\",\\\"InvoiceNumber\\\":\\\"INV-002\\\",\\\"CustomerNumber\\\":\\\"001\\\",\\\"OrigNumber\\\":\\\"9043101723\\\",\\\"TermNumber\\\":\\\"9043101723\\\",\\\"BillToNumber\\\":\\\"9043101723\\\",\\\"TransDate\\\":\\\"05/26/2017\\\",\\\"BillingPeriodStartDate\\\":\\\"\\\",\\\"BillingPeriodEndDate\\\":\\\"\\\",\\\"Revenue\\\":\\\"100\\\",\\\"TaxIncludedCode\\\":\\\"0\\\",\\\"Units\\\":\\\"4\\\",\\\"UnitType\\\":\\\"00\\\",\\\"TaxSitusRule\\\":\\\"01\\\",\\\"TransTypeCode\\\":\\\"050104\\\",\\\"SalesTypeCode\\\":\\\"B\\\",\\\"RegulatoryCode\\\":\\\"99\\\",\\\"TaxExemptionCodeList\\\":[],\\\"ExemptReasonCode\\\":\\\"\\\",\\\"UDF\\\":\\\"\\\",\\\"UDF2\\\":\\\"\\\",\\\"CostCenter\\\":\\\"\\\",\\\"GLAccount\\\":\\\"\\\",\\\"MaterialGroup\\\":\\\"\\\",\\\"BillingDaysInPeriod\\\":\\\"0\\\",\\\"OriginCountryCode\\\":\\\"\\\",\\\"DestCountryCode\\\":\\\"\\\",\\\"Parameter1\\\":\\\"\\\",\\\"Parameter2\\\":\\\"\\\",\\\"Parameter3\\\":\\\"\\\",\\\"Parameter4\\\":\\\"\\\",\\\"Parameter5\\\":\\\"\\\",\\\"Parameter6\\\":\\\"\\\",\\\"Parameter7\\\":\\\"\\\",\\\"Parameter8\\\":\\\"\\\",\\\"Parameter9\\\":\\\"\\\",\\\"Parameter10\\\":\\\"\\\",\\\"CurrencyCode\\\":\\\"\\\",\\\"Seconds\\\":\\\"4\\\",\\\"Address\\\":{\\\"PrimaryAddressLine\\\":\\\"\\\",\\\"SecondaryAddressLine\\\":\\\"\\\",\\\"County\\\":\\\"\\\",\\\"City\\\":\\\"\\\",\\\"State\\\":\\\"\\\",\\\"PostalCode\\\":\\\"\\\",\\\"Plus4\\\":\\\"\\\",\\\"Country\\\":\\\"\\\",\\\"Geocode\\\":\\\"\\\",\\\"VerifyAddress\\\":\\\"false\\\"},\\\"P2PAddress\\\":{\\\"PrimaryAddressLine\\\":\\\"\\\",\\\"SecondaryAddressLine\\\":\\\"\\\",\\\"County\\\":\\\"\\\",\\\"City\\\":\\\"\\\",\\\"State\\\":\\\"\\\",\\\"PostalCode\\\":\\\"\\\",\\\"Plus4\\\":\\\"\\\",\\\"Country\\\":\\\"\\\",\\\"Geocode\\\":\\\"\\\",\\\"VerifyAddress\\\":\\\"false\\\"}}]}\"}"

	if requestBody != expectedBody {
		t.Fatalf("Expected request %s but got %s", expectedBody, requestBody)
	}
}

func Test_parseResponse(t *testing.T) {

	const lineNumber = "0"
	const message = "Bill To Number is Required"
	const responseCode = "9131"
	const transId = 616039832
	const invoiceNumber = "INV-002"
	const taxAmount = "8.46"

	resp, err := testCli.parseResponse(getTestResponse())
	if err != nil {
		t.Fatal(err)
	}

	if resp.TransId != transId {
		t.Fatalf("Expected TransId %v but got %v", transId, resp.TransId)
	}

	if len(resp.ItemMessages) != 1 {
		t.Fatalf("Expected ItemMessages length %v but got %v", 1, len(resp.ItemMessages))
	}

	if resp.ItemMessages[0].LineNumber != lineNumber {
		t.Fatalf("Expected LineNumber %v but got %v", lineNumber, resp.ItemMessages[0].LineNumber)
	}

	if resp.ItemMessages[0].Message != message {
		t.Fatalf("Expected Message %v but got %v", message, resp.ItemMessages[0].Message)
	}

	if resp.ItemMessages[0].ResponseCode != responseCode {
		t.Fatalf("Expected Message %v but got %v", responseCode, resp.ItemMessages[0].ResponseCode)
	}

	if len(resp.GroupList) != 1 {
		t.Fatalf("Expected GroupList length %v but got %v", 1, len(resp.GroupList))
	}

	if resp.GroupList[0].InvoiceNumber != invoiceNumber {
		t.Fatalf("Expected InvoiceNumber %v but got %v", invoiceNumber, resp.GroupList[0].InvoiceNumber)
	}

	if len(resp.GroupList[0].TaxList) != 4 {
		t.Fatalf("Expected TaxList length %v but got %v", 3, len(resp.GroupList[0].TaxList))
	}

	if resp.GroupList[0].TaxList[0].TaxAmount != taxAmount {
		t.Fatalf("Expected TaxAmount length %v but got %v", taxAmount, resp.GroupList[0].TaxList[0].TaxAmount)
	}
}

func Test_getClient_default(t *testing.T) {

	SetHttpClient(nil)

	cli := SuretaxClient{"", "", sync.Mutex{}, nil}

	c := cli.getClient()

	if c == nil {
		t.Fatal("getClient() supposed to return default client")
	}
}

func getTestRequest() *Request {
	r := &Request{}
	r.ClientNumber = "000000001"
	r.BusinessUnit = ""
	r.ValidationKey = "D4E909CF-76C1-4940-A00F-9B80FA363DE3"
	r.DataYear = "2017"
	r.DataMonth = "11"
	r.CmplDataYear = "2016"
	r.CmplDataMonth = "06"
	r.TotalRevenue = "100"
	r.ClientTracking = "Certi"
	r.ResponseType = "D2"
	r.ResponseGroup = "00"
	r.ReturnFileCode = "0"
	r.STAN = ""

	item := RequestItem{}
	item.LineNumber = "01"
	item.InvoiceNumber = "INV-002"
	item.CustomerNumber = "001"
	item.TransDate = "05/26/2017"
	item.BillingPeriodStartDate = ""
	item.BillingPeriodEndDate = ""
	item.Revenue = "100"
	item.TaxIncludedCode = "0"
	item.Units = "4"
	item.UnitType = "00"
	item.TaxSitusRule = "01"
	item.TransTypeCode = "050104"
	item.SalesTypeCode = "B"
	item.RegulatoryCode = "99"
	item.BillingDaysInPeriod = "0"
	item.TaxExemptionCodeList = []string{}

	item.OrigNumber = "9043101723"
	item.TermNumber = "9043101723"
	item.BillToNumber = "9043101723"
	item.Seconds = "4"

	addr := Address{}
	addr.VerifyAddress = "false"

	p2pAddress := P2PAddress{}
	p2pAddress.VerifyAddress = "false"


	item.Address = addr
	item.P2PAddress = p2pAddress

	r.ItemList = []RequestItem{
		item,
	}

	return r
}

func getTestResponse() *http.Response {

	data := "{\"d\":\"{\\\"ClientTracking\\\":\\\"Certi\\\",\\\"ItemMessages\\\":[{\\\"LineNumber\\\":\\\"0\\\",\\\"Message\\\":\\\"Bill To Number is Required\\\",\\\"ResponseCode\\\":\\\"9131\\\"}],\\\"GroupList\\\":[{\\\"CustomerNumber\\\":\\\"001\\\",\\\"InvoiceNumber\\\":\\\"INV-002\\\",\\\"LineNumber\\\":\\\"01\\\",\\\"LocationCode\\\":\\\"\\\",\\\"StateCode\\\":\\\"FL\\\",\\\"TaxList\\\":[{\\\"CityName\\\":\\\"FERNANDINA BEACH\\\",\\\"CountyName\\\":\\\"NASSAU\\\",\\\"FeeRate\\\":0,\\\"Juriscode\\\":\\\"\\\",\\\"PercentTaxable\\\":1.000000,\\\"Revenue\\\":\\\"100.00\\\",\\\"RevenueBase\\\":\\\"113.71\\\",\\\"TaxAmount\\\":\\\"8.46\\\",\\\"TaxAuthorityID\\\":\\\"12009\\\",\\\"TaxAuthorityName\\\":\\\"FLORIDA, STATE OF\\\",\\\"TaxOnTax\\\":\\\"1.02\\\",\\\"TaxRate\\\":0.074400000000,\\\"TaxTypeCode\\\":\\\"127\\\",\\\"TaxTypeDesc\\\":\\\"FL COMMUNICATION SERVICES TAX\\\"},{\\\"CityName\\\":\\\"FERNANDINA BEACH\\\",\\\"CountyName\\\":\\\"NASSAU\\\",\\\"FeeRate\\\":0,\\\"Juriscode\\\":\\\"\\\",\\\"PercentTaxable\\\":0.649000,\\\"Revenue\\\":\\\"100.00\\\",\\\"RevenueBase\\\":\\\"64.89\\\",\\\"TaxAmount\\\":\\\"12.20\\\",\\\"TaxAuthorityID\\\":\\\"16\\\",\\\"TaxAuthorityName\\\":\\\"FEDERAL COMMUNICATIONS COMMISSION\\\",\\\"TaxOnTax\\\":\\\"0.00\\\",\\\"TaxRate\\\":0.188000000000,\\\"TaxTypeCode\\\":\\\"035\\\",\\\"TaxTypeDesc\\\":\\\"FEDERAL UNIVERSAL SERVICE FUND\\\"},{\\\"CityName\\\":\\\"FERNANDINA BEACH\\\",\\\"CountyName\\\":\\\"NASSAU\\\",\\\"FeeRate\\\":0,\\\"Juriscode\\\":\\\"\\\",\\\"PercentTaxable\\\":1.000000,\\\"Revenue\\\":\\\"100.00\\\",\\\"RevenueBase\\\":\\\"113.64\\\",\\\"TaxAmount\\\":\\\"6.50\\\",\\\"TaxAuthorityID\\\":\\\"4542\\\",\\\"TaxAuthorityName\\\":\\\"FERNANDINA BEACH, CITY OF\\\",\\\"TaxOnTax\\\":\\\"0.78\\\",\\\"TaxRate\\\":0.057200000000,\\\"TaxTypeCode\\\":\\\"337\\\",\\\"TaxTypeDesc\\\":\\\"LOCAL COMMUNICATIONS SVC. TAX\\\"},{\\\"CityName\\\":\\\"FERNANDINA BEACH\\\",\\\"CountyName\\\":\\\"NASSAU\\\",\\\"FeeRate\\\":0,\\\"Juriscode\\\":\\\"\\\",\\\"PercentTaxable\\\":0.649000,\\\"Revenue\\\":\\\"100.00\\\",\\\"RevenueBase\\\":\\\"65.09\\\",\\\"TaxAmount\\\":\\\"1.49\\\",\\\"TaxAuthorityID\\\":\\\"16\\\",\\\"TaxAuthorityName\\\":\\\"FEDERAL COMMUNICATIONS COMMISSION\\\",\\\"TaxOnTax\\\":\\\"0.00\\\",\\\"TaxRate\\\":0.022890000000,\\\"TaxTypeCode\\\":\\\"060\\\",\\\"TaxTypeDesc\\\":\\\"FEDERAL COST RECOVERY CHARGE\\\"}]}],\\\"HeaderMessage\\\":\\\"Success\\\",\\\"MasterTransId\\\":616039832,\\\"ResponseCode\\\":\\\"9999\\\",\\\"STAN\\\":\\\"\\\",\\\"Successful\\\":\\\"Y\\\",\\\"TotalTax\\\":\\\"28.65\\\",\\\"TransId\\\":616039832}\"}"

	r := &http.Response{}
	r.Body = ioutil.NopCloser(bytes.NewReader([]byte(data)))

	return r
}

func requestBodyToString(req *http.Request) (string, error) {

	br, err := req.GetBody()
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(br)
	requestBody := buf.String()
	return requestBody, nil
}