// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	networkingv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/networking/v1alpha1"
	runv1alpha3 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/run.tanzu/v1alpha3"
	networkingversioned "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/networking/clientset/versioned"
	networkingscheme "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/networking/clientset/versioned/scheme"
	runversioned "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/run.tanzu/clientset/versioned"
	runscheme "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client/run.tanzu/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
)

type TanzuKubernetesGrid2 struct {
	*ClusterAPI
	KappController *KappController
}

func (w *TanzuKubernetesGrid2) ApplyCluster(namespace, name string, cfg TanzuKubernetesGrid2ClusterConfig, dryrun bool) ([]byte, error) {
	config := ClusterTemplateConfig{
		Name:      name,
		Namespace: namespace,
		Params: map[string]string{
			"Id":                                 cfg.ID.String(),
			"Version":                            cfg.Version,
			"KubernetesVersion":                  cfg.KubernetesVersion,
			"Cni":                                cfg.CNI, // `antrea` or `calico`
			"InfrastructureProvider":             cfg.InfrastructureProvider,
			"VsphereControlPlaneEndpoint":        cfg.VsphereControlPlaneEndpoint,
			"VsphereDatacenter":                  cfg.VsphereDatacenter,
			"VsphereDatastore":                   cfg.VsphereDatastore,
			"VsphereFolder":                      cfg.VsphereFolder,
			"VsphereNetwork":                     cfg.VsphereNetwork,
			"VsphereResourcePool":                cfg.VsphereResourcePool,
			"VsphereServer":                      cfg.VsphereServer,
			"VsphereStoragePolicyID":             cfg.VsphereStoragePolicyID,
			"VsphereTemplate":                    cfg.VsphereTemplate,
			"VsphereTlsThumbprint":               cfg.VsphereTlsThumbprint,
			"VsphereInsecure":                    strconv.FormatBool(cfg.VsphereInsecure),
			"VsphereZone":                        cfg.VsphereZone,
			"VsphereUsername":                    cfg.VsphereUsername,
			"VspherePassword":                    cfg.VspherePassword,
			"VsphereSshAuthorizedKeys":           cfg.VsphereSshAuthorizedKeys,
			"VsphereControlPlaneNumCpus":         strconv.FormatInt(cfg.VsphereControlPlaneNumCpus, 10),
			"VsphereControlPlaneMemMiB":          strconv.FormatInt(cfg.VsphereControlPlaneMemMiB, 10),
			"VsphereControlPlaneDiskGiB":         strconv.FormatInt(cfg.VsphereControlPlaneDiskGiB, 10),
			"VsphereWorkerNumCpus":               strconv.FormatInt(cfg.VsphereWorkerNumCpus, 10),
			"VsphereWorkerMemMiB":                strconv.FormatInt(cfg.VsphereWorkerMemMiB, 10),
			"VsphereWorkerDiskGiB":               strconv.FormatInt(cfg.VsphereWorkerDiskGiB, 10),
			"EnableDefaultStorageClass":          strconv.FormatBool(cfg.EnableDefaultStorageClass),
			"OsName":                             cfg.OSName, // `photon` or `ubuntu`
			"ClusterCIDR":                        cfg.ClusterCIDR,
			"ServiceCIDR":                        cfg.ServiceCIDR,
			"ClusterClass":                       cfg.ClusterClass,
			"ControlPlaneReplicas":               strconv.FormatInt(cfg.ControlPlaneReplicas, 10),
			"WorkerNodePool0Replicas":            strconv.FormatInt(cfg.WorkerNodePool0Replicas, 10),
			"WorkerNodePool0Zone":                cfg.WorkerNodePool0Zone,
			"WorkerNodePool1Replicas":            strconv.FormatInt(cfg.WorkerNodePool1Replicas, 10),
			"WorkerNodePool1Zone":                cfg.WorkerNodePool1Zone,
			"WorkerNodePool2Replicas":            strconv.FormatInt(cfg.WorkerNodePool2Replicas, 10),
			"WorkerNodePool2Zone":                cfg.WorkerNodePool2Zone,
			"AdditionalTrustedCA":                cfg.AdditionalTrustedCA,
			"CustomImageRepository":              cfg.CustomImageRepository,
			"CustomImageRepositoryTlsValidation": strconv.FormatBool(cfg.CustomImageRepositoryTlsValidation),
			"CustomImageRepositoryCaCert":        cfg.CustomImageRepositoryCaCert,
			"AviEnable":                          strconv.FormatBool(cfg.AviEnable),
			"AviControlPlaneHaProvider":          strconv.FormatBool(cfg.AviControlPlaneHaProvider),
			"AviNamespace":                       cfg.AviNamespace,
			"AviAdminCredentialName":             cfg.AviAdminCredentialName,
			"AviCaName":                          cfg.AviCaName,
			"AviCloudName":                       cfg.AviCloudName,
			"AviManagementClusterControlPlaneVipNetworkName": cfg.AviManagementClusterControlPlaneVipNetworkName,
			"AviManagementClusterControlPlaneVipNetworkCidr": cfg.AviManagementClusterControlPlaneVipNetworkCidr,
			"AviManagementClusterVipNetworkName":             cfg.AviManagementClusterVipNetworkName,
			"AviManagementClusterVipNetworkCidr":             cfg.AviManagementClusterVipNetworkCidr,
			"AviDataNetworkCidr":                             cfg.AviDataNetworkCidr,
			"AviDataNetwork":                                 cfg.AviDataNetwork,
			"AviController":                                  cfg.AviController,
			"AviControllerVersion":                           cfg.AviControllerVersion,
			"AviIngressShardVsSize":                          cfg.AviIngressShardVsSize,
			"AviIngressServiceType":                          cfg.AviIngressServiceType,
			"AviIngressNodeNetworkList":                      cfg.AviIngressNodeNetworkList,
			"AviServiceEngineGroup":                          cfg.AviServiceEngineGroup,
			"Ntp":                                            cfg.NTP,
		},
	}
	if cfg.AviDisableStaticRouteSync != nil {
		config.Params["AviDisableStaticRouteSync"] = strconv.FormatBool(*cfg.AviDisableStaticRouteSync)
	}
	if cfg.AviIngressDefaultIngressController != nil {
		config.Params["AviIngressDefaultIngressController"] = strconv.FormatBool(*cfg.AviIngressDefaultIngressController)
	}
	if cfg.AviDisableIngressClass != nil {
		config.Params["AviDisableIngressClass"] = strconv.FormatBool(*cfg.AviDisableIngressClass)
	}
	return w.ClusterAPI.ApplyCluster(config)
}

func (w *TanzuKubernetesGrid2) GetTanzuKubernetesRelease(name string) (*runv1alpha3.TanzuKubernetesRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesReleaseContext(ctx, name)
}

func (w *TanzuKubernetesGrid2) GetTanzuKubernetesReleaseContext(ctx context.Context, name string) (*runv1alpha3.TanzuKubernetesRelease, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha3().TanzuKubernetesReleases().Get(ctx, name, metav1.GetOptions{})
}

func (w *TanzuKubernetesGrid2) GetTanzuKubernetesReleases(selectors ...Selector) (*runv1alpha3.TanzuKubernetesReleaseList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetTanzuKubernetesReleasesContext(ctx, selectors...)
}

func (w *TanzuKubernetesGrid2) GetTanzuKubernetesReleasesContext(ctx context.Context, selectors ...Selector) (*runv1alpha3.TanzuKubernetesReleaseList, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)
	return clientset.RunV1alpha3().TanzuKubernetesReleases().List(ctx, opts)
}

func (w *TanzuKubernetesGrid2) PatchClusterBootstrap(namespace, name string, pt types.PatchType, patch []byte) (*runv1alpha3.ClusterBootstrap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.PatchClusterBootstrapContext(ctx, namespace, name, pt, patch)
}

func (w *TanzuKubernetesGrid2) PatchClusterBootstrapContext(ctx context.Context, namespace, name string, pt types.PatchType, patch []byte) (*runv1alpha3.ClusterBootstrap, error) {
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

func (w *TanzuKubernetesGrid2) GetClusterBootstrap(namespace, name string) (*runv1alpha3.ClusterBootstrap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetClusterBootstrapContext(ctx, namespace, name)
}

func (w *TanzuKubernetesGrid2) GetClusterBootstrapContext(ctx context.Context, namespace, name string) (*runv1alpha3.ClusterBootstrap, error) {
	clientset, err := runversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	return clientset.RunV1alpha3().ClusterBootstraps(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (w *TanzuKubernetesGrid2) GetClusterBootstraps(namespace string, selectors ...Selector) (*runv1alpha3.ClusterBootstrapList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetClusterBootstrapsContext(ctx, namespace, selectors...)
}

func (w *TanzuKubernetesGrid2) GetClusterBootstrapsContext(ctx context.Context, namespace string, selectors ...Selector) (*runv1alpha3.ClusterBootstrapList, error) {
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

func (w *TanzuKubernetesGrid2) GetAKODeploymentConfig(name string) (*networkingv1alpha1.AKODeploymentConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetAKODeploymentConfigContext(ctx, name)
}

func (w *TanzuKubernetesGrid2) GetAKODeploymentConfigContext(ctx context.Context, name string) (*networkingv1alpha1.AKODeploymentConfig, error) {
	clientset, err := networkingversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	return clientset.NetworkingV1alpha1().AKODeploymentConfigs().Get(ctx, name, metav1.GetOptions{})
}

func (w *TanzuKubernetesGrid2) GetAKODeploymentConfigs(selectors ...Selector) (*networkingv1alpha1.AKODeploymentConfigList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetAKODeploymentConfigsContext(ctx, selectors...)
}

func (w *TanzuKubernetesGrid2) GetAKODeploymentConfigsContext(ctx context.Context, selectors ...Selector) (*networkingv1alpha1.AKODeploymentConfigList, error) {
	clientset, err := networkingversioned.NewForConfig(w.ClusterAPI.Config)
	if err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{}
	opts.FieldSelector, opts.LabelSelector = qualifySelectors(selectors...)

	result := &networkingv1alpha1.AKODeploymentConfigList{}
	err = clientset.NetworkingV1alpha1().RESTClient().Get().
		Resource("akodeploymentconfigs").
		VersionedParams(&opts, networkingscheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

func (w *TanzuKubernetesGrid2) GetVSphereMachineTemplate(namespace, name string) (*capvv1beta1.VSphereMachineTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return w.GetVSphereMachineTemplateContext(ctx, namespace, name)
}

func (w *TanzuKubernetesGrid2) GetVSphereMachineTemplateContext(ctx context.Context, namespace, name string) (*capvv1beta1.VSphereMachineTemplate, error) {
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

func NewTanzuKubernetesGrid2(k *Kubernetes) (*TanzuKubernetesGrid2, error) {
	capi, err := NewClusterAPI(k, tkg2ClusterTemplate)
	if err != nil {
		return nil, err
	}
	return &TanzuKubernetesGrid2{
		ClusterAPI:     capi,
		KappController: NewKappController(k),
	}, nil
}

func NewTanzuKubernetesGrid2FromKubeConfig(kubeconfig []byte) (*TanzuKubernetesGrid2, error) {
	k, err := NewKubernetesFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	return NewTanzuKubernetesGrid2(k)
}

type TanzuKubernetesGrid2ClusterConfig struct {
	ID                                             uuid.UUID
	Version                                        string
	KubernetesVersion                              string
	CNI                                            string
	InfrastructureProvider                         string
	VsphereControlPlaneEndpoint                    string
	VsphereDatacenter                              string
	VsphereDatastore                               string
	VsphereFolder                                  string
	VsphereNetwork                                 string
	VsphereResourcePool                            string
	VsphereServer                                  string
	VsphereStoragePolicyID                         string
	VsphereTemplate                                string
	VsphereTlsThumbprint                           string
	VsphereInsecure                                bool
	VsphereZone                                    string
	VsphereUsername                                string
	VspherePassword                                string
	VsphereSshAuthorizedKeys                       string
	VsphereControlPlaneNumCpus                     int64
	VsphereControlPlaneMemMiB                      int64
	VsphereControlPlaneDiskGiB                     int64
	VsphereWorkerNumCpus                           int64
	VsphereWorkerMemMiB                            int64
	VsphereWorkerDiskGiB                           int64
	EnableDefaultStorageClass                      bool
	OSName                                         string
	ClusterCIDR                                    string
	ServiceCIDR                                    string
	ClusterClass                                   string
	ControlPlaneReplicas                           int64
	WorkerNodePool0Replicas                        int64
	WorkerNodePool0Zone                            string
	WorkerNodePool1Replicas                        int64
	WorkerNodePool1Zone                            string
	WorkerNodePool2Replicas                        int64
	WorkerNodePool2Zone                            string
	AdditionalTrustedCA                            string
	CustomImageRepository                          string
	CustomImageRepositoryTlsValidation             bool
	CustomImageRepositoryCaCert                    string
	AviEnable                                      bool
	AviControlPlaneHaProvider                      bool
	AviNamespace                                   string
	AviAdminCredentialName                         string
	AviCaName                                      string
	AviCloudName                                   string
	AviManagementClusterControlPlaneVipNetworkName string
	AviManagementClusterControlPlaneVipNetworkCidr string
	AviManagementClusterVipNetworkName             string
	AviManagementClusterVipNetworkCidr             string
	AviDataNetworkCidr                             string
	AviDataNetwork                                 string
	AviController                                  string
	AviControllerVersion                           string
	AviDisableStaticRouteSync                      *bool
	AviIngressDefaultIngressController             *bool
	AviDisableIngressClass                         *bool
	AviIngressShardVsSize                          string
	AviIngressServiceType                          string
	AviIngressNodeNetworkList                      string
	AviServiceEngineGroup                          string
	NTP                                            string
}

const tkg2ClusterTemplate = `{{$uuid := ""}}{{if .Params.Id}}{{$uuid = .Params.Id}}{{else}}{{$uuid = uuidgen}}{{end}}{{if eq .Params.Cni "calico"}}---
apiVersion: cni.tanzu.vmware.com/v1alpha1
kind: CalicoConfig
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  calico:
    config:
      skipCNIBinaries: true
      vethMTU: 0
{{end}}{{if eq .Params.InfrastructureProvider "vsphere"}}{{if .Params.VsphereControlPlaneEndpoint}}---
apiVersion: cpi.tanzu.vmware.com/v1alpha1
kind: VSphereCPIConfig
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  vsphereCPI:
    mode: vsphereCPI
    tlsCipherSuites: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
    vmNetwork:
      excludeExternalSubnetCidr: {{.Params.VsphereControlPlaneEndpoint}}/32
      excludeInternalSubnetCidr: {{.Params.VsphereControlPlaneEndpoint}}/32
{{end}}---
apiVersion: csi.tanzu.vmware.com/v1alpha1
kind: VSphereCSIConfig
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  vsphereCSI:
    config:
      datacenter: "{{.Params.VsphereDatacenter}}"
      httpProxy: ""
      httpsProxy: ""
      noProxy: ""
      region: ""
      {{if .Params.VsphereTlsThumbprint}}tlsThumbprint: "{{.Params.VsphereTlsThumbprint}}"
      {{else if .Params.VsphereInsecure}}insecureFlag: true
      {{end}}useTopologyCategories: false
      zone: {{if .Params.VsphereZone}}{{.Params.VsphereZone}}{{else}}""{{end}}
    mode: vsphereCSI
---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
stringData:
  password: {{.Params.VspherePassword}}
  username: {{.Params.VsphereUsername}}
{{end}}{{if .Params.AdditionalTrustedCA}}---
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
apiVersion: run.tanzu.vmware.com/v1alpha3
kind: ClusterBootstrap
metadata:
  annotations:
    tkg.tanzu.vmware.com/add-missing-fields-from-tkr: {{.Params.Version}}
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  additionalPackages:
  - refName: metrics-server*
  - refName: secretgen-controller*
  - refName: tkg-storageclass*
    valuesFrom:
      inline:
        infraProvider: {{if (bool .Params.EnableDefaultStorageClass)}}{{.Params.InfrastructureProvider | quote}}
        {{if .Params.VsphereStoragePolicyID}}VSPHERE_STORAGE_POLICY_ID: {{.Params.VsphereStoragePolicyID | quote}}{{end}}
        {{else}}""{{end}}
  - refName: pinniped*
{{if eq .Params.Cni "calico"}}  cni:
    refName: calico*
    valuesFrom:
      providerRef:
        apiGroup: cni.tanzu.vmware.com
        kind: CalicoConfig
        name: {{.Name}}
{{end}}{{if eq .Params.InfrastructureProvider "vsphere"}}{{if .Params.VsphereControlPlaneEndpoint}}  cpi:
    refName: vsphere-cpi*
    valuesFrom:
      providerRef:
        apiGroup: cpi.tanzu.vmware.com
        kind: VSphereCPIConfig
        name: {{.Name}}
{{end}}  csi:
    refName: vsphere-csi*
    valuesFrom:
      providerRef:
        apiGroup: csi.tanzu.vmware.com
        kind: VSphereCSIConfig
        name: {{.Name}}
{{end}}  kapp:
    refName: kapp-controller*
{{if (bool .Params.AviEnable)}}---
apiVersion: networking.tkg.tanzu.vmware.com/v1alpha1
kind: AKODeploymentConfig
metadata:
  name: {{$uuid}}
  labels:
    tkg.tanzu.vmware.com/cluster-name: {{.Name}}
    cluster-id.map.tanzu.vmware.com/workload: {{$uuid}}
spec:
  {{if and .Params.AviAdminCredentialName .Params.AviNamespace}}adminCredentialRef:
    name: {{.Params.AviAdminCredentialName}}
    namespace: {{.Params.AviNamespace}}
  {{end}}{{if and .Params.AviCaName .Params.AviNamespace}}certificateAuthorityRef:
    name: {{.Params.AviCaName}}
    namespace: {{.Params.AviNamespace}}
  {{end}}{{if .Params.AviCloudName}}cloudName: {{.Params.AviCloudName}}
  {{end}}clusterSelector:
    matchLabels:
      tkg.tanzu.vmware.com/cluster-name: {{.Name}}
      cluster-id.map.tanzu.vmware.com/workload: {{$uuid}}
  controlPlaneNetwork:
    {{if and .Params.AviManagementClusterControlPlaneVipNetworkName .Params.AviManagementClusterControlPlaneVipNetworkCidr}}
    cidr: {{.Params.AviManagementClusterControlPlaneVipNetworkCidr}}
    name: {{.Params.AviManagementClusterControlPlaneVipNetworkName}}
    {{else if and .Params.AviManagementClusterVipNetworkName .Params.AviManagementClusterVipNetworkCidr}}
    cidr: {{.Params.AviManagementClusterVipNetworkCidr}}
    name: {{.Params.AviManagementClusterVipNetworkName}}
    {{else}}
    cidr: {{.Params.AviDataNetworkCidr}}
    name: {{.Params.AviDataNetwork}}
    {{end}}
  {{if .Params.AviController}}controller: {{.Params.AviController}}
  {{end}}{{if .Params.AviControllerVersion}}controllerVersion: {{.Params.AviControllerVersion}}
  {{end}}dataNetwork:
    cidr: {{.Params.AviDataNetworkCidr}}
    name: {{.Params.AviDataNetwork}}
  extraConfigs:
    {{if .Params.AviDisableStaticRouteSync}}disableStaticRouteSync: {{.Params.AviDisableStaticRouteSync}}
    {{end}}ingress:
      {{if .Params.AviIngressDefaultIngressController}}defaultIngressController: {{.Params.AviIngressDefaultIngressController}}
      {{end}}{{if .Params.AviDisableIngressClass}}disableIngressClass: {{.Params.AviDisableIngressClass}}
      {{end}}{{if .Params.AviIngressShardVsSize}}
      shardVSSize: {{.Params.AviIngressShardVsSize}}
      {{end}}{{if .Params.AviIngressServiceType}}
      serviceType: {{.Params.AviIngressServiceType}}
      {{end}}{{if or .Params.AviIngressNodeNetworkList .Params.VsphereNetwork}}nodeNetworkList:
      {{if .Params.AviIngressNodeNetworkList}}
      {{json .Params.AviIngressNodeNetworkList}}
      {{else}}
      - networkName: {{last (split "/" .Params.VsphereNetwork)}}
      {{end}}{{end}}
  {{if .Params.AviServiceEngineGroup}}serviceEngineGroup: {{.Params.AviServiceEngineGroup}}{{end}}
{{end}}---
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  annotations:
    osInfo: {{if eq .Params.OsName "ubuntu"}}ubuntu,20.04{{else}}photon,3{{end}},amd64
{{if .Params.VsphereControlPlaneEndpoint}}    tkg.tanzu.vmware.com/cluster-controlplane-endpoint: {{.Params.VsphereControlPlaneEndpoint}}
{{end}}    tkg/plan: prod
  labels:
    tkg.tanzu.vmware.com/cluster-name: {{.Name}}
    cluster-id.map.tanzu.vmware.com/workload: {{$uuid}}
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - {{.Params.ClusterCIDR}}
    services:
      cidrBlocks:
      - {{.Params.ServiceCIDR}}
  topology:
    class: {{.Params.ClusterClass}}
    controlPlane:
      metadata:
        annotations:
          run.tanzu.vmware.com/resolve-os-image: image-type=ova,os-name={{if eq .Params.OsName "ubuntu"}}{{.Params.OsName}}{{else}}photon{{end}}
      replicas: {{.Params.ControlPlaneReplicas}}
    variables:
    - name: controlPlaneCertificateRotation
      value:
        activate: true
        daysBefore: 90
    - name: auditLogging
      value:
        enabled: false
    - name: podSecurityStandard
      value:
        audit: baseline
        deactivated: false
        warn: baseline
    {{if .Params.CustomImageRepository}}- name: imageRepository
      value:
        host: {{.Params.CustomImageRepository | quote}}
        tlsCertificateValidation: {{.Params.CustomImageRepositoryTlsValidation}}
    {{end}}{{if or .Params.CustomImageRepositoryCaCert .Params.AdditionalTrustedCA}}- name: trust
      value:
        additionalTrustedCAs:
        {{if .Params.CustomImageRepositoryCaCert}}- name: imageRepository
          data: {{.Params.CustomImageRepositoryCaCert | atob}}
        {{end}}{{if .Params.AdditionalTrustedCA}}- name: rootca
          data: {{.Params.AdditionalTrustedCA | atob}}
        {{end}}
    {{end}}{{if .Params.CustomImageRepository}}- name: additionalImageRegistries
        value:
        - host: {{.Params.CustomImageRepository | quote}}
          skipTlsVerify: {{not .Params.CustomImageRepositoryTlsValidation}}
          {{if .Params.CustomImageRepositoryCaCert}}caCert: {{.Params.CustomImageRepositoryCaCert | atob}}{{end}}
    {{end}}{{if eq .Params.InfrastructureProvider "vsphere"}}- name: aviAPIServerHAProvider
      value: {{if and (bool .Params.AviEnable) (bool .Params.AviControlPlaneHaProvider)}}true{{else}}false{{end}}
    {{if .Params.VsphereControlPlaneEndpoint}}- name: apiServerEndpoint
      value: {{.Params.VsphereControlPlaneEndpoint}}
    {{end}}- name: vcenter
      value:
        cloneMode: fullClone
        datacenter: {{.Params.VsphereDatacenter | quote}}
        datastore: {{.Params.VsphereDatastore | quote}}
        folder: {{.Params.VsphereFolder | quote}}
        network: {{.Params.VsphereNetwork | quote}}
        resourcePool: {{.Params.VsphereResourcePool | quote}}
        server: {{.Params.VsphereServer | quote}}
        storagePolicyID: {{.Params.VsphereStoragePolicyID | quote}}
        template: {{.Params.VsphereTemplate | quote}}
        tlsThumbprint: {{.Params.VsphereTlsThumbprint | quote}}
    - name: user
      value:
        sshAuthorizedKeys:
        - {{.Params.VsphereSshAuthorizedKeys}}
    - name: controlPlane
      value:
        machine:
          diskGiB: {{.Params.VsphereControlPlaneDiskGiB}}
          memoryMiB: {{.Params.VsphereControlPlaneMemMiB}}
          numCPUs: {{.Params.VsphereControlPlaneNumCpus}}
    - name: worker
      value:
        machine:
          diskGiB: {{.Params.VsphereWorkerDiskGiB}}
          memoryMiB: {{.Params.VsphereWorkerMemMiB}}
          numCPUs: {{.Params.VsphereWorkerNumCpus}}
    {{end}}{{if .Params.Ntp}}- name: ntpServers
      value:{{- range (split "," .Params.Ntp) }}
      - {{.}}
      {{- end }}
    {{end}}version: "{{.Params.KubernetesVersion}}"
    workers:
      machineDeployments:
      - class: tkg-worker
        metadata:
          annotations:
            run.tanzu.vmware.com/resolve-os-image: image-type=ova,os-name={{if eq .Params.OsName "ubuntu"}}{{.Params.OsName}}{{else}}photon{{end}}
        name: md-0
        replicas: {{.Params.WorkerNodePool0Replicas}}
        {{if .Params.WorkerNodePool0Zone}}failureDomain: {{.Params.WorkerNodePool0Zone}}{{end}}
      - class: tkg-worker
        metadata:
          annotations:
            run.tanzu.vmware.com/resolve-os-image: image-type=ova,os-name={{if eq .Params.OsName "ubuntu"}}{{.Params.OsName}}{{else}}photon{{end}}
        name: md-1
        replicas: {{.Params.WorkerNodePool1Replicas}}
        {{if .Params.WorkerNodePool1Zone}}failureDomain: {{.Params.WorkerNodePool1Zone}}{{end}}
      - class: tkg-worker
        metadata:
          annotations:
            run.tanzu.vmware.com/resolve-os-image: image-type=ova,os-name={{if eq .Params.OsName "ubuntu"}}{{.Params.OsName}}{{else}}photon{{end}}
        name: md-2
        replicas: {{.Params.WorkerNodePool2Replicas}}
        {{if .Params.WorkerNodePool2Zone}}failureDomain: {{.Params.WorkerNodePool2Zone}}{{end}}

`
