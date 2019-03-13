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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

func (r *ReconcileRabbitmq) reconcileSecrets(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	labels := map[string]string{
		"rabbitmq.improvado.io/app":       "rabbitmq",
		"rabbitmq.improvado.io/name":      cr.Name,
		"rabbitmq.improvado.io/component": "messaging",
	}

	// standart resource names
	secretNameSA := cr.Name + "-service-account"
	secretNameCredentials := cr.Name + "-credentials"

	/////////////////////////////////////////////////////////////////
	// SERVICE ACCOUNT SECRET, CREATING ONCE AT RABBIT INSTALL
	/////////////////////////////////////////////////////////////////

	// check existance of linked or standart service account secret
	createServiceAccount := false
	if cr.Spec.RabbitmqSecretServiceAccount != "" {
		secretSAResource := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Spec.RabbitmqSecretServiceAccount, Namespace: cr.Namespace}, secretSAResource)

		if err != nil && apierrors.IsNotFound(err) {
			// not found! create new service account
			reqLogger.Info("Service Account: linked resource not found, operator will create new")
			createServiceAccount = true
		} else if err != nil {
			// happend something else, do nothing
			reqLogger.Info("Service Account linked: error happend", err)
			return reconcile.Result{}, err
		}
		secretNameSA = cr.Spec.RabbitmqSecretServiceAccount
	} else {
		// link empty, search standart service account secret
		reqLogger.Info("Service Account: search for standart resource")
		secretSAResource := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretNameSA, Namespace: cr.Namespace}, secretSAResource)

		if err != nil && apierrors.IsNotFound(err) {
			// not found! create new service account
			reqLogger.Info("Service Account: standart resource not found, operator will create new")
			createServiceAccount = true
		} else if err != nil {
			// happend something else, do nothing
			reqLogger.Info("Service Account standart: error happend", err)
			return reconcile.Result{}, err
		}
	}

	// create service account with standart name
	if createServiceAccount {
		reqLogger.Info("Creating new service account secret", "Namespace", cr.Namespace, "Name", secretNameSA)
		secretSAResource := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretNameSA,
				Namespace: cr.Namespace,
				Labels:    labels,
			},
			Data: map[string][]byte{
				"username": []byte("sa"),
				"password": []byte(secretEncode(randomString(30))),
			},
		}

		if err := controllerutil.SetControllerReference(cr, secretSAResource, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		err := r.client.Create(context.TODO(), secretSAResource)

		if err != nil {
			return reconcile.Result{}, err
		}

	}

	/////////////////////////////////////////////////////////////////
	// USER CREDENTIALS SECRET
	/////////////////////////////////////////////////////////////////

	// check existance of linked or standart credentials secret
	// if resource found remove make lists to add, change or remove users
	createCredentialsSecret := false
	if cr.Spec.RabbitmqSecretCredentials != "" {
		secretCredResource := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Spec.RabbitmqSecretCredentials, Namespace: cr.Namespace}, secretCredResource)

		if err != nil && apierrors.IsNotFound(err) {
			// not found! create new service account
			reqLogger.Info("User credentials: linked resource not found, operator will create new")
			createCredentialsSecret = true
		} else if err != nil {
			// happend something else, do nothing
			reqLogger.Info("User credentials: error happend", err)
			return reconcile.Result{}, err
		}
		secretNameCredentials = cr.Spec.RabbitmqSecretCredentials
	} else {
		// link empty, search standart credentials secret
		reqLogger.Info("User credentials: search for standart resource")
		secretCredResource := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretNameCredentials, Namespace: cr.Namespace}, secretCredResource)

		if err != nil && apierrors.IsNotFound(err) {
			// not found! create new service account
			reqLogger.Info("User credentials: standart resource not found, operator will create new")
			createCredentialsSecret = true
		} else if err != nil {
			// happend something else, do nothing
			reqLogger.Info("User credentials standart: error happend", err)
			return reconcile.Result{}, err
		}
	}

	// create credentials secret
	if createCredentialsSecret {
		reqLogger.Info("Creating new user credentials secret", "Namespace", cr.Namespace, "Name", secretNameCredentials)
		secretCredResource := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretNameCredentials,
				Namespace: cr.Namespace,
				Labels:    labels,
			},
			Data: map[string][]byte{},
		}

		if err := controllerutil.SetControllerReference(cr, secretCredResource, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		err := r.client.Create(context.TODO(), secretCredResource)
		if err != nil {
			return reconcile.Result{}, err
		}

	}

	// // try to fetch exiting credentials secret
	// exitingCredSecret := &corev1.Secret{}
	// err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretNameCredentials, Namespace: cr.Namespace}, exitingCredSecret)
	// if err != nil {
	// 	reqLogger.Info("Something went terribly wrong! Credentials secret not found!", "Secret name", secretNameCredentials, err)
	// 	return reconcile.Result{}, err
	// }

	// // lets ensure that crd have same data as secret
	// if !reflect.DeepEqual(exitingCredSecret.Data, cr.Spec.RabbitmqSecretCredentials) {
	// 	apiService := "http://" + cr.Name + "-api:15672"

	// 	exitingUsers := found.Data
	// 	configmapUsers := cr.Spec.RabbitmqSecretCredentials

	// 	// remove all exiting users
	// 	for user in range exitingUsers {
	// 		url := apiService + "/api/users/" + user

	// 	}

	// 	// send api request to set new credentials

	// 	// ok, now equalizing data
	// 	found.Data = rabbitSecret.Data
	// }

	// if err = r.client.Update(context.TODO(), found); err != nil {
	// 	return reconcile.Result{}, err
	// }

	return reconcile.Result{}, nil
}

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}

func searchDifference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}
