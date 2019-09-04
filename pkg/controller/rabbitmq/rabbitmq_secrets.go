package rabbitmq

import (
	"context"
	"encoding/base64"
	"math/rand"

	"github.com/go-logr/logr"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func secretEncode(plaintextSecret string) string {
	return base64.StdEncoding.EncodeToString([]byte(plaintextSecret))
}
func secretDecode(encodedSecret []byte) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(encodedSecret))
	return string(decoded), err
}

func (r *ReconcileRabbitmq) getSecret(name string, namespace string) (corev1.Secret, error) {
	secretObj := corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, &secretObj)
	return secretObj, err
}

func (r *ReconcileRabbitmq) getSecretData(reqLogger logr.Logger, secretNamespace string, secretName string, secretDataField string) (string, error) {
	secretObj := corev1.Secret{}
	reqLogger.Info("Get secret request", "secret namespace", secretNamespace, "secret name", secretName, "secret data field", secretDataField)
	secretObj, err := r.getSecret(secretName, secretNamespace)
	if err != nil {
		reqLogger.Info("Get secret request ERROR", "secret namespace", secretNamespace, "secret name", secretName, "secret data field", secretDataField)
		return "", err
	}
	return string(secretObj.Data[secretDataField]), nil
}

func (r *ReconcileRabbitmq) reconcileSecrets(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (secretResouces, error) {

	var secretNames secretResouces

	// standart resource names
	secretNames.ServiceAccount = cr.Name + "-service-account"
	secretNames.Credentials = cr.Name + "-credentials"

	/////////////////////////////////////////////////////////////////
	// SERVICE ACCOUNT SECRET, CREATING ONCE AT RABBIT INSTALL
	/////////////////////////////////////////////////////////////////

	// check existance of linked or standart service account secret
	createServiceAccount := false
	if cr.Spec.RabbitmqSecretServiceAccount != "" {
		secretNames.ServiceAccount = cr.Spec.RabbitmqSecretServiceAccount
		secretSAResource := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretNames.ServiceAccount, Namespace: cr.Namespace}, secretSAResource)

		if err != nil && apierrors.IsNotFound(err) {
			// not found! create new service account
			reqLogger.Info("Service Account: linked resource not found, operator will create new")
			createServiceAccount = true
		} else if err != nil {
			// happend something else, do nothing
			reqLogger.Info("Service Account linked: error happend", err)
			return secretResouces{}, err
		}
	} else {
		// link empty, search standart service account secret
		reqLogger.Info("Service Account: search for standart resource")
		secretSAResource := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretNames.ServiceAccount, Namespace: cr.Namespace}, secretSAResource)

		if err != nil && apierrors.IsNotFound(err) {
			// not found! create new service account
			reqLogger.Info("Service Account: standart resource not found, operator will create new")
			createServiceAccount = true
		} else if err != nil {
			// happend something else, do nothing
			reqLogger.Info("Service Account standart: error happend", err)
			return secretResouces{}, err
		}
	}

	// create service account with standart name
	if createServiceAccount {
		reqLogger.Info("Creating new service account secret", "Namespace", cr.Namespace, "Name", secretNames.ServiceAccount)
		secretSAResource := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretNames.ServiceAccount,
				Namespace: cr.Namespace,
				Labels:    returnLabels(cr),
			},
			Data: map[string][]byte{
				"username": []byte("sa"),
				"password": []byte(randomString(30)),
				"cookie":   []byte(randomString(30)),
			},
		}

		if err := controllerutil.SetControllerReference(cr, secretSAResource, r.scheme); err != nil {
			return secretResouces{}, err
		}

		err := r.client.Create(context.TODO(), secretSAResource)

		if err != nil {
			return secretResouces{}, err
		}

	}

	/////////////////////////////////////////////////////////////////
	// USER CREDENTIALS SECRET
	/////////////////////////////////////////////////////////////////

	// check existance of linked or standart credentials secret
	// if resource found remove make lists to add, change or remove users
	createCredentialsSecret := false
	secretCredResource := &corev1.Secret{}
	if cr.Spec.RabbitmqSecretCredentials != "" {
		// try to find secret under name from CRD
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Spec.RabbitmqSecretCredentials, Namespace: cr.Namespace}, secretCredResource)
		if err != nil && apierrors.IsNotFound(err) {
			// not found! create new service account
			reqLogger.Info("User credentials: linked resource not found, operator will create new")
			createCredentialsSecret = true
		} else if err != nil {
			// happend something else, do nothing
			reqLogger.Info("User credentials: error happend", err)
			return secretResouces{}, err
		}
		secretNames.Credentials = cr.Spec.RabbitmqSecretCredentials
	} else {
		// link empty, search standart credentials secret name
		reqLogger.Info("User credentials: search for standart resource")
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretNames.Credentials, Namespace: cr.Namespace}, secretCredResource)

		if err != nil && apierrors.IsNotFound(err) {
			// not found! create new service account
			reqLogger.Info("User credentials: standart resource not found, operator will create new")
			createCredentialsSecret = true
		} else if err != nil {
			// happend something else, do nothing
			reqLogger.Info("User credentials standart: error happend", err)
			return secretResouces{}, err
		}
	}

	// create credentials secret
	if createCredentialsSecret {
		reqLogger.Info("Creating new user credentials secret", "Namespace", cr.Namespace, "Name", secretNames.Credentials)
		secretCredResourceNew := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretNames.Credentials,
				Namespace: cr.Namespace,
				Labels:    returnLabels(cr),
			},
			Data: map[string][]byte{},
		}

		if err := controllerutil.SetControllerReference(cr, secretCredResourceNew, r.scheme); err != nil {
			return secretResouces{}, err
		}

		err := r.client.Create(context.TODO(), secretCredResourceNew)
		if err != nil {
			return secretResouces{}, err
		}

	}

	return secretNames, nil
}

func randomString(l int) string {
	var letterRunes = []rune("ABCDEFGHIJKLMNOabcdefghijklmn67890PQRSTUVWXYZ12345opqrstuvwxyz")
	b := make([]rune, l)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
