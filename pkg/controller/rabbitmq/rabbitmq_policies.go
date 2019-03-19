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
func cleanPolicies(apiService string, reqLogger logr.Logger, secret basicAuthCredentials) {
	reqLogger.Info("Removing all policies")
	url := apiService + "/api/policies"
	response := getRequest(url, secret)
	// request will return something like that:
	// [{"vhost":"dts","name":"ha-three","pattern":".*","apply-to":"all","definition":
	//{"ha-mode":"exactly","ha-params":3,"ha-sync-mode":"automatic"},"priority":0}]

	reqLogger.Info("Removing all policies list", "response", string(response))
	var policies []rabbitmqv1.RabbitmqPolicy
	err := json.Unmarshal(response, &policies)
	if err != nil {
		// something bad
		reqLogger.Info("Error parsing json!", err.Error())
	} else {
		for _, policyRecord := range policies {
			reqLogger.Info("Removing " + policyRecord.Name)
			deleteRequest(url+"/"+policyRecord.Vhost+"/"+policyRecord.Name, secret)
		}
	}
}

// setPolicies run as go routine
func (r *ReconcileRabbitmq) setPolicies(ctx context.Context, reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secretNames secretResouces) {
	var secret basicAuthCredentials

	username, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "username")
	secret.username = username
	if err != nil {
		reqLogger.Info("Policies: auth username not found")
	}

	password, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "password")
	secret.password = password
	if err != nil {
		reqLogger.Info("Policies: auth password not found")
	}

	// wait http connection to api port
	timeout := time.Duration(5 * time.Second)
	apiHostname := cr.Name + "-api:15672"
	apiService := "http://" + apiHostname

	_, err = net.DialTimeout("tcp", apiHostname, timeout)
	if err != nil {
		reqLogger.Info("Rabbitmq API service failed", "Error", err.Error())
	} else {
		reqLogger.Info("Using API service: "+apiService, secret.username, secret.password)
	}

	//clean rabbit before fulfilling policies list
	cleanPolicies(apiService, reqLogger, secret)

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
		putRequest(url, string(policyJSON), secret)
	}
}
