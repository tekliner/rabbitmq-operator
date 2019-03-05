package rabbitmq

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileRabbitmq) reconcileEpmdService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {
	labels := map[string]string{
		"rabbitmq.improvado.io/app":       "rabbitmq",
		"rabbitmq.improvado.io/name":      cr.Name,
		"rabbitmq.improvado.io/component": "messaging",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-empd",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.IntOrString{IntVal: 4369},
					Port:       4369,
					Protocol:   corev1.ProtocolTCP,
					Name:       "empd",
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(cr, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("Reconciling Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)

		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(found.Spec.Ports, service.Spec.Ports) {
		found.Spec.Ports = service.Spec.Ports
	}

	if !reflect.DeepEqual(found.Spec.Selector, service.Spec.Selector) {
		found.Spec.Selector = service.Spec.Selector
	}

	if err = r.client.Update(context.TODO(), found); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
func (r *ReconcileRabbitmq) reconcileNodeService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {
	labels := map[string]string{
		"rabbitmq.improvado.io/app":       "rabbitmq",
		"rabbitmq.improvado.io/name":      cr.Name,
		"rabbitmq.improvado.io/component": "messaging",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-node",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.IntOrString{IntVal: 5672},
					Port:       5672,
					Protocol:   corev1.ProtocolTCP,
					Name:       "node",
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cr, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("Reconciling Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)

		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(found.Spec.Ports, service.Spec.Ports) {
		found.Spec.Ports = service.Spec.Ports
	}

	if !reflect.DeepEqual(found.Spec.Selector, service.Spec.Selector) {
		found.Spec.Selector = service.Spec.Selector
	}

	if err = r.client.Update(context.TODO(), found); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
func (r *ReconcileRabbitmq) reconcileHTTPService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	labels := map[string]string{
		"rabbitmq.improvado.io/app":       "rabbitmq",
		"rabbitmq.improvado.io/name":      cr.Name,
		"rabbitmq.improvado.io/component": "messaging",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-api",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.IntOrString{IntVal: 15672},
					Port:       15672,
					Protocol:   corev1.ProtocolTCP,
					Name:       "api",
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cr, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("Reconciling Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)

		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(found.Spec.Ports, service.Spec.Ports) {
		found.Spec.Ports = service.Spec.Ports
	}

	if !reflect.DeepEqual(found.Spec.Selector, service.Spec.Selector) {
		found.Spec.Selector = service.Spec.Selector
	}

	if err = r.client.Update(context.TODO(), found); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
