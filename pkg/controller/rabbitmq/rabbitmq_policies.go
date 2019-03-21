package rabbitmq

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/go-logr/logr"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
)

// cleanPolicies return policies and remove all by name, no other method supported
func cleanPolicies(apiService string, reqLogger logr.Logger, secret basicAuthCredentials) error {
	reqLogger.Info("Removing all policies")
	url := apiService + "/api/policies"

	response, err := getRequest(url, secret)
	if err != nil {
		reqLogger.Info("Error while receiving policies list", "Error", err.Error())
		return err
	}
	reqLogger.Info("Removing all policies list", "response", string(response))

	var policies []rabbitmqv1.RabbitmqPolicy
	err = json.Unmarshal(response, &policies)
	if err != nil {
		// something bad
		reqLogger.Info("Error parsing json!", err.Error())
		return err
	} else {
		for _, policyRecord := range policies {
			reqLogger.Info("Removing " + policyRecord.Name)
			err = deleteRequest(url+"/"+policyRecord.Vhost+"/"+policyRecord.Name, secret)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// setPolicies run as go routine
func (r *ReconcileRabbitmq) setPolicies(ctx context.Context, reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secretNames secretResouces) error {
	var secret basicAuthCredentials

	username, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "username")
	secret.username = username
	if err != nil {
		reqLogger.Info("Policies: auth username not found")
		return err
	}

	password, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "password")
	secret.password = password
	if err != nil {
		reqLogger.Info("Policies: auth password not found")
		return err
	}

	// wait http connection to api port
	timeout := time.Duration(5 * time.Second)
	apiHostname := cr.Name + "-api." + cr.Namespace + ":15672"
	apiService := "http://" + apiHostname

	_, err = net.DialTimeout("tcp", apiHostname, timeout)
	if err != nil {
		reqLogger.Info("Rabbitmq API service failed", "Error", err.Error())
		return err
	} else {
		reqLogger.Info("Using API service: "+apiService, secret.username, secret.password)
	}

	//clean rabbit before fulfilling policies list
	err = cleanPolicies(apiService, reqLogger, secret)
	if err != nil {
		reqLogger.Info("Error happened while cleaning policies list")
		return err
	}

	reqLogger.Info("Uploading policies from CRD")

	// detect default vhost for policies
	policiesDefaultVhost := "%2f"
	if cr.Spec.RabbitmqVhost != "" {
		policiesDefaultVhost = cr.Spec.RabbitmqVhost
	}

	//fulfill policies list
	for _, policy := range cr.Spec.RabbitmqPolicies {
		policyJSON, _ := json.Marshal(policy)

		// detect vhost to use
		policyVhost := ""
		if policy.Vhost != "" {
			policyVhost = policy.Vhost
		} else {
			policyVhost = policiesDefaultVhost
		}

		url := apiService + "/api/policies/" + policyVhost + "/" + policy.Name
		// send policy to api service
		reqLogger.Info("Adding policy " + policy.Name + ", URL: " + url + ", JSON: " + string(policyJSON))
		err = putRequest(url, string(policyJSON), secret)
		if err != nil {
			reqLogger.Info("Error adding policy " + policy.Name + ", URL: " + url + ", JSON: " + string(policyJSON))
			return err
		}
	}

	return nil
}
