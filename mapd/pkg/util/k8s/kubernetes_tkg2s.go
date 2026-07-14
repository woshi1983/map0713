// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	runv1alpha3 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/run.tanzu/v1alpha3"
	vmoperatorv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/vmoperator/v1alpha1"
	runversioned "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/run.tanzu/clientset/versioned"
	runscheme "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/run.tanzu/clientset/versioned/scheme"
	vmoperatorversioned "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/vmoperator/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
)

type VSphereWithTanzu2 struct {
	*ClusterAPI
	KappController *KappController

	LoginUser     string
	LoginPassword []byte
}

func (w *VSphereWithTanzu2) IsTokenValid() bool {
	off, err := validateUnverifiedJWTToken(w.ClusterAPI.Config.BearerToken)
	return err == nil && off > 0
}

func (w *VSphereWithTanzu2) RefreshTokenOnDemand() (t string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.RefreshTokenOnDemandContext(ctx)
}

func (w *VSphereWithTanzu2) RefreshTokenOnDemandContext(ctx context.Context) (t string, err error) {
	if w.IsTokenValid() {
		return w.ClusterAPI.Config.BearerToken, nil
	}

	var token string
	for r := 0; r < 3; r++ {
		ctxL, cancel := context.WithTimeout(ctx, time.Second*5)
		token, err = w.LoginContext(ctxL, "", "")
		cancel()
		if err != nil {
			continue
		}
		w.ClusterAPI.Config.BearerToken = token
		break
	}
	if err != nil {
		return w.ClusterAPI.Config.BearerToken, errors.Wrap(err, "failed to login")
	}
	return w.ClusterAPI.Config.BearerToken, nil
}

func (w *VSphereWithTanzu2) ApplyCluster(namespace, name string, cfg VSphereWithTanzu2ClusterConfig, dryrun bool) ([]byte, error) {
	return w.ClusterAPI.ApplyCluster(
		ClusterTemplateConfig{
			Name:      name,
			Namespace: namespace,
			Params: map[string]string{
				"Id":                         cfg.ID.String(),
				"Version":                    cfg.Version,
				"StorageClass":               cfg.StorageClass,
				"AdditionalTrustedCA":        cfg.AdditionalTrustedCA,
				"OsName":                     cfg.OSName, // `photon` or `ubuntu`
				"Cni":                        cfg.CNI,    // `antrea`, `calico`, or `antrea_nsx_pod_routing`
				"ControlPlaneReplicas":       strconv.FormatInt(cfg.ControlPlaneReplicas, 10),
				"ControlPlaneVmClass":        cfg.ControlPlaneVmClass,
				"ControlPlaneEtcdVolumeSize": strconv.FormatInt(cfg.ControlPlaneEtcdVolumeSizeGiB, 10) + "Gi",
				"WorkerNodePool0Replicas":    strconv.FormatInt(cfg.WorkerNodePool0Replicas, 10),
				"WorkerNodePool0Zone":        cfg.WorkerNodePool0Zone,
				"WorkerNodePool1Replicas":    strconv.FormatInt(cfg.WorkerNodePool1Replicas, 10),
				"WorkerNodePool1Zone":        cfg.WorkerNodePool1Zone,
				"WorkerNodePool2Replicas":    strconv.FormatInt(cfg.WorkerNodePool2Replicas, 10),
				"WorkerNodePool2Zone":        cfg.WorkerNodePool2Zone,
				"WorkerVmClass":              cfg.WorkerVmClass,
				"WorkerContainerdVolumeSize": strconv.FormatInt(cfg.WorkerContainerdVolumeSizeGiB, 10) + "Gi",
				"ClusterCIDR":                cfg.ClusterCIDR,
				"ServiceCIDR":                cfg.ServiceCIDR,
				"Ntp":                        cfg.NTP,
			},
		},
	)
}

func (w *VSphereWithTanzu2) GetTkgServiceConfiguration(name string) (*runv1alpha3.TkgServiceConfiguration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTkgServiceConfigurationContext(ctx, name)
}

func (w *VSphereWithTanzu2) GetTkgServiceConfigurationContext(ctx context.Context, name string) (*runv1alpha3.TkgServiceConfiguration, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha3().TkgServiceConfigurations().Get(ctx, name, metav1.GetOptions{})
}

func (w *VSphereWithTanzu2) GetTanzuKubernetesRelease(name string) (*runv1alpha3.TanzuKubernetesRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesReleaseContext(ctx, name)
}

func (w *VSphereWithTanzu2) GetTanzuKubernetesReleaseContext(ctx context.Context, name string) (*runv1alpha3.TanzuKubernetesRelease, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha3().TanzuKubernetesReleases().Get(ctx, name, metav1.GetOptions{})
}

func (w *VSphereWithTanzu2) GetTanzuKubernetesReleases(selectors ...Selector) (*runv1alpha3.TanzuKubernetesReleaseList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesReleasesContext(ctx, selectors...)
}

func (w *VSphereWithTanzu2) GetTanzuKubernetesReleasesContext(ctx context.Context, selectors ...Selector) (*runv1alpha3.TanzuKubernetesReleaseList, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	return clientset.RunV1alpha3().TanzuKubernetesReleases().List(ctx, opts)
}

func (w *VSphereWithTanzu2) GetTanzuKubernetesCluster(namespace, name string) (*runv1alpha3.TanzuKubernetesCluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesClusterContext(ctx, namespace, name)
}

func (w *VSphereWithTanzu2) GetTanzuKubernetesClusterContext(ctx context.Context, namespace, name string) (*runv1alpha3.TanzuKubernetesCluster, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha3().TanzuKubernetesClusters(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (w *VSphereWithTanzu2) GetTanzuKubernetesClusters(namespace string, selectors ...Selector) (*runv1alpha3.TanzuKubernetesClusterList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesClustersContext(ctx, namespace, selectors...)
}

func (w *VSphereWithTanzu2) GetTanzuKubernetesClustersContext(ctx context.Context, namespace string, selectors ...Selector) (*runv1alpha3.TanzuKubernetesClusterList, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &runv1alpha3.TanzuKubernetesClusterList{}
	err = clientset.RunV1alpha3().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("tanzukubernetesclusters").
		VersionedParams(&opts, runscheme.ParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *VSphereWithTanzu2) PatchClusterBootstrap(namespace, name string, pt types.PatchType, patch []byte) (*runv1alpha3.ClusterBootstrap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.PatchClusterBootstrapContext(ctx, namespace, name, pt, patch)
}

func (w *VSphereWithTanzu2) PatchClusterBootstrapContext(ctx context.Context, namespace, name string, pt types.PatchType, patch []byte) (*runv1alpha3.ClusterBootstrap, error) {
	clientset, err := runversioned.NewForConfig(w.Config)
	if err != nil {
		return nil, err
	}
	result := &runv1alpha3.ClusterBootstrap{}
	err = clientset.RunV1alpha3().RESTClient().Patch(pt).
		Namespace(namespace).
		Resource("clusterbootstraps").
		Name(name).
		VersionedParams(&metav1.PatchOptions{}, runscheme.ParameterCodec).
		Body(patch).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *VSphereWithTanzu2) GetClusterBootstrap(namespace, name string) (*runv1alpha3.ClusterBootstrap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetClusterBootstrapContext(ctx, namespace, name)
}

func (w *VSphereWithTanzu2) GetClusterBootstrapContext(ctx context.Context, namespace, name string) (*runv1alpha3.ClusterBootstrap, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha3().ClusterBootstraps(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (w *VSphereWithTanzu2) GetClusterBootstraps(namespace string, selectors ...Selector) (*runv1alpha3.ClusterBootstrapList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetClusterBootstrapsContext(ctx, namespace, selectors...)
}

func (w *VSphereWithTanzu2) GetClusterBootstrapsContext(ctx context.Context, namespace string, selectors ...Selector) (*runv1alpha3.ClusterBootstrapList, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &runv1alpha3.ClusterBootstrapList{}
	err = clientset.RunV1alpha3().RESTClient().Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("clusterbootstraps").
		VersionedParams(&opts, runscheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

func (w *VSphereWithTanzu2) GetVirtualMachineClass(name string) (*vmoperatorv1alpha1.VirtualMachineClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetVirtualMachineClassContext(ctx, name)
}

func (w *VSphereWithTanzu2) GetVirtualMachineClassContext(ctx context.Context, name string) (*vmoperatorv1alpha1.VirtualMachineClass, error) {
	clientset, err := vmoperatorversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	return clientset.VmoperatorV1alpha1().VirtualMachineClasses().Get(ctx, name, metav1.GetOptions{})
}

func (w *VSphereWithTanzu2) GetVirtualMachineClasses(selectors ...Selector) (*vmoperatorv1alpha1.VirtualMachineClassList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetVirtualMachineClassesContext(ctx, selectors...)
}

func (w *VSphereWithTanzu2) GetVirtualMachineClassesContext(ctx context.Context, selectors ...Selector) (*vmoperatorv1alpha1.VirtualMachineClassList, error) {
	clientset, err := vmoperatorversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	return clientset.VmoperatorV1alpha1().VirtualMachineClasses().List(ctx, opts)
}

func (w *VSphereWithTanzu2) GetVSphereMachineTemplate(namespace, name string) (*capvv1beta1.VSphereMachineTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetVSphereMachineTemplateContext(ctx, namespace, name)
}

func (w *VSphereWithTanzu2) GetVSphereMachineTemplateContext(ctx context.Context, namespace, name string) (*capvv1beta1.VSphereMachineTemplate, error) {
	config := *w.Config
	config.GroupVersion = &capvv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = VSphereProviderAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	result := &capvv1beta1.VSphereMachineTemplate{}
	err = client.Get().
		Namespace(namespace).
		Resource("vspheremachinetemplates").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, VSphereProviderAPIParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *VSphereWithTanzu2) GetVSphereMachineTemplates(namespace string, selectors ...Selector) (*capvv1beta1.VSphereMachineTemplateList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetVSphereMachineTemplatesContext(ctx, namespace, selectors...)
}

func (w *VSphereWithTanzu2) GetVSphereMachineTemplatesContext(ctx context.Context, namespace string, selectors ...Selector) (*capvv1beta1.VSphereMachineTemplateList, error) {
	config := *w.Config
	config.GroupVersion = &capvv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = VSphereProviderAPICodecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &capvv1beta1.VSphereMachineTemplateList{}
	err = client.Get().
		NamespaceIfScoped(namespace, namespace != "").
		Resource("vspheremachinetemplates").
		VersionedParams(&opts, VSphereProviderAPIParameterCodec).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *VSphereWithTanzu2) Login(namespace, name string) (token string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.LoginContext(ctx, namespace, name)
}

func (w *VSphereWithTanzu2) LoginContext(ctx context.Context, namespace, name string) (token string, err error) {
	var clientset *runversioned.Clientset
	clientset, err = runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return
	}

	var b, reqBody []byte
	reqBody, err = json.Marshal(&wcpLoginRequest{GuestClusterNamespace: namespace, GuestClusterName: name})
	if err != nil {
		return token, fmt.Errorf("unable to construct the login request: %v", err)
	}

	b, err = clientset.Discovery().RESTClient().Post().
		AbsPath("/wcp/login").
		Body(reqBody).
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(w.LoginUser+":"+string(w.LoginPassword)))).
		Do(ctx).
		Raw()
	if err != nil {
		return token, fmt.Errorf("error submitting the login request: %v", err)
	}

	d := &wcpLoginResponse{}
	if err := json.Unmarshal(b, d); err != nil {
		return token, fmt.Errorf("unable to parse the login response: %v", err)
	}
	return d.SessionID, nil
}

// NewVSphereWithTanzu2 creates a new Tanzu Kubernetes Grid 2.0 supervisor instance.
func NewVSphereWithTanzu2(endpoint, user string, password []byte, clientCfg rest.TLSClientConfig) (*VSphereWithTanzu2, error) {
	k8s, err := NewKubernetesFromToken(endpoint, []byte{}, clientCfg)
	if err != nil {
		return nil, err
	}
	capi, err := NewClusterAPI(k8s, tkg2sClusterTemplate)
	if err != nil {
		return nil, err
	}
	s := &VSphereWithTanzu2{
		ClusterAPI:     capi,
		KappController: NewKappController(k8s),
		LoginUser:      user,
		LoginPassword:  password,
	}
	s.RefreshTokenOnDemand()
	return s, nil
}

var (
	VSphereProviderAPIScheme         = runtime.NewScheme()
	VSphereProviderAPICodecs         = serializer.NewCodecFactory(VSphereProviderAPIScheme)
	VSphereProviderAPIParameterCodec = runtime.NewParameterCodec(VSphereProviderAPIScheme)
)

func init() {
	metav1.AddToGroupVersion(VSphereProviderAPIScheme, metav1.Unversioned)
	metav1.AddToGroupVersion(VSphereProviderAPIScheme, capvv1beta1.GroupVersion)

	capvv1beta1.AddToScheme(VSphereProviderAPIScheme)
}

type VSphereWithTanzu2ClusterConfig struct {
	ID                            uuid.UUID
	Version                       string
	StorageClass                  string
	AdditionalTrustedCA           string
	OSName                        string
	CNI                           string // antrea, calico, antrea_nsx_pod_routing
	ControlPlaneReplicas          int64
	ControlPlaneVmClass           string
	ControlPlaneEtcdVolumeSizeGiB int64
	WorkerNodePool0Replicas       int64
	WorkerNodePool0Zone           string
	WorkerNodePool1Replicas       int64
	WorkerNodePool1Zone           string
	WorkerNodePool2Replicas       int64
	WorkerNodePool2Zone           string
	WorkerVmClass                 string
	WorkerContainerdVolumeSizeGiB int64
	ClusterCIDR                   string
	ServiceCIDR                   string
	NTP                           string
}

const tkg2sClusterTemplate = `{{$uuid := ""}}{{if .Params.Id}}{{$uuid = .Params.Id}}{{else}}{{$uuid = uuidgen}}{{end}}{{if eq .Params.Cni "calico"}}---
apiVersion: cni.tanzu.vmware.com/v1alpha1
kind: CalicoConfig
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  calico:
    config:
      vethMTU: 0
---
apiVersion: run.tanzu.vmware.com/v1alpha3
kind: ClusterBootstrap
metadata:
  annotations:
    tkg.tanzu.vmware.com/add-missing-fields-from-tkr: {{.Params.Version}}
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  cni:
    refName: calico*
    valuesFrom:
      providerRef:
        apiGroup: cni.tanzu.vmware.com
        kind: CalicoConfig
        name: {{.Name}}
{{else if eq .Params.Cni "antrea_nsx_pod_routing"}}---
apiVersion: cpi.tanzu.vmware.com/v1alpha1
kind: VSphereCPIConfig
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  vsphereCPI:
    antreaNSXPodRoutingEnabled: true
    insecure: false
    mode: vsphereParavirtualCPI
---
apiVersion: cni.tanzu.vmware.com/v1alpha1
kind: AntreaConfig
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  antrea:
    config:
      featureGates:
        AntreaProxy: true
        EndpointSlice: true
        AntreaPolicy: true
        FlowExporter: true
        Egress: true
        NodePortLocal: true
        AntreaTraceflow: true
        NetworkPolicyStats: true
      noSNAT: true
      trafficEncapMode: noEncap
{{end}}{{if .Params.AdditionalTrustedCA}}---
apiVersion: v1
data:
  {{$uuid}}: {{.Params.AdditionalTrustedCA | atob | atob}}
kind: Secret
metadata:
  name: {{.Name}}-user-trusted-ca-secret
  namespace: {{.Namespace}}
type: Opaque
---
apiVersion: run.tanzu.vmware.com/v1alpha3
kind: KappControllerConfig
metadata:
  name: {{.Name}}-kapp-controller-package
  namespace: {{.Namespace}}
spec:
  kappController:
    createNamespace: false
    config:
      caCerts: {{.Params.AdditionalTrustedCA | quote}}
      dangerousSkipTLSVerify: ""
      httpProxy: ""
      httpsProxy: ""
      noProxy: ""
    deployment:
      apiPort: 10100
      concurrency: 4
      hostNetwork: true
      metricsBindAddress: "0"
      priorityClassName: system-cluster-critical
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      - effect: NoSchedule
        key: node-role.kubernetes.io/control-plane
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - effect: NoSchedule
        key: node.kubernetes.io/not-ready
      - effect: NoSchedule
        key: node.cloudprovider.kubernetes.io/uninitialized
        value: "true"
    globalNamespace: tkg-system
  namespace: tkg-system
{{end}}---
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - {{.Params.ClusterCIDR}}
    serviceDomain: cluster.local
    services:
      cidrBlocks:
      - {{.Params.ServiceCIDR}}
  topology:
    class: tanzukubernetescluster
    controlPlane:
      replicas: {{.Params.ControlPlaneReplicas}}
      {{if eq .Params.OsName "ubuntu"}}metadata:
        annotations:
          run.tanzu.vmware.com/resolve-os-image: os-name=ubuntu{{else}}metadata: {}{{end}}
    variables:
    - name: storageClasses
      value:
      - {{.Params.StorageClass}}
    {{if .Params.AdditionalTrustedCA}}- name: trust
      value:
        additionalTrustedCAs:
        - name: {{$uuid}}
    {{end}}- name: storageClass
      value: {{.Params.StorageClass}}
    - name: defaultStorageClass
      value: {{.Params.StorageClass}}
    - name: controlPlaneVolumes
      value:
      - capacity:
          storage: {{.Params.ControlPlaneEtcdVolumeSize}}
        mountPath: /var/lib/etcd
        name: etcd
        storageClass: {{.Params.StorageClass}}
    - name: vmClass
      value: {{.Params.ControlPlaneVmClass}}
    - name: nodePoolVolumes
      value: []
    {{if .Params.Ntp}}- name: ntp
      value: {{.Params.Ntp}}{{end}}
    version: {{.Params.Version}}
    workers:
      machineDeployments:
      - class: node-pool
        {{if eq .Params.OsName "ubuntu"}}metadata:
          annotations:
            run.tanzu.vmware.com/resolve-os-image: os-name=ubuntu{{else}}metadata: {}{{end}}
        name: md-0
        replicas: {{.Params.WorkerNodePool0Replicas}}
        {{if .Params.WorkerNodePool0Zone}}failureDomain: {{.Params.WorkerNodePool0Zone}}
        {{end}}variables:
          overrides:
          - name: nodePoolVolumes
            value:
            - capacity:
                storage: {{.Params.WorkerContainerdVolumeSize}}
              mountPath: /var/lib/containerd
              name: containerd
              storageClass: {{.Params.StorageClass}}
          {{if ne .Params.WorkerVmClass .Params.ControlPlaneVmClass}}- name: vmClass
            value: {{.Params.WorkerVmClass}}{{end}}
      - class: node-pool
        {{if eq .Params.OsName "ubuntu"}}metadata:
          annotations:
            run.tanzu.vmware.com/resolve-os-image: os-name=ubuntu{{else}}metadata: {}{{end}}
        name: md-1
        replicas: {{.Params.WorkerNodePool1Replicas}}
        {{if .Params.WorkerNodePool1Zone}}failureDomain: {{.Params.WorkerNodePool1Zone}}
        {{end}}variables:
          overrides:
          - name: nodePoolVolumes
            value:
            - capacity:
                storage: {{.Params.WorkerContainerdVolumeSize}}
              mountPath: /var/lib/containerd
              name: containerd
              storageClass: {{.Params.StorageClass}}
          {{if ne .Params.WorkerVmClass .Params.ControlPlaneVmClass}}- name: vmClass
            value: {{.Params.WorkerVmClass}}{{end}}
      - class: node-pool
        {{if eq .Params.OsName "ubuntu"}}metadata:
          annotations:
            run.tanzu.vmware.com/resolve-os-image: os-name=ubuntu{{else}}metadata: {}{{end}}
        name: md-2
        replicas: {{.Params.WorkerNodePool2Replicas}}
        {{if .Params.WorkerNodePool2Zone}}failureDomain: {{.Params.WorkerNodePool2Zone}}
        {{end}}variables:
          overrides:
          - name: nodePoolVolumes
            value:
            - capacity:
                storage: {{.Params.WorkerContainerdVolumeSize}}
              mountPath: /var/lib/containerd
              name: containerd
              storageClass: {{.Params.StorageClass}}
          {{if ne .Params.WorkerVmClass .Params.ControlPlaneVmClass}}- name: vmClass
            value: {{.Params.WorkerVmClass}}{{end}}
`
