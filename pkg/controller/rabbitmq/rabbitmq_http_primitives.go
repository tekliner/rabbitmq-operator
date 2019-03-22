package rabbitmq

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func getRequest(url string, secret basicAuthCredentials) ([]byte, error) {
	reqLogger := log.WithValues("getRequest", url)
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		reqLogger.Info("Failed creating request", "Error", err.Error())
		return nil, err
	}
	request.SetBasicAuth(secret.username, secret.password)
	resp, err := client.Do(request)
	if err != nil {
		reqLogger.Info("Failed executing request", "Error", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Info("Failed reading responce", "Error", err.Error())
		return nil, err
	}
	return body, nil
}

func putRequest(url string, secret basicAuthCredentials, data string) error {
	reqLogger := log.WithValues("putRequest", url)
	reqLogger.Info("putRequest", "URL", url, "Data", data)
	client := &http.Client{}
	request, _ := http.NewRequest("PUT", url, strings.NewReader(data))
	request.SetBasicAuth(secret.username, secret.password)
	resp, err := client.Do(request)
	if err != nil {
		reqLogger.Info("Failed reading responce", "Error", err.Error())
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Info("Failed reading body of responce", "Code", string(body))
		return err
	}
	reqLogger.Info("Reading body of responce", "Body", string(body), "Code", resp.StatusCode)
	return nil
}

func deleteRequest(url string, secret basicAuthCredentials, data string) error {
	reqLogger := log.WithValues("deleteRequest", url)
	client := &http.Client{}
	request, _ := http.NewRequest("DELETE", url, nil)
	request.SetBasicAuth(secret.username, secret.password)
	resp, err := client.Do(request)
	if err != nil {
		reqLogger.Info("Failed reading responce", "Error", err.Error())
		return err
	}
	defer resp.Body.Close()
	reqLogger.Info("Reading responce code", "Code", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Info("Failed reading body of responce", "Code", string(body))
		return err
	}
	reqLogger.Info("Reading body of responce", "Code", string(body))
	return nil
}
