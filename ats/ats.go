package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	AtsServiceHost     = "ats.amazonaws.com"
	AtsServiceEndpoint = "ats.us-west-1.amazonaws.com"
	AtsUri             = "/api"
	AtsApiUrl          = "https://" + AtsServiceHost + AtsUri

	AtsServiceName   = "AlexaTopSites"
	AtsServiceRegion = "us-west-1"

	DateFormat   = "20060102"
	DateTzFormat = "20060102T150405Z"
)

func main() {
	if len(os.Args) != 6 {
		Help()
		os.Exit(0)
	}

	accessKey := os.Args[1]
	secretKey := os.Args[2]
	country := os.Args[3]
	start := os.Args[4]
	count := os.Args[5]

	response := SendRequest(accessKey, secretKey, country, start, count)
	fmt.Print(response)
}

func SendRequest(accessKey, secretKey, country, start, count string) string {
	dateTz, query, authorization := CreateHeaders(accessKey, secretKey, country, start, count)
	url := AtsApiUrl + "?" + query

	// new request
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("X-Amz-Date", dateTz)
	req.Header.Set("Authorization", authorization)

	// send request
	client := &http.Client{
		// set 10s timeout
		Timeout: time.Duration(10 * time.Second),
	}
	response, err := client.Do(req)

	if err != nil {
		fmt.Println("Request Failed")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer response.Body.Close()

	// return response
	body, _ := ioutil.ReadAll(response.Body)
	return string(body)
}

func CreateHeaders(accessKey, secretKey, country, start, count string) (string, string, string) {
	location, _ := time.LoadLocation("UTC")
	now := time.Now().In(location)
	date := now.Format(DateFormat)
	dateTz := now.Format(DateTzFormat)

	query := "Action=TopSites&Count=" + count + "&CountryCode=" + country + "&ResponseGroup=Country" + "&Start=" + start
	headers := "host:" + AtsServiceEndpoint + "\n" + "x-amz-date:" + dateTz + "\n"
	signedHeaders := "host;x-amz-date"
	payload := Sha256Hex("")
	request := "GET" + "\n" + AtsUri + "\n" + query + "\n" + headers + "\n" + signedHeaders + "\n" + payload

	algorithm := "AWS4-HMAC-SHA256"
	scope := date + "/" + AtsServiceRegion + "/" + AtsServiceName + "/aws4_request"

	signingData := algorithm + "\n" + dateTz + "\n" + scope + "\n" + Sha256Hex(request)
	signingKey := GenSignatureKey(secretKey, date)
	signature := HmacSha256Hex(signingData, signingKey)

	authorization := algorithm + " " + "Credential=" + accessKey + "/" + scope + ", " + "SignedHeaders=" + signedHeaders + ", " + "Signature=" + signature
	return dateTz, query, authorization
}

func GenSignatureKey(secretKey, date string) string {
	kSecret := []byte("AWS4" + secretKey)
	kDate := HmacSha256(date, kSecret)
	kRegion := HmacSha256(AtsServiceRegion, kDate)
	kService := HmacSha256(AtsServiceName, kRegion)
	kSignature := HmacSha256("aws4_request", kService)
	return string(kSignature)
}

func Help() {
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("    ./ats access_key secret_key [country_code] [start_number] [count]")
	fmt.Println()
}
