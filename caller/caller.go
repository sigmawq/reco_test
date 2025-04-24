package main

import (
	"encoding/json"
	"fmt"
	"io"
	"moul.io/http2curl"
	"net/http"
	"os"
	"strings"
	"time"
)

type UseCase struct {
	Name       string
	Calls      []EndpointCall
	FailStatus bool
	FailReason string
}

type EndpointCall struct {
	URL        string
	StatusCode int
	Taken      time.Duration
	Body       string
}

var baseUrl string
var output strings.Builder
var i int
var sessionToken string
var sessionTokenExpiry time.Time
var summary []UseCase
var thisUseCase *UseCase

type RequestResponse struct {
	SessionToken       string
	SessionTokenExpiry time.Time
	StatusCode         int
	Response           map[string]interface{}
}

func Call(method string, url, body string) *RequestResponse {
	i += 1

	thisUseCase.Calls = append(thisUseCase.Calls, EndpointCall{})
	thisCall := &thisUseCase.Calls[len(thisUseCase.Calls)-1]
	thisCall.URL = url

	url = fmt.Sprintf("%v/%v", baseUrl, url)
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		fmt.Println("Error while creating request:", err)
		panic("Cannot continue")
	}

	if sessionToken != "" {
		request.AddCookie(&http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: sessionTokenExpiry,
		})
	}

	curl, _ := http2curl.GetCurlCommand(request)
	output.WriteString(fmt.Sprintf("%v. %v:\n", i, url))
	output.WriteString(curl.String())
	output.WriteString("\n\n\n")

	fmt.Printf("POST %v\n", url)

	now := time.Now()
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println("Error:", err)
		panic("Cannot continue")
	}
	then := time.Now()
	thisCall.Taken = then.Sub(now)

	responseBodyText, _ := io.ReadAll(res.Body)
	thisCall.Body = string(responseBodyText)

	fmt.Println("Status:", res.StatusCode)

	responseBody := make(map[string]interface{})
	err = json.Unmarshal(responseBodyText, &responseBody)
	var outputString string
	if err != nil {
		fmt.Printf("Response text: %v\n", string(responseBodyText))
		outputString = string(responseBodyText)
	} else {
		indentedData, _ := json.MarshalIndent(responseBody, "", "\t")
		outputString = string(indentedData)
		fmt.Printf("Response JSON: %v\n", string(indentedData))
	}

	output.WriteString(fmt.Sprintf("Response: %v\n", outputString))
	output.WriteString("---------------------------------------------------\n\n")

	thisCall.StatusCode = res.StatusCode

	cookies := res.Cookies()
	sessionToken := ""
	var sessionTokenExpiry time.Time
	for _, cookie := range cookies {
		if cookie.Name == "session_token" {
			sessionToken = cookie.Value
			sessionTokenExpiry = cookie.Expires
		}
	}

	r := RequestResponse{
		SessionToken:       sessionToken,
		SessionTokenExpiry: sessionTokenExpiry,
		Response:           responseBody,
		StatusCode:         thisCall.StatusCode,
	}

	return &r
}

func BeginUseCase(name string) {
	output = strings.Builder{}
	output.WriteString(fmt.Sprintf("[%v]\n", name))
	output.WriteString(time.Now().String())
	output.WriteString("\n\n")
	summary = append(summary, UseCase{
		Name: name,
	})
	thisUseCase = &summary[len(summary)-1]
	i = 0
}

func FlushUseCase() {
	if thisUseCase == nil {
		panic("Please call BeginUseCase() before FlushUseCase()")
	}

	os.WriteFile(fmt.Sprintf("%v.txt", thisUseCase.Name), []byte(output.String()), 0777)
	thisUseCase = nil
}

func Fail(reason string) {
	if thisUseCase == nil {
		panic("Success() has been called before BeginUseCase()")
	}

	thisUseCase.FailStatus = true
	thisUseCase.FailReason = reason
}

func Failf(format string, args ...any) {
	if thisUseCase == nil {
		panic("Success() has been called before BeginUseCase()")
	}

	thisUseCase.FailStatus = true
	reason := fmt.Sprintf(format, args)
	thisUseCase.FailReason = reason
}

func CallGoogle() {
	BeginUseCase("CallGoogle")
	defer FlushUseCase()

	r := Call("GET", ``, ``)
	if r.StatusCode != 200 {
		Failf("Status code %v", r.StatusCode)
		return
	}
}

func main() {
	baseUrl = "https://google.com"

	// *** Context independent use cases ***
	CallGoogle()

	// *** Context dependent calls ***

	fmt.Println("=========================================")
	for _, useCase := range summary {
		fmt.Printf("[%v]\n", useCase.Name)
		status := ""
		if !useCase.FailStatus {
			status = "OK"
		} else {
			status = "FAILED"
		}
		for _, ec := range useCase.Calls {
			fmt.Printf("        %v - [%v] Code: %v Time: %v (%v bytes)\n", status, ec.URL, ec.StatusCode, ec.Taken, len(ec.Body))
			if useCase.FailReason != "" {
				fmt.Printf("            Reason: %v\n", useCase.FailReason)
			}
		}

	}
}
