package rabbitmq

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type basicAuthCredentials struct {
	username string
	password string
}

func getRequest(url string, secret basicAuthCredentials) []byte {
	reqLogger := log.WithValues("getRequest", url)
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		reqLogger.Info("Failed creating request", "Error", err.Error())
	}
	request.SetBasicAuth(secret.username, secret.password)
	resp, err := client.Do(request)
	if err != nil {
		reqLogger.Info("Failed executing request", "Error", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Info("Failed reading responce", "Error", err.Error())
	}
	return body
}

func putRequest(url string, data string, secret basicAuthCredentials) {
	reqLogger := log.WithValues("putRequest", url)
	client := &http.Client{}
	request, _ := http.NewRequest("PUT", url, strings.NewReader(data))
	request.SetBasicAuth(secret.username, secret.password)
	resp, err := client.Do(request)
	if err != nil {
		reqLogger.Info("Failed reading responce", "Error", err.Error())
	}

	defer resp.Body.Close()
	reqLogger.Info("Reading responce code", "Code", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Info("Failed reading body of responce", "Code", string(body))
	}
	reqLogger.Info("Reading body of responce", "Code", string(body))
}

func deleteRequest(url string, secret basicAuthCredentials) {
	reqLogger := log.WithValues("deleteRequest", url)
	client := &http.Client{}
	request, _ := http.NewRequest("DELETE", url, nil)
	request.SetBasicAuth(secret.username, secret.password)
	resp, err := client.Do(request)
	if err != nil {
		reqLogger.Info("Failed reading responce", "Error", err.Error())
	}
	defer resp.Body.Close()
	reqLogger.Info("Reading responce code", "Code", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Info("Failed reading body of responce", "Code", string(body))
	}
	reqLogger.Info("Reading body of responce", "Code", string(body))

}
