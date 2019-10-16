package rabbitmq

import (
	"github.com/go-logr/logr"
	v1beta1policy "k8s.io/api/policy/v1beta1"
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

	return reconcile.Result{}, nil
}
