package rabbitmq

import (
	"github.com/go-logr/logr"
	"k8s.io/api/policy/v1beta1"
	"reflect"
)

func reconcilePdb(reqLogger logr.Logger, foundPdb v1beta1.PodDisruptionBudget, newPdb v1beta1.PodDisruptionBudget) (bool, v1beta1.PodDisruptionBudget) {

	reconcileRequired := false
	if !reflect.DeepEqual(foundPdb.Spec.Selector, newPdb.Spec.Selector) {
		reqLogger.Info("Selectors not deep equal", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		foundPdb.Spec.Selector = newPdb.Spec.Selector
		reconcileRequired = true
	}

	if !reflect.DeepEqual(foundPdb.Spec.MaxUnavailable, newPdb.Spec.MaxUnavailable) {
		reqLogger.Info("MaxUnavailable not deep equal", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		foundPdb.Spec.MaxUnavailable = newPdb.Spec.MaxUnavailable
		reconcileRequired = true
	}

	if !reflect.DeepEqual(foundPdb.Spec.MinAvailable, newPdb.Spec.MinAvailable) {
		reqLogger.Info("MinAvailable not deep equal", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		foundPdb.Spec.MinAvailable = newPdb.Spec.MinAvailable
		reconcileRequired = true
	}
	return reconcileRequired, foundPdb
}
