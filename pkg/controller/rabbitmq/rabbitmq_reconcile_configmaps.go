package rabbitmq

import (
	"bytes"
	"context"
	"reflect"

	"github.com/go-logr/logr"
	gtf "github.com/leekchan/gtf"
	rabbitmqv1 "github.com/tekliner/rabbitmq-operator/pkg/apis/rabbitmq/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type templateDataStruct struct {
	Spec            rabbitmqv1.RabbitmqSpec
	DefaultUser     string
	DefaultPassword string
}

const defaultRabbitmqConfig = `# RabbitMQ operator templated config
default_user = {{ .DefaultUser | default "rabbit" }}
default_pass = {{ .DefaultPassword | default "rabbit" }}
default_vhost = {{ .Spec.RabbitmqVhost | default "rabbit" }}

cluster_formation.peer_discovery_backend  = {{ .Spec.RabbitmqK8SPeerDiscoveryBackend | default "rabbit_peer_discovery_k8s" }}
cluster_formation.k8s.host = {{ .Spec.RabbitmqK8SHost | default "kubernetes.default.svc.cluster.local" }}
cluster_formation.k8s.address_type = {{ .Spec.RabbitmqK8SAddrType | default "hostname" }}
cluster_formation.node_cleanup.interval = {{ .Spec.RabbitmqClusterFormationNodeCleanup | default "10" }}
cluster_formation.node_cleanup.only_log_warning = true
cluster_partition_handling = {{ .Spec.RabbitmqClusterPartitionHandling | default "autoheal" }}
loopback_users.guest = false
hipe_compile = {{ .Spec.RabbitmqHipeCompile | default "false" }}
vm_memory_high_watermark.absolute = {{ .Spec.RabbitmqMemoryHighWatermark }}
`

const defaultRabbitmqPlugins = `[
rabbitmq_consistent_hash_exchange,
rabbitmq_federation,
rabbitmq_federation_management,
rabbitmq_management,
rabbitmq_peer_discovery_k8s,
rabbitmq_shovel,
rabbitmq_shovel_management
{{range .Spec.RabbitmqPlugins}}
{{ . }},
{{end}}].
`

func applyDataOnTemplate(reqLogger logr.Logger, templateContent string, cr templateDataStruct) (string, error) {
	var buf bytes.Buffer
	templateObj, err := gtf.New("config").Parse(templateContent)
	if err != nil {
		reqLogger.Error(err, "applyDataOnTemplate")
	}
	err = templateObj.Execute(&buf, cr)
	if err != nil {
		reqLogger.Error(err, "templateObj.Execute")
	}
	return buf.String(), err
}

func (r *ReconcileRabbitmq) reconcileConfigMap(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq) (reconcile.Result, error) {
	var err error
	var templateData templateDataStruct
	secretObj := corev1.Secret{}

	// detect right secret name
	secretNameSA := cr.Name + "-service-account"
	if cr.Spec.RabbitmqSecretServiceAccount != "" {
		secretObj, err = r.getSecret(cr.Spec.RabbitmqSecretServiceAccount, cr.Namespace)
		if err != nil {
			return reconcile.Result{}, err
		}
		secretNameSA = cr.Spec.RabbitmqSecretServiceAccount
	}

	secretObj, err = r.getSecret(secretNameSA, cr.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	defaultUsername, err := secretDecode(secretObj.Data["username"])
	templateData.DefaultUser = defaultUsername
	defaultPassword, err := secretDecode(secretObj.Data["password"])
	templateData.DefaultPassword = defaultPassword

	templateData.Spec = cr.Spec

	resultConfig, err := applyDataOnTemplate(reqLogger, defaultRabbitmqConfig, templateData)
	if err != nil {
		return reconcile.Result{}, err
	}
	resultPlugins, err := applyDataOnTemplate(reqLogger, defaultRabbitmqPlugins, templateData)
	if err != nil {
		return reconcile.Result{}, err
	}

	labels := map[string]string{
		"rabbitmq.improvado.io/app":       "rabbitmq",
		"rabbitmq.improvado.io/name":      cr.Name,
		"rabbitmq.improvado.io/component": "messaging",
	}

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"rabbitmq.conf":   resultConfig,
			"enabled_plugins": resultPlugins,
		},
	}

	if err := controllerutil.SetControllerReference(cr, configmap, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: configmap.Name, Namespace: configmap.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("Reconciling ConfigMap", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
		err = r.client.Create(context.TODO(), configmap)

		if err != nil {
			return reconcile.Result{}, err
		}

	} else if err != nil {
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(found.Data, configmap.Data) {
		found.Data = configmap.Data
	}

	if err = r.client.Update(context.TODO(), found); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
