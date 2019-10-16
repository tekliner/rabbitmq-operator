package rabbitmq

import (
	"context"
	"github.com/go-logr/logr"
	v1beta1policy "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
)

func (r *ReconcileRabbitmq) reconcilePdb(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, pdb *v1beta1policy.PodDisruptionBudget) (reconcile.Result, error) {
	reqLogger.Info("Started reconciling PodDisruptionBudget", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)

	if err := controllerutil.SetControllerReference(cr, pdb, r.scheme); err != nil {
		reqLogger.Info("Error setting controller reference for PodDisruptionBudget", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)
		return reconcile.Result{}, err
	}
	// Check existing pdb
	foundPdb := &v1beta1policy.PodDisruptionBudget{}
	reqLogger.Info("Getting PodDisruptionBudget", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: pdb.Name, Namespace: pdb.Namespace}, foundPdb)

	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("No PodDisruptionBudget found, creating new", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)
		err = r.client.Create(context.TODO(), pdb)

		foundPdb = pdb

		if err != nil {
			reqLogger.Info("Error creating new PodDisruptionBudget", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Info("Error getting PodDisruptionBudget", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(foundPdb.Spec.Selector, pdb.Spec.Selector) {
		reqLogger.Info("Selectors not deep equal", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)
		foundPdb.Spec.Selector = pdb.Spec.Selector
	}

	if !reflect.DeepEqual(foundPdb.Spec.MaxUnavailable, pdb.Spec.MaxUnavailable) {
		reqLogger.Info("MaxUnavailable not deep equal", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)
		foundPdb.Spec.MaxUnavailable = pdb.Spec.MaxUnavailable
	}

	if !reflect.DeepEqual(foundPdb.Spec.MinAvailable, pdb.Spec.MinAvailable) {
		reqLogger.Info("MinAvailable not deep equal", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)
		foundPdb.Spec.MaxUnavailable = pdb.Spec.MaxUnavailable
	}

	if err = r.client.Update(context.TODO(), foundPdb); err != nil {
		reqLogger.Info("Error updating PodDisruptionBudget", "Pdb.Namespace", pdb.Namespace, "Pdb.Name", pdb.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
