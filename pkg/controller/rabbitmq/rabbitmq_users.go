package rabbitmq

import (
	"context"
	"net"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/go-logr/logr"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
)

// Like policies, we need to remove all users and add them from secret

func (r *ReconcileRabbitmq) syncUsersCredentials(ctx context.Context, reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secretNames secretResouces) error {

	// get service account credentials
	var serviceAccount basicAuthCredentials

	username, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "username")
	serviceAccount.username = username
	if err != nil {
		reqLogger.Info("Users: auth username not found")
		raven.CaptureErrorAndWait(err, nil)
		return err
	}

	password, err := r.getSecretData(reqLogger, cr.Namespace, secretNames.ServiceAccount, "password")
	serviceAccount.password = password
	if err != nil {
		reqLogger.Info("Users: auth password not found")
		raven.CaptureErrorAndWait(err, nil)
		return err
	}

	// wait http connection to api port
	timeout := time.Duration(5 * time.Second)

	_, err = net.DialTimeout("tcp", r.apiServiceHostname(cr), timeout)
	if err != nil {
		reqLogger.Info("Rabbitmq API service failed", "Service name", r.apiServiceHostname(cr), "Error", err.Error())
		raven.CaptureErrorAndWait(err, nil)
		return err
	}
	reqLogger.Info("Users: Using API service: "+r.apiServiceAddress(cr), "username", serviceAccount.username)

	// get user from secret
	usersSecret, err := r.getSecret(secretNames.Credentials, cr.Namespace)
	reqLogger.Info("Users from secret", "CRD", cr.Name, "SecretNames", secretNames, "Credentials secret", usersSecret.Name, "ServiceAccount", serviceAccount.username)

	// get users from rabbit api
	reqLogger.Info("Reading all users from rabbitmq")
	usersRabbit, err := r.apiUserList(reqLogger, cr, serviceAccount)
	if err != nil {
		reqLogger.Info("Error while receiving users list", "Error", err.Error())
		raven.CaptureErrorAndWait(err, nil)
		return err
	}

	reqLogger.Info("Sync users started")

	// search users to remove
	for _, userRabbitName := range usersRabbit {

		userFound := false

		for userSecretName, _ := range usersSecret.Data {
			if userSecretName == userRabbitName.Name {
				userFound = true
			}
		}

		// user from RabbitMQ not found in secret resource, so add to remove list
		if (!userFound) && (userRabbitName.Name != serviceAccount.username) {
			reqLogger.Info("Removing " + userRabbitName.Name)
			err = r.apiUserRemove(reqLogger, cr, serviceAccount, rabbitmqUserStruct{Name: userRabbitName.Name})
			if err != nil {
				raven.CaptureErrorAndWait(err, nil)
				return err
			}
		}
	}

	reqLogger.Info("Uploading users from secret")

	// add new users to Rabbit
	for userName, userPassword := range usersSecret.Data {
		reqLogger.Info("Adding user " + userName)

		err = r.apiUserAdd(reqLogger, cr, serviceAccount, rabbitmqUserStruct{Name: userName, Password: string(userPassword), Tags: "administrator"})
		if err != nil {
			reqLogger.Info("Error adding user "+userName, "Error", err)
			raven.CaptureErrorAndWait(err, nil)
			return err
		}
	}

	return nil
}
