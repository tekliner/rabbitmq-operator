package rabbitmq

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/getsentry/raven-go"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func init() {
	sentryDSN := os.Getenv("SENTRY_DSN")
	if sentryDSN != "" {
		raven.SetDSN(sentryDSN)
	}
}

var log = logf.Log.WithName("controller_rabbitmq")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Rabbitmq Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRabbitmq{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, reconciler reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rabbitmq-controller", mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return err
	}

	// Watch for changes to primary resource Rabbitmq
	err = c.Watch(&source.Kind{Type: &rabbitmqv1.Rabbitmq{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Rabbitmq
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rabbitmqv1.Rabbitmq{},
	})
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return err
	}

	// Watch for changes to secondary resource StatefulSets and requeue the owner
	err = c.Watch(&source.Kind{Type: &v1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		OwnerType: &rabbitmqv1.Rabbitmq{},
	})
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return err
	}

	mapFn := handler.ToRequestsFunc(
		func(a handler.MapObject) []reconcile.Request {
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{Name: a.Meta.GetLabels()["rabbitmq.improvado.io/instance"], Namespace: a.Meta.GetNamespace()}},
			}
		})

	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			if _, ok := e.MetaOld.GetLabels()["rabbitmq.improvado.io/instance"]; !ok {
				return false
			}
			return e.ObjectOld != e.ObjectNew
		},
		CreateFunc: func(e event.CreateEvent) bool {
			if _, ok := e.Meta.GetLabels()["rabbitmq.improvado.io/instance"]; !ok {
				return false
			}
			return true
		},
	}

	err = c.Watch(
		&source.Kind{Type: &corev1.Secret{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: mapFn,
		}, p)

	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRabbitmq{}

// ReconcileRabbitmq reconciles a Rabbitmq object
type ReconcileRabbitmq struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Rabbitmq object and makes changes based on the state read
// and what is in the Rabbitmq.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.

func mergeMaps(itermaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, rv := range itermaps {
		for k, v := range rv {
			result[k] = v
		}
	}
	return result
}

func returnLabels(cr *rabbitmqv1.Rabbitmq) map[string]string {
	labels := map[string]string{
		"rabbitmq.improvado.io/instance":      cr.Name,
	}
	return labels
}

func returnAnnotationsPrometheus(cr *rabbitmqv1.Rabbitmq) map[string]string {
	return map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   strconv.Itoa(int(cr.Spec.RabbitmqPrometheusExporterPort)),
	}
}

func returnAnnotations(cr *rabbitmqv1.Rabbitmq) map[string]string {
	annotations := map[string]string{}
	if cr.Spec.RabbitmqPrometheusExporterPort > 0 {
		annotations = mergeMaps(annotations, returnAnnotationsPrometheus(cr))
	}
	return annotations
}

// Reconcile method
func (r *ReconcileRabbitmq) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq v1")

	// Fetch the Rabbitmq instance
	instance := &rabbitmqv1.Rabbitmq{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not statefulsetFound, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	// secrets used for API requests and user control
	reqLogger.Info("Reconciling secrets")
	secretNames, err := r.reconcileSecrets(reqLogger, instance)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	statefulset := newStatefulSet(instance, secretNames)
	if err := controllerutil.SetControllerReference(instance, statefulset, r.scheme); err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	statefulsetFound := &v1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: statefulset.Name, Namespace: statefulset.Namespace}, statefulsetFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new statefulset", "statefulset.Namespace", statefulset.Namespace, "statefulset.Name", statefulset.Name)
		err = r.client.Create(context.TODO(), statefulset)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			return reconcile.Result{}, err
		}

		// statefulset created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(statefulsetFound.Spec, statefulset.Spec) {
		statefulsetFound.Spec.Replicas = statefulset.Spec.Replicas
		statefulsetFound.Spec.Template = statefulset.Spec.Template
		statefulsetFound.Spec.Selector = statefulset.Spec.Selector
	}

	if !reflect.DeepEqual(statefulsetFound.Annotations, statefulset.Annotations) {
		statefulsetFound.Annotations = statefulset.Annotations
	}

	if !reflect.DeepEqual(statefulsetFound.Labels, statefulset.Labels) {
		statefulsetFound.Labels = statefulset.Labels
	}

	reqLogger.Info("Reconcile statefulset", "statefulset.Namespace", statefulsetFound.Namespace, "statefulset.Name", statefulsetFound.Name)
	if err = r.client.Update(context.TODO(), statefulsetFound); err != nil {
		reqLogger.Info("Reconcile statefulset error", "statefulset.Namespace", statefulsetFound.Namespace, "statefulset.Name", statefulsetFound.Name)
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	// creating services
	reqLogger.Info("Reconciling services")

	_, err = r.reconcileHTTPService(reqLogger, instance)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	// all-in-one service
	_, err = r.reconcileHAService(reqLogger, instance)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	_, err = r.reconcileDiscoveryService(reqLogger, instance)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	// configmap
	reqLogger.Info("Reconciling configmap")

	_, err = r.reconcileConfigMap(reqLogger, instance, secretNames)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	// check prometheus exporter flag
	if instance.Spec.RabbitmqPrometheusExporterPort > 0 {
		_, err = r.reconcilePrometheusExporterService(reqLogger, instance)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			return reconcile.Result{}, err
		}
	}

	// use ServiceMonitor?
	if instance.Spec.RabbitmqUseServiceMonitor {
		_, err = r.reconcilePrometheusExporterServiceMonitor(reqLogger, instance)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			return reconcile.Result{}, err
		}
	}

	// set policies
	reqLogger.Info("Setting up policies")
	timeoutPolicies, _ := time.ParseDuration("30")
	timeoutFlagPolicies := false
	ctxPolicies, ctxPoliciesCancelTimeout := context.WithTimeout(context.Background(), timeoutPolicies)
	defer ctxPoliciesCancelTimeout()
	go r.setPolicies(ctxPolicies, reqLogger, instance, secretNames)
	for {
		if timeoutFlagPolicies {
			break
		}
		select {
		case <-ctxPolicies.Done():
			timeoutFlagPolicies = true
		}
	}

	// set additional users
	reqLogger.Info("Setting up additional users")
	timeoutUsers, _ := time.ParseDuration("30")
	timeoutFlagUsers := false
	ctxUsers, ctxUsersCancelTimeout := context.WithTimeout(context.Background(), timeoutUsers)
	defer ctxUsersCancelTimeout()
	go r.syncUsersCredentials(ctxUsers, reqLogger, instance, secretNames)
	for {
		if timeoutFlagUsers {
			break
		}
		select {
		case <-ctxUsers.Done():
			timeoutFlagUsers = true
		}
	}

	// reconcile PodDisruptionBudget
	reqLogger.Info("Reconciling PodDisruptionBudget")

	_, err = r.reconcilePdb(reqLogger, instance)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	_, err = r.reconcileFinalizers(reqLogger, instance)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil

}

func appendNodeVariables(env []corev1.EnvVar, cr *rabbitmqv1.Rabbitmq) []corev1.EnvVar {
	return append(env,
		corev1.EnvVar{
			Name:  "RABBITMQ_USE_LONGNAME",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "K8S_HOSTNAME_SUFFIX",
			Value: "." + cr.Name + "-discovery." + cr.Namespace + "." + cr.Spec.RabbitmqK8SServiceDiscovery,
		},
		corev1.EnvVar{
			Name:  "K8S_SERVICE_NAME",
			Value: cr.Name + "-discovery",
		},
		corev1.EnvVar{
			Name: "MY_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		corev1.EnvVar{
			Name:  "RABBITMQ_NODENAME",
			Value: "rabbit@$(MY_POD_NAME)." + cr.Name + "-discovery." + cr.Namespace + "." + cr.Spec.RabbitmqK8SServiceDiscovery,
		},
	)
}

func newStatefulSet(cr *rabbitmqv1.Rabbitmq, secretNames secretResouces) *v1.StatefulSet {

	// prepare containers for pod
	podContainers := []corev1.Container{}

	// check affinity rules
	affinity := &corev1.Affinity{}
	if cr.Spec.RabbitmqAffinity != nil {
		affinity = cr.Spec.RabbitmqAffinity
	}

	// container with rabbitmq
	rabbitmqContainer := corev1.Container{
		Name:  "rabbitmq",
		Image: cr.Spec.K8SImage.Name + ":" + cr.Spec.K8SImage.Tag,
		Env: append(appendNodeVariables(cr.Spec.K8SENV, cr), corev1.EnvVar{
			Name:      "RABBITMQ_ERLANG_COOKIE",
			ValueFrom: &corev1.EnvVarSource{ConfigMapKeyRef: &corev1.ConfigMapKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name}, Key: ".erlang.cookie"}},
		}),
		Resources: corev1.ResourceRequirements{
			Requests: cr.Spec.RabbitmqPodRequests,
			Limits:   cr.Spec.RabbitmqPodLimits,
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "rabbit-etc",
				MountPath: "/etc/rabbitmq",
			},
			{
				Name:      "rabbit-data",
				MountPath: "/var/lib/rabbitmq",
			},
			{
				Name:      "rabbit-config",
				MountPath: "/rabbit-config",
			},
		},
	}

	podContainers = append(podContainers, rabbitmqContainer)

	// if prometheus exporter enabled add additional container to pod
	if cr.Spec.RabbitmqPrometheusExporterPort > 0 {

		exporterImageAndTag := "kbudde/rabbitmq-exporter:latest"
		if cr.Spec.RabbitmqPrometheusImage != "" {
			exporterImageAndTag = cr.Spec.RabbitmqPrometheusImage
		}

		exporterContainer := corev1.Container{
			Name:            "prometheus-exporter",
			Image:           exporterImageAndTag,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env: []corev1.EnvVar{
				{
					Name:      "RABBIT_USER",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: secretNames.ServiceAccount}, Key: "username"}},
				},
				{
					Name:      "RABBIT_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: secretNames.ServiceAccount}, Key: "password"}},
				},
			},
			Ports: []corev1.ContainerPort{
				{
					Name:          "exporter",
					Protocol:      corev1.ProtocolTCP,
					ContainerPort: cr.Spec.RabbitmqPrometheusExporterPort,
				},
			},
		}
		podContainers = append(podContainers, exporterContainer)
	}

	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: returnLabels(cr),
			Annotations: returnAnnotations(cr),
		},
		Spec: corev1.PodSpec{
			Affinity:           affinity,
			ServiceAccountName: cr.Spec.RabbitmqK8SServiceAccount,
			InitContainers: []corev1.Container{
				{
					Name:    "copy-rabbitmq-config",
					Image:   "busybox",
					Command: []string{"sh", "/rabbit-config/init.sh"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "rabbit-etc",
							MountPath: "/etc/rabbitmq",
						},
						{
							Name:      "rabbit-config",
							MountPath: "/rabbit-config",
						},
						{
							Name:      "rabbit-data",
							MountPath: "/var/lib/rabbitmq",
						},
					},
				},
			},
			Containers:   podContainers,
			Tolerations:  cr.Spec.Tolerations,
			NodeSelector: cr.Spec.NodeSelector,
			Volumes: []corev1.Volume{
				{
					Name: "rabbit-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: cr.Name,
							},
						},
					},
				},
				{
					Name: "rabbit-etc",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	PVCTemplate := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "rabbit-data",
			Finalizers: cr.ObjectMeta.Finalizers,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: cr.Spec.RabbitmqVolumeSize,
				},
			},
		},
	}

	return &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels: mergeMaps(returnLabels(cr),
				map[string]string{"rabbitmq.improvado.io/component": "messaging"},
			),
		},
		Spec: v1.StatefulSetSpec{
			Replicas:    &cr.Spec.RabbitmqReplicas,
			ServiceName: cr.Name + "-discovery",
			Selector: &metav1.LabelSelector{
				MatchLabels: returnLabels(cr),
			},
			Template:             podTemplate,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{PVCTemplate},
			UpdateStrategy: v1.StatefulSetUpdateStrategy{
				Type: v1.RollingUpdateStatefulSetStrategyType,
			},
		},
	}
}
