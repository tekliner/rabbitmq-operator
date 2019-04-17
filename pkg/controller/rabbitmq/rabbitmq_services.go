package rabbitmq

import (
	"context"
	v12 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
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

func (r *ReconcileRabbitmq) reconcileServiceMonitor(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, serviceMonitorResource v12.ServiceMonitor) (reconcile.Result, error) {
	reqLogger.Info("Started reconciling ServiceMonitor", "ServiceMonitor.Namespace", serviceMonitorResource.Namespace, "ServiceMonitor.Name", serviceMonitorResource.Name)

	// Check if this ServiceMonitor already exists
	found := v12.ServiceMonitor{}
	reqLogger.Info("Getting ServiceMonitor", "Service.Namespace", serviceMonitorResource.Namespace, "Service.Name", serviceMonitorResource.Name)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: serviceMonitorResource.Name, Namespace: serviceMonitorResource.Namespace}, &found)

	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("No ServiceMonitor found, creating new", "Service.Namespace", serviceMonitorResource.Namespace, "Service.Name", serviceMonitorResource.Name)
		err = r.client.Create(context.TODO(), &serviceMonitorResource)
		if err != nil {
			reqLogger.Info("Error creating new service", "Service.Namespace", serviceMonitorResource.Namespace, "Service.Name", serviceMonitorResource.Name)
			return reconcile.Result{}, err
		}

	} else if err != nil {
		reqLogger.Info("Error getting ServiceMonitor", "Service.Namespace", serviceMonitorResource.Namespace, "Service.Name", serviceMonitorResource.Name)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}


func (r *ReconcileRabbitmq) reconcileDiscoveryService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-discovery",
			Namespace: cr.Namespace,
			Labels:    mergeMaps(returnLabels(cr),
				map[string]string{"service": "discovery"},
				map[string]string{"component": "networking"},
			),
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

func (r *ReconcileRabbitmq) reconcileHAService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    mergeMaps(returnLabels(cr),
				map[string]string{"service": "general"},
				map[string]string{"component": "networking"},
			),
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

func (r *ReconcileRabbitmq) reconcileHTTPService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-api",
			Namespace: cr.Namespace,
			Labels:    mergeMaps(returnLabels(cr),
				map[string]string{"service": "api"},
				map[string]string{"component": "networking"},
			),
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

func (r *ReconcileRabbitmq) reconcilePrometheusExporterService(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-exporter",
			Namespace: cr.Namespace,
			Labels:    mergeMaps(returnLabels(cr),
				map[string]string{"service": "prometheus-exporter"},
				map[string]string{"component": "monitoring"},
				map[string]string{"component": "networking"},
			),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: returnLabels(cr),
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.IntOrString{IntVal: cr.Spec.RabbitmqPrometheusExporterPort},
					Port:       cr.Spec.RabbitmqPrometheusExporterPort,
					Protocol:   corev1.ProtocolTCP,
					Name:       "exporter",
				},
			},
		},
	}

	reconcileResult, err := r.reconcileService(reqLogger, cr, service)

	return reconcileResult, err
}

func (r *ReconcileRabbitmq) reconcilePrometheusExporterServiceMonitor(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {

	serviceMonitor := v12.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    mergeMaps(returnLabels(cr), map[string]string{"rabbitmq.improvado.io/component": "monitoring"}),
		},
		Spec: v12.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: mergeMaps(returnLabels(cr),
					map[string]string{"service": "prometheus-exporter"},
					map[string]string{"component": "monitoring"},
				),
			},
			Endpoints: []v12.Endpoint{
				{
					Port: "exporter",
					Interval: "10s",
				},
			},
			NamespaceSelector: v12.NamespaceSelector{
				Any:true,
			},
		},
	}

	reconcileResult, err := r.reconcileServiceMonitor(reqLogger, cr, serviceMonitor)

	return reconcileResult, err
}
