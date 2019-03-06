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
func cleanPolicies(apiService string, reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) {
	url := apiService + "/api/policies"

	// request will return something like that:
	// [{"vhost":"dts","name":"ha-three","pattern":".*","apply-to":"all","definition":
	//{"ha-mode":"exactly","ha-params":3,"ha-sync-mode":"automatic"},"priority":0}]

	response := getRequest(url)
	var policies []rabbitmqv1.RabbitmqPolicy
	err := json.Unmarshal(response, &policies)
	if err != nil {
		// something bad
	}

	for _, policyRecord := range policies {
		deleteRequest(apiService + "/" + policyRecord.Vhost + "/" + policyRecord.Name)
	}

}

// setPolicies run as go routine
func setPolicies(ctx context.Context, reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) {
	// wait http connection to api port
	timeout := time.Duration(5 * time.Second)
	apiService := cr.Name + "-node:15672"

	for {
		_, err := net.DialTimeout("tcp", apiService, timeout)
		if err != nil {
			reqLogger.Info("Rabbitmq API service failed")
			break
		}
	}

	//clean rabbit before fulfilling policies list
	cleanPolicies(apiService, reqLogger, cr)

	//fulfill policies list
	for _, policy := range cr.Spec.RabbitmqPolicies {
		policyJSON, _ := json.Marshal(policy)
		url := apiService + "/api/policies/" + cr.Name + "/" + policy.Name
		// send policy to api service
		reqLogger.Info("Adding policy " + policy.Name + ", URL: " + url + ", JSON: " + string(policyJSON))
		putRequest(url, string(policyJSON))
	}
}
