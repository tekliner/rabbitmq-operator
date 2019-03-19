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

func (r *ReconcileRabbitmq) reconcileService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, service *corev1.Service) (reconcile.Result, error) {
	reqLogger.Info("Started reconciling service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)

	if err := controllerutil.SetControllerReference(cr, service, r.scheme); err != nil {
		reqLogger.Info("Error setting controller reference for service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Service{}
	reqLogger.Info("Getting service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("No service found, creating new", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)

		found = service

		if err != nil {
			reqLogger.Info("Error creating new service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Info("Error getting service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(found.Spec.Ports, service.Spec.Ports) {
		reqLogger.Info("Ports not deep equal", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		found.Spec.Ports = service.Spec.Ports
	}

	if !reflect.DeepEqual(found.Spec.Selector, service.Spec.Selector) {
		reqLogger.Info("Selector not deep equal", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		found.Spec.Selector = service.Spec.Selector
	}

	if err = r.client.Update(context.TODO(), found); err != nil {
		reqLogger.Info("Error updating service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileRabbitmq) reconcileDiscoveryService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-discovery",
			Namespace: cr.Namespace,
			Labels:    returnLabels(cr),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: returnLabels(cr),
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.IntOrString{IntVal: 5672},
					Port:       5672,
					Protocol:   corev1.ProtocolTCP,
					Name:       "amqp",
				},
				{
					TargetPort: intstr.IntOrString{IntVal: 4369},
					Port:       4369,
					Protocol:   corev1.ProtocolTCP,
					Name:       "empd",
				},
			},
		},
	}

	reconcileResult, err := r.reconcileService(reqLogger, cr, service)

	return reconcileResult, err
}

func (r *ReconcileRabbitmq) reconcileEpmdService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-empd",
			Namespace: cr.Namespace,
			Labels:    returnLabels(cr),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: returnLabels(cr),
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
	reconcileResult, err := r.reconcileService(reqLogger, cr, service)

	return reconcileResult, err
}

func (r *ReconcileRabbitmq) reconcileAmqpService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-amqp",
			Namespace: cr.Namespace,
			Labels:    returnLabels(cr),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: returnLabels(cr),
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.IntOrString{IntVal: 5672},
					Port:       5672,
					Protocol:   corev1.ProtocolTCP,
					Name:       "amqp",
				},
			},
		},
	}

	reconcileResult, err := r.reconcileService(reqLogger, cr, service)

	return reconcileResult, err
}
func (r *ReconcileRabbitmq) reconcileHTTPService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-api",
			Namespace: cr.Namespace,
			Labels:    returnLabels(cr),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: returnLabels(cr),
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

	reconcileResult, err := r.reconcileService(reqLogger, cr, service)

	return reconcileResult, err
}
