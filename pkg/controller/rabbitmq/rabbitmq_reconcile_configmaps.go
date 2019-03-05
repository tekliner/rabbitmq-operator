package rabbitmq

import (
	"context"

	"github.com/go-logr/logr"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const defaultRabbitmqConfig = `
#
`

func (r *ReconcileRabbitmq) reconcileConfigMap(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {
	labels := map[string]string{
		"rabbitmq.improvado.io/app":       "rabbitmq",
		"rabbitmq.improvado.io/name":      cr.Name,
		"rabbitmq.improvado.io/component": "messaging",
	}

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"dsas.conf": defaultRabbitmqConfig,
		},
	}

	if err := controllerutil.SetControllerReference(cr, config, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: config.Name, Namespace: config.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("Reconciling config", "config.Namespace", config.Namespace, "config.Name", config.Name)
		err = r.client.Create(context.TODO(), config)

		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if err = r.client.Update(context.TODO(), found); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
