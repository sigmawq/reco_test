package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type AsanaUser struct {
	Gid          string `json:"gid"`
	Name         string `json:"name"`
	ResourceType string `json:"resource_type"`
}

func AsanaGetUsers() ([]AsanaUser, error) {
	type Response struct {
		Data []AsanaUser `json:"data"`
	}

	request, err := http.NewRequest("GET", "https://app.asana.com/api/1.0/users", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+config["asana_secret"])

	response, err := CallWithRetry(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	var responseJson Response
	err = json.Unmarshal(body, &responseJson)
	if err != nil {
		return nil, err
	}

	return responseJson.Data, nil
}

type AsanaWorkspace struct {
	Gid          string `json:"gid"`
	Name         string `json:"name"`
	ResourceType string `json:"resource_type"`
}

func AsanaGetWorkspaceList() ([]AsanaWorkspace, error) {
	type Response struct {
		Data []AsanaWorkspace `json:"data"`
	}

	request, err := http.NewRequest("GET", "https://app.asana.com/api/1.0/workspaces", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+config["asana_secret"])

	response, err := CallWithRetry(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	var responseJson Response
	err = json.Unmarshal(body, &responseJson)
	if err != nil {
		return nil, err
	}

	return responseJson.Data, nil
}

type AsanaProject struct {
	Gid          string `json:"gid"`
	Name         string `json:"name"`
	ResourceType string `json:"resource_type"`
}

func AsanaGetProjects(workspaceGID string) ([]AsanaProject, error) {
	type Response struct {
		Data []AsanaProject `json:"data"`
	}

	url := "https://app.asana.com/api/1.0/workspaces/%v/projects"
	url = fmt.Sprintf(url, workspaceGID)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+config["asana_secret"])

	response, err := CallWithRetry(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)

	var responseJson Response
	err = json.Unmarshal(body, &responseJson)
	if err != nil {
		return nil, err
	}

	return responseJson.Data, nil
}

var asanaTotalCallsLastMinute int
var asanaLastMinuteUnix int64

func CallWithRetry(r *http.Request) (*http.Response, error) {
	log.Printf("call %v", r.URL)

	currentMinute := (time.Now().Unix() / 60)
	if currentMinute != asanaLastMinuteUnix {
		log.Println("Rate limiter window swap..")
		asanaLastMinuteUnix = currentMinute
		asanaTotalCallsLastMinute = 0
	}

	_asanaRpm := config["asana_max_rpm"]
	asanaRpm, _ := strconv.ParseInt(_asanaRpm, 10, 64)
	if asanaRpm <= 0 {
		asanaRpm = 150
	}

	_retryCount := config["retry_count"]
	_retryPeriodMilliseconds := config["retry_period_milliseconds"]

	retryCount, _ := strconv.ParseInt(_retryCount, 10, 64)
	retryPeriodMilliseconds, _ := strconv.ParseInt(_retryPeriodMilliseconds, 10, 64)

	if retryCount < 1 {
		retryCount = 1
	}
	if retryPeriodMilliseconds < 50 {
		retryPeriodMilliseconds = 50
	}

	callI := 0
	var err error
	for callI < int(retryCount) {
		if int64(asanaTotalCallsLastMinute) < asanaRpm {
			asanaTotalCallsLastMinute++
			var response *http.Response
			response, err = http.DefaultClient.Do(r)
			if err == nil {
				if response.StatusCode == 200 {
					return response, nil
				}
			}

			err = errors.New("status code != 200")
			response.Body.Close()
		} else {
			log.Println("slow down...")
		}

		time.Sleep(time.Duration(retryPeriodMilliseconds) * time.Millisecond)
	}
	return nil, err
}
