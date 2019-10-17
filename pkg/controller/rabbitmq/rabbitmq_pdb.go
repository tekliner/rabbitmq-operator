package rabbitmq

import (
	"context"
	"github.com/go-logr/logr"
	v1beta1policy "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
)

func (r *ReconcileRabbitmq) reconcilePdb(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	newPdb := getDisruptionBudget(cr)
	reqLogger.Info("Started reconciling PodDisruptionBudget", "Pdb.Namespace", newPdb.Namespace, "Pdb.Name", newPdb.Name)

	if err := controllerutil.SetControllerReference(cr, &newPdb, r.scheme); err != nil {
		reqLogger.Info("Error setting controller reference for PodDisruptionBudget", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		return reconcile.Result{}, err
	}
	// Check existing pdb
	foundPdb := &v1beta1policy.PodDisruptionBudget{}
	reqLogger.Info("Getting PodDisruptionBudget", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: newPdb.Name, Namespace: newPdb.Namespace}, foundPdb)

	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("No PodDisruptionBudget found, creating new", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		err = r.client.Create(context.TODO(), &newPdb)

		foundPdb = &newPdb

		if err != nil {
			reqLogger.Info("Error creating new PodDisruptionBudget", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Info("Error getting PodDisruptionBudget", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(foundPdb.Spec.Selector, &newPdb.Spec.Selector) {
		reqLogger.Info("Selectors not deep equal", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		foundPdb.Spec.Selector = newPdb.Spec.Selector
	}

	if !reflect.DeepEqual(foundPdb.Spec.MaxUnavailable, &newPdb.Spec.MaxUnavailable) {
		reqLogger.Info("MaxUnavailable not deep equal", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		foundPdb.Spec.MaxUnavailable = newPdb.Spec.MaxUnavailable
	}

	if !reflect.DeepEqual(foundPdb.Spec.MinAvailable, &newPdb.Spec.MinAvailable) {
		reqLogger.Info("MinAvailable not deep equal", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		foundPdb.Spec.MaxUnavailable = newPdb.Spec.MaxUnavailable
	}

	if err = r.client.Update(context.TODO(), foundPdb); err != nil {
		reqLogger.Info("Error updating PodDisruptionBudget", "Pdb.Namespace", &newPdb.Namespace, "Pdb.Name", &newPdb.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func getDisruptionBudget(cr *rabbitmqv1.Rabbitmq) v1beta1policy.PodDisruptionBudget {
	podDisruptionBudget := v1beta1policy.PodDisruptionBudget{}
	labelSelector := metav1.LabelSelector{MatchLabels: cr.Labels}
	if cr.Spec.RabbitmqReplicas >= 2 {
		specPDB := v1beta1policy.PodDisruptionBudgetSpec{Selector: &labelSelector}

		if cr.Spec.RabbitmqPdb.Spec.MinAvailable != nil && cr.Spec.RabbitmqPdb.Spec.MaxUnavailable != nil {
			specPDB.MinAvailable = cr.Spec.RabbitmqPdb.Spec.MinAvailable
		} else if cr.Spec.RabbitmqPdb.Spec.MinAvailable != nil {
			specPDB.MinAvailable = cr.Spec.RabbitmqPdb.Spec.MinAvailable
		} else if cr.Spec.RabbitmqPdb.Spec.MaxUnavailable != nil {
			specPDB.MaxUnavailable = cr.Spec.RabbitmqPdb.Spec.MaxUnavailable
		} else {
			specPDB.MinAvailable = func() *intstr.IntOrString { v := intstr.FromInt(1); return &v }()
		}

		podDisruptionBudget = v1beta1policy.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.Name,
				Namespace: cr.Namespace,
			},
			Spec: specPDB,
		}
	} else {
		maxUnavailable := intstr.FromInt(1)
		specPDB := v1beta1policy.PodDisruptionBudgetSpec{
			Selector:       &labelSelector,
			MaxUnavailable: &maxUnavailable,
		}
		podDisruptionBudget = v1beta1policy.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.Name,
				Namespace: cr.Namespace,
			},
			Spec: specPDB,
		}
	}
	return podDisruptionBudget

}
