package gardener

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerApi "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-incubator/hydroform/provision/internal/operator/native/gardener/aws"
	"github.com/kyma-incubator/hydroform/provision/internal/operator/native/gardener/azure"
	"github.com/kyma-incubator/hydroform/provision/internal/operator/native/gardener/gcp"
	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/clientcmd"
)

/*-- Gardener native operator --*/

func Create(ops *types.Options, cfg map[string]interface{}) (*types.ClusterInfo, error) {
	client, err := seedClient(cfg["credentials_file_path"].(string))
	if err != nil {
		return nil, errors.Wrap(err, "error creating the gardener client from credentials")
	}
	shoot, err := toShoot(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error generating shoot spec from config")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute) // TODO use the timeouts in the Options param
	defer cancel()

	_, err = client.Shoots(cfg["namespace"].(string)).Create(ctx, shoot, v1.CreateOptions{})
	if err != nil {
		return &types.ClusterInfo{
			Status: &types.ClusterStatus{
				Phase: types.Errored,
			},
		}, err
	}

	if err := waitForShoot(ctx, client, cfg["cluster_name"].(string), cfg["namespace"].(string)); err != nil {
		return nil, err
	}

	return &types.ClusterInfo{
		Status: &types.ClusterStatus{
			Phase: types.Provisioned,
		},
	}, nil
}

func Status(ops *types.Options, info *types.ClusterInfo, cfg map[string]interface{}) (*types.ClusterStatus, error) {
	client, err := seedClient(cfg["credentials_file_path"].(string))
	if err != nil {
		return nil, errors.Wrap(err, "error creating the gardener client from credentials")
	}
	_, err = client.Shoots(cfg["namespace"].(string)).Get(context.TODO(), cfg["cluster_name"].(string), v1.GetOptions{})
	if err != nil {
		return &types.ClusterStatus{
			Phase: types.Errored,
		}, err
	}
	return &types.ClusterStatus{
		Phase: types.Provisioned,
	}, nil
}

func Delete(ops *types.Options, info *types.ClusterInfo, cfg map[string]interface{}) error {
	client, err := seedClient(cfg["credentials_file_path"].(string))
	if err != nil {
		return errors.Wrap(err, "error creating the gardener client from credentials")
	}

	return client.Shoots(cfg["namespace"].(string)).Delete(context.TODO(), cfg["cluster_name"].(string), v1.DeleteOptions{})
}

func waitForShoot(ctx context.Context, client *gardenerApi.CoreV1beta1Client, name, namespace string) error {

	timer := time.NewTicker(15 * time.Second)

	for {
		select {
		case <-timer.C:
			sh, err := client.Shoots(namespace).Get(context.Background(), name, v1.GetOptions{})
			if err != nil {
				return err
			}
			// TODO refactor hacky readiness check
			if sh.Status.LastOperation.Progress == 100 && sh.Status.LastOperation.State == gardenerTypes.LastOperationStateSucceeded {
				fmt.Println("I got to status success!")
				return nil
			}
		case <-ctx.Done():
			return errors.New("Provisioning timed out")
		}
	}
}

/*-- Gardener client --*/

func seedClient(credentialsFile string) (*gardenerApi.CoreV1beta1Client, error) {
	kubeBytes, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.RESTConfigFromKubeConfig(kubeBytes)
	if err != nil {
		return nil, err
	}
	return gardenerApi.NewForConfig(config)
}

/*-- Shoot building functions --*/

func toShoot(cfg map[string]interface{}) (*gardenerTypes.Shoot, error) {

	shoot := &gardenerTypes.Shoot{
		ObjectMeta: shootMeta(cfg),
		Spec:       shootSpec(cfg),
	}

	err := injectProvider(&shoot.Spec, cfg)

	return shoot, err
}

func shootMeta(cfg map[string]interface{}) v1.ObjectMeta {
	o := v1.ObjectMeta{}

	if v, ok := cfg["cluster_name"].(string); ok && len(v) > 0 {
		o.Name = v
	}
	if v, ok := cfg["namespace"].(string); ok && len(v) > 0 {
		o.Namespace = v
	}
	if v, ok := cfg["annotations"].(map[string]string); ok && len(v) > 0 {
		o.Annotations = v
	}

	return o
}

func shootSpec(cfg map[string]interface{}) gardenerTypes.ShootSpec {
	o := gardenerTypes.ShootSpec{}

	if v, ok := cfg["target_profile"].(string); ok && len(v) > 0 {
		o.CloudProfileName = v
	}
	if v, ok := cfg["target_secret"].(string); ok && len(v) > 0 {
		o.SecretBindingName = v
	}
	if v, ok := cfg["seed_name"].(string); ok && len(v) > 0 {
		o.SeedName = &v
	}
	if v, ok := cfg["location"].(string); ok && len(v) > 0 {
		o.Region = v
	}
	if v, ok := cfg["purpose"].(string); ok && len(v) > 0 {
		*o.Purpose = gardenerTypes.ShootPurpose(v)
	}

	o.Kubernetes = shootK8s(cfg)
	o.Networking = shootNetworking(cfg)
	o.Maintenance = shootMaintenance()
	o.Hibernation = shootHibernation(cfg)
	return o
}

func shootK8s(cfg map[string]interface{}) gardenerTypes.Kubernetes {
	k := gardenerTypes.Kubernetes{}

	if v, ok := cfg["kubernetes_version"].(string); ok && len(v) > 0 {
		k.Version = v
	}
	if v, ok := cfg["privileged_containers"].(bool); ok {
		k.AllowPrivilegedContainers = &v
	}

	k.KubeAPIServer = &gardenerTypes.KubeAPIServerConfig{}
	if v, ok := cfg["enable_basic_auth"].(bool); ok {
		k.KubeAPIServer.EnableBasicAuthentication = &v
		if v {
			k.KubeAPIServer.OIDCConfig = oidcConfig(cfg)
		}
	}
	return k
}

func oidcConfig(cfg map[string]interface{}) *gardenerTypes.OIDCConfig {
	o := &gardenerTypes.OIDCConfig{}

	if v, ok := cfg["oidc_ca_bundle"].(string); ok && len(v) > 0 {
		o.CABundle = &v
	}
	if v, ok := cfg["oidc_client_id"].(string); ok && len(v) > 0 {
		o.ClientID = &v
	}
	if v, ok := cfg["oidc_groups_claim"].(string); ok && len(v) > 0 {
		o.GroupsClaim = &v
	}
	if v, ok := cfg["oidc_groups_prefix"].(string); ok && len(v) > 0 {
		o.GroupsPrefix = &v
	}
	if v, ok := cfg["oidc_issuer_url"].(string); ok && len(v) > 0 {
		o.IssuerURL = &v
	}
	if v, ok := cfg["oidc_required_claims"].(map[string]string); ok {
		o.RequiredClaims = v
	}
	if v, ok := cfg["oidc_signing_algs"].([]string); ok {
		o.SigningAlgs = v
	}
	if v, ok := cfg["oidc_username_claim"].(string); ok && len(v) > 0 {
		o.UsernameClaim = &v
	}
	if v, ok := cfg["oidc_username_prefix"].(string); ok && len(v) > 0 {
		o.UsernamePrefix = &v
	}
	return o
}

func shootNetworking(cfg map[string]interface{}) gardenerTypes.Networking {
	n := gardenerTypes.Networking{}

	if v, ok := cfg["networking_type"].(string); ok && len(v) > 0 {
		n.Type = v
	}
	if v, ok := cfg["networking_nodes"].(string); ok && len(v) > 0 {
		n.Nodes = &v
	}
	return n
}

func shootMaintenance() *gardenerTypes.Maintenance {
	return &gardenerTypes.Maintenance{
		AutoUpdate: &gardenerTypes.MaintenanceAutoUpdate{
			KubernetesVersion:   true,
			MachineImageVersion: true,
		},
		TimeWindow: &gardenerTypes.MaintenanceTimeWindow{
			Begin: "030000+0000",
			End:   "040000+0000",
		},
	}
}

func shootHibernation(cfg map[string]interface{}) *gardenerTypes.Hibernation {
	h := gardenerTypes.Hibernation{}

	if v, ok := cfg["hibernation_start"].(string); ok && len(v) > 0 {
		enabled := true
		h.Enabled = &enabled
		h.Schedules = append(h.Schedules, gardenerTypes.HibernationSchedule{
			Start: &v,
		})
		if v, ok := cfg["hibernation_end"].(string); ok && len(v) > 0 {
			h.Schedules[0].End = &v
		}
		if v, ok := cfg["hibernation_location"].(string); ok && len(v) > 0 {
			h.Schedules[0].Location = &v
		}
	}

	return &h
}

//injectProvider adds the provider config to the given shoot.
// It is done after building the shoot to minimize error propagation
func injectProvider(spec *gardenerTypes.ShootSpec, cfg map[string]interface{}) error {
	p := gardenerTypes.Provider{}

	if v, ok := cfg["target_provider"].(string); ok && len(v) > 0 {
		p.Type = v
	}
	var err error
	switch p.Type {
	case string(types.Azure):
		if p.ControlPlaneConfig, err = azure.ControlPlaneConfig(cfg); err != nil {
			return err
		}
		if p.InfrastructureConfig, err = azure.InfraConfig(cfg); err != nil {
			return err
		}
	case string(types.AWS):
		if p.ControlPlaneConfig, err = aws.ControlPlaneConfig(cfg); err != nil {
			return err
		}
		if p.InfrastructureConfig, err = aws.InfraConfig(cfg); err != nil {
			return err
		}

	case string(types.GCP):
		if p.ControlPlaneConfig, err = gcp.ControlPlaneConfig(cfg); err != nil {
			return err
		}
		if p.InfrastructureConfig, err = gcp.InfraConfig(cfg); err != nil {
			return err
		}
	}

	p.Workers = append(p.Workers, shootWorker(cfg))

	spec.Provider = p
	return nil
}

func shootWorker(cfg map[string]interface{}) gardenerTypes.Worker {
	w := gardenerTypes.Worker{
		Name:   "cpu-worker",
		Volume: &gardenerTypes.Volume{},
		Machine: gardenerTypes.Machine{
			Image: &gardenerTypes.ShootMachineImage{},
		},
	}

	if v, ok := cfg["zones"].([]string); ok && len(v) > 0 {
		w.Zones = v
	}
	if v, ok := cfg["worker_max_surge"].(int); ok {
		i := intstr.FromInt(v)
		w.MaxSurge = &i
	}
	if v, ok := cfg["worker_max_unavailable"].(int); ok {
		i := intstr.FromInt(v)
		w.MaxUnavailable = &i
	}
	if v, ok := cfg["worker_maximum"].(int); ok {
		w.Maximum = int32(v)
	}
	if v, ok := cfg["worker_minimum"].(int); ok {
		w.Minimum = int32(v)
	}
	if v, ok := cfg["disk_size"].(int); ok && v > 0 {
		w.Volume.VolumeSize = fmt.Sprintf("%dGi", v)
	}
	if v, ok := cfg["disk_type"].(string); ok && len(v) > 0 {
		w.Volume.Type = &v
	}
	if v, ok := cfg["machine_image_name"].(string); ok && len(v) > 0 {
		w.Machine.Image.Name = v
	}
	if v, ok := cfg["machine_image_version"].(string); ok && len(v) > 0 {
		w.Machine.Image.Version = &v
	}
	if v, ok := cfg["machine_type"].(string); ok && len(v) > 0 {
		w.Machine.Type = v
	}
	return w
}
