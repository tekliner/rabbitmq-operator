package rabbitmq

import (
	"context"

	"github.com/go-logr/logr"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const RabbitmqPVCFinalizer = "rabbitmq.improvado.io/pvc"

func (r *ReconcileRabbitmq) deleteDependentResoucePVC(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {
	reqLogger.Info("Deleting PVC")
	// Finalizer PVC: remove finalizer from CR
	if containsString(cr.ObjectMeta.Finalizers, "PVC") {
		// get dependent PVCs
		foundDependentPVCs := &corev1.PersistentVolumeClaimList{}

		err := r.client.List(context.TODO(), client.InNamespace(cr.Namespace).MatchingLabels(mergeMaps(returnLabels(cr))), foundDependentPVCs)
		if err != nil {
			reqLogger.Info("Listing dependent PVCs error", "Namespace", cr.Namespace, ".Name", cr.Name)
			return reconcile.Result{}, err
		}

		for _, pvc := range foundDependentPVCs.Items {
			// delete dependent PVCs
			if err := r.client.Delete(context.Background(), &pvc); err != nil {
				reqLogger.Info("Deleting PVC failed")
				return reconcile.Result{}, err
			}
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileRabbitmq) reconcileFinalizers(reqLogger logr.Logger, instance *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {
	reqLogger.Info("Processing finalizers", "Namespace", instance.Namespace, ".Name", instance.Name)
	// Define finalizer strings to prevent deletion of CR before dependent resources deletion
	finalizersList := []string{RabbitmqPVCFinalizer}

	// Check CR is being deleted or not
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Its a new CR, add finalizers from list
		instance.ObjectMeta.Finalizers = finalizersList
		if err := r.client.Update(context.Background(), instance); err != nil {
			return reconcile.Result{}, err
		}

	} else {
		// The object is being deleted, removing dependencies and finalizers after it

		// Finalizer "PVC": check to remove dependent PVCs
		if instance.Spec.RabbitmqPurgePVC {
			_, err := r.deleteDependentResoucePVC(reqLogger, instance)
			if err != nil {
				reqLogger.Info("PVC deletion error", "Namespace", instance.Namespace, ".Name", instance.Name)
				return reconcile.Result{}, err
			}
		}

		// Finalizer "PVC": remove "PVC" finalizer from CR
		instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, RabbitmqPVCFinalizer)
		if err := r.client.Update(context.Background(), instance); err != nil {
			reqLogger.Info("Removing PVC finalizer failed")
			return reconcile.Result{}, err
		}

		// Our finalizer has finished, so the reconciler can do nothing.
	}

	return reconcile.Result{}, nil
}
