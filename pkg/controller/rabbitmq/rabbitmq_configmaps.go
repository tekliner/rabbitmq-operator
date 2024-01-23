package rabbitmq

import (
	"bytes"
	"context"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/leekchan/gtf"
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
	Watermark       string
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
vm_memory_high_watermark_paging_ratio = {{ .Spec.RabbitmqMemoryHighWatermarkPagingRatio | default "0.8" }}
vm_memory_high_watermark.absolute = {{ .Watermark }}
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

const initRabbitmqScript = `# RabbitMQ Init script
rm -f /var/lib/rabbitmq/.erlang.cookie
cp /rabbit-config/* /etc/rabbitmq
cp /rabbit-config/.* /etc/rabbitmq
chmod 600 /etc/rabbitmq/.erlang.cookie  
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

func (r *ReconcileRabbitmq) reconcileConfigMap(reqLogger logr.Logger, cr *rabbitmqv1.Rabbitmq, secretNames secretResouces) (reconcile.Result, error) {
	reqLogger.Info("Started reconciling Configmap", "ConfigMap.Namespace", cr.Namespace, "ConfigMap.Name", cr.Name)
	var err error
	var templateData templateDataStruct

	memoryLimit, _ := cr.Spec.RabbitmqPodLimits[corev1.ResourceMemory]
	memoryLimitBytes := memoryLimit.Value()
	watermarkLimitBytes := memoryLimitBytes / 2

	if cr.Spec.RabbitmqMemoryHighWatermark != "" {
		templateData.Watermark = cr.Spec.RabbitmqMemoryHighWatermark
	} else {
		templateData.Watermark = strconv.FormatInt(watermarkLimitBytes, 10)
	}

	secretObj := corev1.Secret{}
	reqLogger.Info("Configmap receiving secret", "ConfigMap.Namespace", cr.Namespace, "ConfigMap.Name", cr.Name, "Secret name", secretNames.ServiceAccount)
	secretObj, err = r.getSecret(secretNames.ServiceAccount, cr.Namespace)
	if err != nil {
		reqLogger.Info("Configmap can't receive secret", "ConfigMap.Namespace", cr.Namespace, "ConfigMap.Name", cr.Name, "Secret name", secretNames.ServiceAccount)
		return reconcile.Result{}, err
	}

	defaultUsername := string(secretObj.Data["username"])
	if defaultUsername == "" {
		reqLogger.Info("Empty service username", "ConfigMap.Namespace", cr.Namespace, "ConfigMap.Name", cr.Name)
		return reconcile.Result{}, err
	}
	templateData.DefaultUser = defaultUsername
	defaultPassword := string(secretObj.Data["password"])
	if defaultPassword == "" {
		reqLogger.Info("Empty service password", "ConfigMap.Namespace", cr.Namespace, "ConfigMap.Name", cr.Name)
		return reconcile.Result{}, err
	}
	templateData.DefaultPassword = defaultPassword
	cookieData := string(secretObj.Data["cookie"])
	if cookieData == "" {
		reqLogger.Info("Empty cookie data", "ConfigMap.Namespace", cr.Namespace, "ConfigMap.Name", cr.Name)
		return reconcile.Result{}, err
	}
	reqLogger.Info("Configmap decoded secret", "ConfigMap.Namespace", cr.Namespace, "ConfigMap.Name", cr.Name, "Secret cookie", cookieData)

	templateData.Spec = cr.Spec

	resultConfig, err := applyDataOnTemplate(reqLogger, defaultRabbitmqConfig, templateData)
	if err != nil {
		return reconcile.Result{}, err
	}
	resultPlugins, err := applyDataOnTemplate(reqLogger, defaultRabbitmqPlugins, templateData)
	if err != nil {
		return reconcile.Result{}, err
	}

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    returnLabels(cr),
		},
		Data: map[string]string{
			"rabbitmq.conf":   resultConfig,
			"enabled_plugins": resultPlugins,
			".erlang.cookie":  cookieData,
			"init.sh":         initRabbitmqScript,
		},
	}

	if err := controllerutil.SetControllerReference(cr, configmap, r.scheme); err != nil {
		reqLogger.Info("Configmap can't set controller reference", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
		return reconcile.Result{}, err
	}

	found := &corev1.ConfigMap{}
	reqLogger.Info("Trying to receive configmap", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: configmap.Name, Namespace: configmap.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		reqLogger.Info("Creating ConfigMap", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
		err = r.client.Create(context.TODO(), configmap)
		found = configmap

		if err != nil {
			reqLogger.Info("Creating ConfigMap error", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
			return reconcile.Result{}, err
		}

	} else if err != nil {
		reqLogger.Info("Unknown error while getting ConfigMap", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(found.Data, configmap.Data) {
		reqLogger.Info("Configmap not equal to received", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
		found.Data = configmap.Data
	}

	if err = r.client.Update(context.TODO(), found); err != nil {
		reqLogger.Info("Configmap can't be updated", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
		return reconcile.Result{}, err
	}

	reqLogger.Info("Configmap successfuly reconciled", "ConfigMap.Namespace", configmap.Namespace, "ConfigMap.Name", configmap.Name)
	return reconcile.Result{}, nil
}
