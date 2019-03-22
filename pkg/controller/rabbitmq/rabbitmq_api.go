package rabbitmq

import (
	"encoding/json"

	"github.com/go-logr/logr"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
)

func (r *ReconcileRabbitmq) apiServiceAddress(cr *rabbitmqv1.Rabbitmq) string {
	apiService := "http://" + r.apiServiceHostname(cr)
	return apiService
}

func (r *ReconcileRabbitmq) apiServiceHostname(cr *rabbitmqv1.Rabbitmq) string {
	apiHostname := cr.Name + "-api." + cr.Namespace + ":15672"
	return apiHostname
}

// Policies block

func (r *ReconcileRabbitmq) apiPolicyList(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secret basicAuthCredentials) ([]rabbitmqv1.RabbitmqPolicy, error) {
	url := r.apiServiceAddress(cr) + "/api/policies"

	response, err := getRequest(url, secret)
	if err != nil {
		reqLogger.Info("Error while receiving policies list", "Error", err)
		return nil, err
	}

	var policies []rabbitmqv1.RabbitmqPolicy
	err = json.Unmarshal(response, &policies)
	if err != nil {
		// something bad
		reqLogger.Info("Error parsing json!", "Error", err, "Data", string(response))
		return nil, err
	}

	return policies, nil
}

func (r *ReconcileRabbitmq) apiPolicyAdd(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secret basicAuthCredentials, policyVHost string, policyData rabbitmqv1.RabbitmqPolicy) error {
	reqLogger.Info("Adding " + policyData.Name + " to " + policyVHost + " vhost")
	policyJSON, _ := json.Marshal(policyData)
	url := r.apiServiceAddress(cr) + "/api/policies/" + policyVHost + "/" + policyData.Name
	err := putRequest(url, secret, string(policyJSON))
	if err != nil {
		return err
	}
	return nil
}

func (r *ReconcileRabbitmq) apiPolicyRemove(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secret basicAuthCredentials, policyVHost string, policyName string) error {
	reqLogger.Info("Removing " + policyName + " from " + policyVHost + " vhost")
	url := r.apiServiceAddress(cr) + "/api/policies/" + policyVHost + "/" + policyName
	err := deleteRequest(url, secret, "")
	if err != nil {
		return err
	}
	return nil
}

// Users block

func (r *ReconcileRabbitmq) apiUserList(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secret basicAuthCredentials) ([]rabbitmqUserStruct, error) {
	url := r.apiServiceAddress(cr) + "/api/users"

	response, err := getRequest(url, secret)
	if err != nil {
		reqLogger.Info("Error while receiving user list", "Error", err)
		return nil, err
	}

	var users []rabbitmqUserStruct
	err = json.Unmarshal(response, &users)
	if err != nil {
		// something bad
		reqLogger.Info("Error parsing json!", "Error", err, "Data", string(response))
		return nil, err
	}

	return users, nil
}

func (r *ReconcileRabbitmq) apiUserAdd(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secret basicAuthCredentials, user rabbitmqUserStruct) error {
	userJSON, err := json.Marshal(user)
	reqLogger.Info("Adding "+user.Name+" with tag: "+user.Tags, "JSON", string(userJSON))
	if err != nil {
		reqLogger.Info("Error while marshaling to JSON", "User", user.Name, "Error", err)
	}
	url := r.apiServiceAddress(cr) + "/api/users/" + user.Name
	err = putRequest(url, secret, string(userJSON))
	if err != nil {
		reqLogger.Info("Error while putRequest", "User", user.Name, "Error", err)
		return err
	}
	return nil
}

func (r *ReconcileRabbitmq) apiUserRemove(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secret basicAuthCredentials, user rabbitmqUserStruct) error {
	reqLogger.Info("Remove user " + user.Name)
	url := r.apiServiceAddress(cr) + "/api/users/" + user.Name
	err := deleteRequest(url, secret, "")
	if err != nil {
		return err
	}
	return nil
}

func (r *ReconcileRabbitmq) apiUserBulkRemove(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secret basicAuthCredentials, users []string) error {
	reqLogger.Info("Remove users", "bulk", users)
	usersJSON, _ := json.Marshal(rabbitmqUsersListStruct{Users: users})
	url := r.apiServiceAddress(cr) + "/api/users/bulk-delete"
	err := deleteRequest(url, secret, string(usersJSON))
	if err != nil {
		return err
	}
	return nil
}
