package rabbitmq

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func getRequest(url string) []byte {
	reqLogger := log.WithValues("getRequest", url)
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		reqLogger.Info("Failed creating request", err)
	}
	resp, err := client.Do(request)
	if err != nil {
		reqLogger.Info("Failed executing request", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Info("Failed reading responce", err)
	}
	return body
}

func putRequest(url string, data string) {
	client := &http.Client{}
	request, _ := http.NewRequest("PUT", url, strings.NewReader(data))
	resp, err := client.Do(request)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
}

func deleteRequest(url string) {
	client := &http.Client{}
	request, _ := http.NewRequest("DELETE", url, nil)
	resp, err := client.Do(request)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
}
