package rabbitmq

import (
	"context"
	"net"
	"time"

	"github.com/go-logr/logr"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
)

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

	_, err = net.DialTimeout("tcp", r.apiServiceHostname(cr), timeout)
	if err != nil {
		reqLogger.Info("Rabbitmq API service failed", "Service name", r.apiServiceHostname(cr), "Error", err.Error())
		return err
	}
	reqLogger.Info("Using API service: "+r.apiServiceAddress(cr), "username", secret.username, "password", secret.password)

	//clean rabbit before fulfilling policies list
	reqLogger.Info("Removing all policies")

	policies, err := r.apiPolicyList(reqLogger, cr, secret)
	if err != nil {
		reqLogger.Info("Error while receiving policies list", "Error", err.Error())
		return err
	}
	reqLogger.Info("Removing all policies from list", "Policies", policies)
	for _, policyRecord := range policies {
		reqLogger.Info("Removing " + policyRecord.Name)
		err = r.apiPolicyRemove(reqLogger, cr, secret, policyRecord.Vhost, policyRecord.Name)
		if err != nil {
			return err
		}
	}

	reqLogger.Info("Uploading policies from CRD")

	// detect default vhost for all policies
	policiesDefaultVhost := "%2f"
	if cr.Spec.RabbitmqVhost != "" {
		policiesDefaultVhost = cr.Spec.RabbitmqVhost
	}

	// add new policies to Rabbit
	for _, policy := range cr.Spec.RabbitmqPolicies {
		// detect vhost to use
		policyVhost := ""
		if policy.Vhost != "" {
			policyVhost = policy.Vhost
		} else {
			policyVhost = policiesDefaultVhost
		}

		// send policy to api service
		reqLogger.Info("Adding policy " + policy.Name + " to vhost " + policyVhost)
		err = r.apiPolicyAdd(reqLogger, cr, secret, policyVhost, policy)
		if err != nil {
			reqLogger.Info("Error adding policy "+policy.Name+" to vhost "+policyVhost, "Error", err)
			return err
		}
	}

	return nil
}
