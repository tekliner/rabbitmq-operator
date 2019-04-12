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

	// get service account credentials
	var serviceAccount basicAuthCredentials

	username, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "username")
	serviceAccount.username = username
	if err != nil {
		reqLogger.Info("Users: auth username not found")
		return err
	}

	password, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "password")
	serviceAccount.password = password
	if err != nil {
		reqLogger.Info("Users: auth password not found")
		return err
	}

	// wait http connection to api port
	timeout := time.Duration(5 * time.Second)

	_, err = net.DialTimeout("tcp", r.apiServiceHostname(cr), timeout)
	if err != nil {
		reqLogger.Info("Rabbitmq API service failed", "Service name", r.apiServiceHostname(cr), "Error", err.Error())
		return err
	}
	reqLogger.Info("Policies: Using API service: "+r.apiServiceAddress(cr), "username", serviceAccount.username, "password", serviceAccount.password)

	var policiesCR []rabbitmqv1.RabbitmqPolicy

	// get exiting policies
	reqLogger.Info("Reading exiting policies")

	policiesRabbit, err := r.apiPolicyList(reqLogger, cr, serviceAccount)
	if err != nil {
		reqLogger.Info("Error while receiving policies list", "Error", err.Error())
		return err
	}

	// get policies from CR
	reqLogger.Info("Reading policies from CRD")

	// set default vhost for all policies
	policiesDefaultVhost := "%2f"
	if cr.Spec.RabbitmqVhost != "" {
		policiesDefaultVhost = cr.Spec.RabbitmqVhost
	}

	// detect vhost to use
	for _, policy := range cr.Spec.RabbitmqPolicies {
		// detect vhost to use
		policyVhost := ""
		if policy.Vhost != "" {
			policyVhost = policy.Vhost
		} else {
			policyVhost = policiesDefaultVhost
		}

		policy.Vhost = policyVhost

		policiesCR = append(policiesCR, policy)

	}

	// ok, now syncing

	// remove policies from rabbit
	for _, policyRabbit := range policiesRabbit {

		// search
		policyFound := false
		for _, policyCR := range policiesCR {
			if policyCR.Name == policyRabbit.Name {
				policyFound =true
			}
		}

		if !policyFound {
			reqLogger.Info("Removing " + policyRabbit.Name)
			err = r.apiPolicyRemove(reqLogger, cr, serviceAccount, policyRabbit.Vhost, policyRabbit.Name)
			if err != nil {
				return err
			}
		}

	}

	// add to rabbit from CR
	for _, policyCR := range policiesCR {
		reqLogger.Info("Adding policy " + policyCR.Name + " to vhost " + policyCR.Vhost)
		err = r.apiPolicyAdd(reqLogger, cr, serviceAccount, policyCR.Vhost, policyCR)
		if err != nil {
			reqLogger.Info("Error adding policy "+policyCR.Name+" to vhost "+policyCR.Vhost, "Error", err)
			return err
		}
	}

	return nil
}
