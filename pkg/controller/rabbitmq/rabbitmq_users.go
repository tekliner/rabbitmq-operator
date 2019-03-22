package rabbitmq

import (
	"context"
	"net"
	"time"

	"github.com/go-logr/logr"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
)

// Like policies, we need to remove all users and add them from secret

func (r *ReconcileRabbitmq) syncUsersCredentials(ctx context.Context, reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secretNames secretResouces) error {
	var secret basicAuthCredentials

	username, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "username")
	secret.username = username
	if err != nil {
		reqLogger.Info("Users: auth username not found")
		return err
	}

	password, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "password")
	secret.password = password
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
	reqLogger.Info("Using API service: "+r.apiServiceAddress(cr), "username", secret.username, "password", secret.password)

	//clean rabbit before fulfilling users list
	reqLogger.Info("Removing all users")

	users, err := r.apiUserList(reqLogger, cr, secret)
	if err != nil {
		reqLogger.Info("Error while receiving users list", "Error", err.Error())
		return err
	}
	reqLogger.Info("Removing all users from list", "Users", users)
	for _, user := range users {
		if user.Name == secret.username {
			// do not delete service account
			continue
		}
		reqLogger.Info("Removing " + user.Name)
		err = r.apiUserRemove(reqLogger, cr, secret, user)
		if err != nil {
			return err
		}
	}

	reqLogger.Info("Uploading users from secret")

	// get secret with users
	credentialsSecret, err := r.getSecret(secretNames.Credentials, cr.Namespace)

	// add new users to Rabbit
	for user, password := range credentialsSecret.Data {
		reqLogger.Info("Adding user " + user + " Password " + string(password))

		err = r.apiUserAdd(reqLogger, cr, secret, rabbitmqUserStruct{Name: user, Password: string(password), Tags: "management"})
		if err != nil {
			reqLogger.Info("Error adding user "+user, "Error", err)
			return err
		}
	}

	return nil
}
