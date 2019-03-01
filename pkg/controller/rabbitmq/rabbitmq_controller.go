package rabbitmq

import (
	"context"
	"encoding/json"
	"net"
	"time"

	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

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
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rabbitmq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Rabbitmq
	err = c.Watch(&source.Kind{Type: &rabbitmqv1.Rabbitmq{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Rabbitmq
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rabbitmqv1.Rabbitmq{},
	})
	if err != nil {
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
func (r *ReconcileRabbitmq) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq")

	// Fetch the Rabbitmq instance
	instance := &rabbitmqv1.Rabbitmq{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	statefulset := newStatefulSet(instance)
	if err := controllerutil.SetControllerReference(instance, statefulset, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &v1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: statefulset.Name, Namespace: statefulset.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new statefulset", "statefulset.Namespace", statefulset.Namespace, "statefulset.Name", statefulset.Name)
		err = r.client.Create(context.TODO(), statefulset)
		if err != nil {
			return reconcile.Result{}, err
		}

		// statefulset created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// statefulset already exists - don't requeue
	reqLogger.Info("Skip reconcile: statefulset already exists", "statefulset.Namespace", found.Namespace, "statefulset.Name", found.Name)

	// set policies
	timeout, _ := time.ParseDuration("30")
	ctx, ctxCancelTimeout := context.WithTimeout(context.Background(), timeout)
	defer ctxCancelTimeout()
	go setPolicies(ctx, instance)

	return reconcile.Result{}, nil

}

func newStatefulSet(cr *rabbitmqv1.Rabbitmq) *v1.StatefulSet {
	labels := map[string]string{
		"rabbitmq.improvado.io/app":       "rabbitmq",
		"rabbitmq.improvado.io/name":      cr.Name,
		"rabbitmq.improvado.io/component": "messaging",
	}

	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "rabbitmq",
					Image: cr.Spec.Image.Name + ":" + cr.Spec.Image.Tag,
					Env:   cr.Spec.ENV,
					Resources: corev1.ResourceRequirements{
						Requests: cr.Spec.RabbitmqPodRequests,
						Limits:   cr.Spec.RabbitmqPodLimits,
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "rabbit-data",
							MountPath: "/var/lib/rabbitmq",
						},
					},
				},
			},
		},
	}

	PVCTemplate := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rabbit-data",
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
			Labels:    labels,
		},
		Spec: v1.StatefulSetSpec{
			Replicas: &cr.Spec.RabbitmqReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template:             podTemplate,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{PVCTemplate},
			UpdateStrategy: v1.StatefulSetUpdateStrategy{
				Type: v1.RollingUpdateStatefulSetStrategyType,
			},
		},
	}
}

// sendPolicy send policy to api port
func putPolicy(url string, policy []byte) {
	// send request
}

// cleanPolicies return policies and remove all by name, no other method supported
func cleanPolicies(apiService string, cr *rabbitmqv1.Rabbitmq) {
	url := apiService + "/api/policies"

	// request will return something like that:
	// [{"vhost":"dts","name":"ha-three","pattern":".*","apply-to":"all","definition":
	//{"ha-mode":"exactly","ha-params":3,"ha-sync-mode":"automatic"},"priority":0}]

	response := getRequest(url)
	var policies []rabbitmqv1.RabbitmqPolicy
	err := json.Unmarshal(response, &policies)
	if err != nil {
		// something bad
	}

	for _, policyRecord := range policies {
		deleteRequest(apiService + "/" + policyRecord.Vhost + "/" + policyRecord.Name)
	}

}

// setPolicies run as go routine
func setPolicies(ctx context.Context, cr *rabbitmqv1.Rabbitmq) {

	reqLogger := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.Name)
	reqLogger.Info("Setting up policies")

	// wait http connection to api port
	timeout := time.Duration(5 * time.Second)
	apiService := cr.Name + "-node:5672"

	for {
		_, err := net.DialTimeout("tcp", apiService, timeout)
		if err != nil {
			break
		}
	}

	//clean rabbit before fulfilling policies list
	cleanPolicies(apiService, cr)

	//fulfill policies list
	for _, policy := range cr.Spec.RabbitmqPolicies {
		policyJSON, _ := json.Marshal(policy)
		url := apiService + "/api/policies/" + cr.Name + "/" + policy.Name
		// send policy to api service
		reqLogger.Info("Adding policy " + policy.Name + ", URL: " + url + ", JSON: " + string(policyJSON))
		putPolicy(url, policyJSON)
	}
}

func newEpmdService(cr *rabbitmqv1.Rabbitmq) *corev1.Service {
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

	return service
}

func newNodeService(cr *rabbitmqv1.Rabbitmq) *corev1.Service {
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

	return service
}

func newHTTPService(cr *rabbitmqv1.Rabbitmq) *corev1.Service {
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

	return service
}
