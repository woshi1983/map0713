// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	runv1alpha3 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/run.tanzu/v1alpha3"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/sshutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/vc"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlserializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

type TanzuKubernetesGrid struct {
	*k8s.TanzuKubernetesGrid2

	clusterName string
}

func (k *TanzuKubernetesGrid) GetAvailableTanzuKubernetesReleases(selectors ...k8s.Selector) (*runv1alpha3.TanzuKubernetesReleaseList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.GetAvailableTanzuKubernetesReleasesContext(ctx, selectors...)
}

func (w *TanzuKubernetesGrid) GetAvailableTanzuKubernetesReleasesContext(ctx context.Context, selectors ...k8s.Selector) (*runv1alpha3.TanzuKubernetesReleaseList, error) {
	list, err := w.GetTanzuKubernetesReleasesContext(ctx, selectors...)
	if err != nil {
		return nil, err
	}
	items := list.Items
	list.Items = []runv1alpha3.TanzuKubernetesRelease{}

	configValues, err := w.GetEnvironment()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve user-config environment")
	}
	infraProvider := getStringVariable(configValues, ConfigVariableInfraProvider)

LOOP:
	for _, each := range items {
		for _, cond := range each.Status.Conditions {
			switch cond.Type {
			case "Ready":
				if cond.Status != corev1.ConditionTrue {
					continue LOOP
				}
			case "Compatible":
				if cond.Status != corev1.ConditionTrue {
					continue LOOP
				}
			}
		}

		switch infraProvider {
		case InfrastructureProviderVSphere:
			vm, err := w.searchAvailableVsphereTemplateByHint(each.Name, configValues)
			if err != nil || vm == nil {
				continue LOOP
			}
		}

		list.Items = append(list.Items, each)
	}
	return list, nil
}

func (w *TanzuKubernetesGrid) SearchAvailableVsphereTemplateByHint(version string) (*vc.VSphereVirtualMachine, error) {
	return w.searchAvailableVsphereTemplateByHint(version, nil)
}

func (w *TanzuKubernetesGrid) searchAvailableVsphereTemplateByHint(version string, configValues map[string]any) (*vc.VSphereVirtualMachine, error) {
	if configValues == nil {
		var err error
		configValues, err = w.GetEnvironment()
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve user-config environment")
		}
	}

	infraProvider := getStringVariable(configValues, ConfigVariableInfraProvider)
	switch infraProvider {
	case InfrastructureProviderVSphere:
		vcClient, err := w.getVSphereClient(configValues)
		if err != nil {
			return nil, err
		}

		tkrs, err := w.SearchAvailableTKRsByHint(version)
		if err != nil {
			return nil, err
		}
		if len(tkrs) < 1 {
			return nil, fmt.Errorf("unable to find TKR with version %s", version)
		}

		tkr := tkrs[0]
		bom, err := w.GetBOMByTKR(tkr.Name)
		if err != nil {
			return nil, err
		}

		return w.GetVirtualMachineTemplate(vcClient, bom.GetOVAVersions(), tkr.Name, "", getStringVariable(configValues, ConfigVariableVsphereDatacenter), configValues)
	}
	return nil, nil
}

func (w *TanzuKubernetesGrid) GetVirtualMachineTemplate(
	vcClient *vc.Client,
	ovaVersions []string,
	tkrVersion string,
	templateName,
	dc string,
	configValues map[string]any) (*vc.VSphereVirtualMachine, error) {

	dcMOID, err := vcClient.FindDataCenter(context.TODO(), dc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the datacenter MOID")
	}

	vms, err := vcClient.GetVirtualMachineImages(context.TODO(), dcMOID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get K8s VM templates")
	}

	// if VSphereVMTemplateName is provided by user then verify the template with K8s version
	if templateName != "" {
		return ValidateAndGetVirtualMachineTemplateForTKRVersion(vcClient, tkrVersion, ovaVersions, templateName, dc, vms)
	}

	// If templateName is not specified by user then
	// Find the matching template by scanning vsphere inventory

	matchedTemplateVMs, nonTemplateVMs := FindMatchingVirtualMachineTemplate(vms, ovaVersions)

	// filter vm templates based on user provided or default os options
	userProvidedOSOptions := GetUserProvidedOsOptions(configValues)
	vm := SelectTemplateForVsphereProviderBasedonOSOptions(matchedTemplateVMs, userProvidedOSOptions.Name, userProvidedOSOptions.Version, userProvidedOSOptions.Arch)
	if vm != nil {
		return vm, nil
	}

	if len(nonTemplateVMs) != 0 {
		return nil, errors.Errorf("unable to find VM Template associated with TanzuKubernetesRelease %s, but found these VM(s) [%s] that can be used once converted to a VM Template", tkrVersion, strings.Join(nonTemplateVMs, ","))
	}

	return nil, errors.Errorf("unable to find VM Template associated with TanzuKubernetesRelease %s. Please upload at least one VM Template from versions [%v] to continue", tkrVersion, strings.Join(ovaVersions, ","))
}

func (w *TanzuKubernetesGrid) getVSphereTemplateFromClusterClass(clusterClass capiv1beta1.ClusterClass) (string, string, error) {
	ref := clusterClass.Spec.ControlPlane.MachineInfrastructure.Ref
	namespace := ref.Namespace
	if namespace == "" {
		namespace = clusterClass.Namespace
	}
	tpl, err := w.GetVSphereMachineTemplate(namespace, ref.Name)
	if err != nil {
		return "", "", err
	}
	template := tpl.Spec.Template.Spec.Template
	osName := "ubuntu"
	if strings.Contains(strings.ToLower(template), "photon") {
		osName = "photon"
	}
	return template, osName, nil
}

func (w *TanzuKubernetesGrid) SearchAvailableTKRsByHint(version string) ([]runv1alpha3.TanzuKubernetesRelease, error) {
	// remove tinyTKG releases
	tkrs, err := w.GetTanzuKubernetesReleases(k8s.LabelExists("v"+strings.TrimPrefix(version, "v")), k8s.LabelNotExists("tkr.tanzu.vmware.com/tiny"))
	if err != nil {
		return nil, errors.Wrapf(err, "error searching tkr")
	}
	return tkrs.Items, nil
}

func (w *TanzuKubernetesGrid) GetEnvironment() (map[string]any, error) {
	ns, secretName := "tkg-system", "tkg-pkg-tkg-system-values"
	secret, err := w.GetSecret(ns, secretName)
	if err != nil {
		if !k8s.IsErrObjectNotExists(err) {
			return nil, err
		}

		secretName = w.clusterName + "-config-values"
		secret, err = w.GetSecret(ns, secretName)
		if err != nil {
			return nil, err
		}

		p, ok := secret.Data["value"]
		if !ok {
			return nil, fmt.Errorf("unable to find value from secret %s/%s", ns, secretName)
		}
		configValues := make(map[string]interface{})
		err = yaml.Unmarshal(p, &configValues)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to yaml unmashal the data.value of %s secret", secretName)
		}

		err = mergeAKOOVariablesFromSecret(w.Kubernetes, w.clusterName, configValues)
		if err != nil {
			return nil, errors.Wrap(err, "unable to handle the akoo specific variables")
		}
		err = mergeExtraAKOOVariablesFromDeploymentConfig(w.TanzuKubernetesGrid2, configValues)
		if err != nil {
			return nil, errors.Wrap(err, "unable to handle the akoo specific variables")
		}
		return configValues, nil
	}

	p, ok := secret.Data["tkgpackagevalues.yaml"]
	if !ok {
		return nil, fmt.Errorf("unable to find tkgpackagevalues from secret %s/%s", ns, secretName)
	}
	var tkgPackageConfig TKGPackageConfig
	err = yaml.Unmarshal(p, &tkgPackageConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal configuration from secret %s/%s", ns, secretName)
	}
	configValues := tkgPackageConfig.ConfigValues
	err = mergeExtraAKOOVariablesFromDeploymentConfig(w.TanzuKubernetesGrid2, configValues)
	if err != nil {
		return nil, errors.Wrap(err, "unable to handle the akoo specific variables")
	}
	return configValues, nil
}

func (w *TanzuKubernetesGrid) CreateCluster(c *model.Cluster, spec *model.ClusterTKG2MSpec, dryrun bool) ([]byte, error) {
	tkrs, err := w.SearchAvailableTKRsByHint(c.Version)
	if err != nil {
		return nil, err
	}
	if len(tkrs) < 1 {
		return nil, fmt.Errorf("cannot find any TKR with version hint %s", c.Version)
	}
	sort.Slice(tkrs, func(i, j int) bool { return tkrs[i].Name < tkrs[j].Name })
	tkr := tkrs[0]

	configValues, err := w.GetEnvironment()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve user-config environment")
	}

	infraProvider := getStringVariable(configValues, ConfigVariableInfraProvider)
	switch infraProvider {
	case InfrastructureProviderVSphere:
		aviEnabled := getBoolVariable(configValues, ConfigVariableAviEnable)
		if !aviEnabled {
			return nil, fmt.Errorf("avi is not enabled on this cluster")
		}
	default:
		return nil, fmt.Errorf("infrastructure %s is currently not supported", infraProvider)
	}

	clusterClasses, err := w.GetClusterClasses("")
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve cluster classes")
	}
	sort.Slice(clusterClasses.Items, func(i, j int) bool { return clusterClasses.Items[i].Namespace > clusterClasses.Items[j].Namespace })
	sort.SliceStable(clusterClasses.Items, func(i, j int) bool { return clusterClasses.Items[i].Name > clusterClasses.Items[j].Name })
	if len(clusterClasses.Items) < 1 {
		return nil, errors.Wrap(err, "unable to find any available cluster class")
	}

	switch spec.OSName {
	case "ubuntu":
		configValues[ConfigVariableOSName] = spec.OSName
	default:
		configValues[ConfigVariableOSName] = "photon"
	}
	userProvidedOSOptions := GetUserProvidedOsOptions(configValues)
	configValues[ConfigVariableOSVersion] = userProvidedOSOptions.Version
	configValues[ConfigVariableOSArch] = userProvidedOSOptions.Arch

	vmTemplateName := ""
	osName := spec.OSName
	vmTemplate, err := w.searchAvailableVsphereTemplateByHint(c.Version, configValues)
	if err != nil {
		vmTemplateName, osName, err = w.getVSphereTemplateFromClusterClass(clusterClasses.Items[0])
		if err != nil {
			return nil, errors.Wrap(err, "failed to find available templates")
		}
		configValues[ConfigVariableOSName] = osName
		configValues[ConfigVariableOSVersion] = ""
		configValues[ConfigVariableOSArch] = ""
		userProvidedOSOptions = GetUserProvidedOsOptions(configValues)
	}
	if vmTemplate != nil {
		vmTemplateName = vmTemplate.Name
	}
	if vmTemplateName == "" {
		return nil, fmt.Errorf("unable to find any available template")
	}

	cfg := k8s.TanzuKubernetesGrid2ClusterConfig{
		ID:                                 c.ID,
		Version:                            tkr.Name,
		KubernetesVersion:                  tkr.Spec.Kubernetes.Version,
		CNI:                                spec.CNI,
		InfrastructureProvider:             infraProvider,
		VsphereDatacenter:                  getStringVariable(configValues, ConfigVariableVsphereDatacenter),
		VsphereDatastore:                   getStringVariable(configValues, ConfigVariableVsphereDatastore),
		VsphereFolder:                      getStringVariable(configValues, ConfigVariableVsphereFolder),
		VsphereNetwork:                     getStringVariable(configValues, ConfigVariableVsphereNetwork),
		VsphereResourcePool:                getStringVariable(configValues, ConfigVariableVsphereResourcePool),
		VsphereServer:                      getStringVariable(configValues, ConfigVariableVsphereServer),
		VsphereStoragePolicyID:             getStringVariable(configValues, ConfigVariableVsphereStoragePolicyID),
		VsphereTemplate:                    vmTemplateName,
		VsphereTlsThumbprint:               getStringVariable(configValues, ConfigVariableVsphereTLSThumbprint),
		VsphereInsecure:                    getBoolVariable(configValues, ConfigVariableVsphereInsecure),
		VsphereZone:                        getStringVariable(configValues, ConfigVariableVsphereZone),
		VsphereUsername:                    getStringVariable(configValues, ConfigVariableVsphereUsername),
		VspherePassword:                    getStringVariable(configValues, ConfigVariableVspherePassword),
		VsphereSshAuthorizedKeys:           getStringVariable(configValues, ConfigVariableVsphereSSHAuthorizedKey),
		VsphereControlPlaneNumCpus:         getIntVariable(configValues, ConfigVariableVsphereCPNumCpus),
		VsphereControlPlaneMemMiB:          getIntVariable(configValues, ConfigVariableVsphereCPMemMib),
		VsphereControlPlaneDiskGiB:         getIntVariable(configValues, ConfigVariableVsphereCPDiskGib),
		VsphereWorkerNumCpus:               getIntVariable(configValues, ConfigVariableVsphereWorkerNumCpus),
		VsphereWorkerMemMiB:                getIntVariable(configValues, ConfigVariableVsphereWorkerMemMib),
		VsphereWorkerDiskGiB:               getIntVariable(configValues, ConfigVariableVsphereWorkerDiskGib),
		EnableDefaultStorageClass:          true,
		OSName:                             osName,
		ClusterCIDR:                        spec.ClusterCIDR,
		ServiceCIDR:                        spec.ServiceCIDR,
		ClusterClass:                       clusterClasses.Items[0].Name,
		ControlPlaneReplicas:               spec.CtlPlaneReplicas,
		WorkerNodePool0Replicas:            spec.WorkerNodePool0Replicas,
		WorkerNodePool0Zone:                spec.WorkerNodePool0Zone,
		WorkerNodePool1Replicas:            spec.WorkerNodePool1Replicas,
		WorkerNodePool1Zone:                spec.WorkerNodePool1Zone,
		WorkerNodePool2Replicas:            spec.WorkerNodePool2Replicas,
		WorkerNodePool2Zone:                spec.WorkerNodePool2Zone,
		CustomImageRepository:              getStringVariable(configValues, ConfigVariableCustomImageRepository),
		CustomImageRepositoryTlsValidation: !getBoolVariable(configValues, ConfigVariableCustomImageRepositorySkipTLSVerify),
		AdditionalTrustedCA:                string(bytes.Join(spec.TrustedCAs, []byte("\n"))),
		AviEnable:                          true,
		AviControlPlaneHaProvider:          true,
		AviNamespace:                       getStringVariable(configValues, ConfigVariableAviNamespace),
		AviAdminCredentialName:             getStringVariable(configValues, ConfigVariableAviAdminCredentialName),
		AviCaName:                          getStringVariable(configValues, ConfigVariableAviCaName),
		AviCloudName:                       getStringVariable(configValues, ConfigVariableAviCloudName),
		AviManagementClusterControlPlaneVipNetworkName: getStringVariable(configValues, ConfigVariableAviManagementClusterControlPlaneVipNetworkName),
		AviManagementClusterControlPlaneVipNetworkCidr: getStringVariable(configValues, ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR),
		AviManagementClusterVipNetworkName:             getStringVariable(configValues, ConfigVariableAviManagementClusterDataPlaneNetworkName),
		AviManagementClusterVipNetworkCidr:             getStringVariable(configValues, ConfigVariableAviManagementClusterDataPlaneNetworkCIDR),
		AviDataNetworkCidr:                             getStringVariable(configValues, ConfigVariableAviDataPlaneNetworkCIDR),
		AviDataNetwork:                                 getStringVariable(configValues, ConfigVariableAviDataPlaneNetworkName),
		AviController:                                  getStringVariable(configValues, ConfigVariableAviControllerAddress),
		AviControllerVersion:                           getStringVariable(configValues, ConfigVariableAviControllerVersion),
		AviIngressShardVsSize:                          getStringVariable(configValues, ConfigVariableAviIngressShardVsSize),
		AviIngressServiceType:                          getStringVariable(configValues, ConfigVariableAviIngressServiceType),
		AviIngressNodeNetworkList:                      getStringVariable(configValues, ConfigVariableAviIngressNodeNetworkList),
		AviServiceEngineGroup:                          getStringVariable(configValues, ConfigVariableAviServiceEngineGroup),
		NTP:                                            spec.NTP,
	}

	if spec.CtlPlaneNumCpus > 0 {
		cfg.VsphereControlPlaneNumCpus = spec.CtlPlaneNumCpus
	}
	if spec.CtlPlaneMemMiB > 0 {
		cfg.VsphereControlPlaneMemMiB = spec.CtlPlaneMemMiB
	}
	if spec.CtlPlaneDiskGiB > 0 {
		cfg.VsphereControlPlaneDiskGiB = spec.CtlPlaneDiskGiB
	}
	if spec.WorkerNumCpus > 0 {
		cfg.VsphereWorkerNumCpus = spec.WorkerNumCpus
	}
	if spec.WorkerMemMiB > 0 {
		cfg.VsphereWorkerMemMiB = spec.WorkerMemMiB
	}
	if spec.WorkerDiskGiB > 0 {
		cfg.VsphereWorkerDiskGiB = spec.WorkerDiskGiB
	}
	if cfg.WorkerNodePool0Zone != "" || cfg.WorkerNodePool1Zone != "" || cfg.WorkerNodePool2Zone != "" {
		cfg.VsphereZone = ""
	}
	if spec.SSHPrivateKey != nil && len(spec.SSHPrivateKey) > 0 {
		privateKey, err := sshutil.LoadPrivateKey(spec.SSHPrivateKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load private key")
		}
		publicKey, err := sshutil.ExportPublicKey(privateKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to export public key")
		}
		cfg.VsphereSshAuthorizedKeys = string(publicKey)
	}
	if vsphereNetwork := strings.TrimSpace(spec.VsphereNetwork); vsphereNetwork != "" {
		if vsphereNetwork[0] != '/' {
			vsphereNetwork = filepath.Join(cfg.VsphereDatacenter, "network", vsphereNetwork)
		}
		cfg.VsphereNetwork = vsphereNetwork
	}
	if vsphereResourcePool := strings.TrimSpace(spec.VsphereResourcePool); vsphereResourcePool != "" {
		if vsphereResourcePool[0] != '/' {
			vsphereResourcePool = filepath.Join(cfg.VsphereDatacenter, "host", vsphereResourcePool, "Resources")
		}
		cfg.VsphereResourcePool = vsphereResourcePool
	}
	if vsphereFolder := strings.TrimSpace(spec.VsphereFolder); vsphereFolder != "" {
		if vsphereFolder[0] != '/' {
			vsphereFolder = filepath.Join(cfg.VsphereDatacenter, "vm", vsphereFolder)
		}
		cfg.VsphereFolder = vsphereFolder
	}
	if vsphereDatastore := strings.TrimSpace(spec.VsphereDatastore); vsphereDatastore != "" {
		if vsphereDatastore[0] != '/' {
			vsphereDatastore = filepath.Join(cfg.VsphereDatacenter, "datastore", vsphereDatastore)
		}
		cfg.VsphereDatastore = vsphereDatastore
	}

	var buf bytes.Buffer
	if cfg.CustomImageRepository != "" {
		customImageRepositoryCaCertificateBase64 := getStringVariable(configValues, ConfigVariableCustomImageRepositoryCaCertificate)
		if customImageRepositoryCaCertificateBase64 != "" {
			p, err := base64.StdEncoding.DecodeString(customImageRepositoryCaCertificateBase64)
			if err != nil {
				panic(errors.Wrapf(err, "error parsing %s", ConfigVariableCustomImageRepositoryCaCertificate))
			}
			buf.Write(bytes.TrimSpace(p))
			buf.WriteRune('\n')
		}
	}
	if cfg.AdditionalTrustedCA != "" {
		buf.WriteString(cfg.AdditionalTrustedCA)
	}
	cfg.CustomImageRepositoryCaCert = buf.String()

	if spec.AviServiceEngineGroup != "" {
		cfg.AviServiceEngineGroup = spec.AviServiceEngineGroup
	}
	if spec.AviDataNetwork != "" {
		cfg.AviDataNetwork = spec.AviDataNetwork
	}
	if spec.AviDataNetworkCidr != "" {
		cfg.AviDataNetworkCidr = spec.AviDataNetworkCidr
	}
	if doesVariableExists(configValues, ConfigVariableAviDisableStaticRouteSync) {
		cfg.AviDisableStaticRouteSync = util.BoolPtr(getBoolVariable(configValues, ConfigVariableAviDisableStaticRouteSync))
	}
	if doesVariableExists(configValues, ConfigVariableAviIngressDefaultIngressController) {
		cfg.AviIngressDefaultIngressController = util.BoolPtr(getBoolVariable(configValues, ConfigVariableAviIngressDefaultIngressController))
	}
	if doesVariableExists(configValues, ConfigVariableAviDisableIngressClass) {
		cfg.AviDisableIngressClass = util.BoolPtr(getBoolVariable(configValues, ConfigVariableAviDisableIngressClass))
	}

	return w.ApplyCluster(spec.Namespace, c.Name, cfg, false)
}

func (w *TanzuKubernetesGrid) GetClusterClient(namespace, name string) (*k8s.Kubernetes, error) {
	b, err := w.GetClusterKubeConfig(namespace, name)
	if err != nil {
		return nil, err
	}
	return k8s.NewKubernetesFromKubeConfig(b)
}

func (w *TanzuKubernetesGrid) GetClusterKubeConfig(namespace, name string) ([]byte, error) {
	secret, err := w.GetSecret(namespace, fmt.Sprintf("%s-kubeconfig", name))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster kubeconfig")
	}
	v, ok := secret.Data["value"]
	if !ok {
		return nil, errors.New("failed to get kubeconfig")
	}
	return v, nil
}

func (w *TanzuKubernetesGrid) GetClusterInfo(namespace, name string) (*ClusterInfo, error) {
	clusterInfo, err := w.GetClusterObjects(namespace, name)
	if err != nil {
		return nil, err
	}
	info := &ClusterInfo{
		Name:              clusterInfo.Cluster.Name,
		Namespace:         clusterInfo.Cluster.Namespace,
		Roles:             getClusterRoles(clusterInfo.Cluster.Labels),
		Status:            getClusterStatus(clusterInfo),
		ControlPlaneCount: getClusterControlPlaneCount(clusterInfo),
		WorkerCount:       getClusterWorkerCount(clusterInfo),
		K8sVersion:        clusterInfo.KubeadmControlPlane.Spec.Version,
		Labels:            clusterInfo.Cluster.Labels,
	}
	return info, nil
}

func (w *TanzuKubernetesGrid) GetClusterObjects(namespace, name string) (*ClusterObjects, error) {
	var cluster *capiv1beta1.Cluster

	if err := util.Retry(func() error {
		c, err := w.ClusterAPI.GetCluster(namespace, name)
		if err != nil {
			return err
		}
		cluster = c
		return nil
	}, 2, 500*time.Millisecond); err != nil {
		return nil, err
	}
	kubeadmControlPlane, err := w.getKubeadmControlPlaneForCluster(namespace, name)
	if err != nil {
		return nil, err
	}
	machineDeployments, err := w.getMachineDeploymentsForCluster(namespace, name)
	if err != nil {
		return nil, err
	}
	machines, err := w.getMachinesForCluster(namespace, name)
	if err != nil {
		return nil, err
	}
	return &ClusterObjects{
		Cluster:             cluster,
		KubeadmControlPlane: kubeadmControlPlane,
		MachineDeployments:  machineDeployments,
		Machines:            machines,
	}, nil
}

func (w *TanzuKubernetesGrid) GetVSphereClient() (*vc.Client, error) {
	configValues, err := w.GetEnvironment()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve user-config environment")
	}
	return w.getVSphereClient(configValues)
}

func (w *TanzuKubernetesGrid) getVSphereClient(configValues map[string]any) (*vc.Client, error) {
	host := strings.TrimSpace(getStringVariable(configValues, ConfigVariableVsphereServer))
	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}
	vcURL, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse vc host")
	}
	vcURL.Path = "/sdk"

	thumbprint := getStringVariable(configValues, ConfigVariableVsphereTLSThumbprint)
	vsphereInsecure := getBoolVariable(configValues, ConfigVariableVsphereInsecure)

	vcClient, err := vc.NewClient(vcURL, thumbprint, vsphereInsecure)
	if err != nil {
		if e, ok := err.(*url.Error); ok && !vsphereInsecure {
			if _, ok := e.Err.(*tls.CertificateVerificationError); ok {
				vcClient, err = vc.NewClient(vcURL, thumbprint, true)
				if err != nil {
					return nil, errors.Wrap(err, "failed to create vc client")
				}
				goto PROCEED
			}
		}
		return nil, errors.Wrap(err, "failed to create vc client")
	}

PROCEED:
	vcUsername, vcPassword := getStringVariable(configValues, ConfigVariableVsphereUsername), getStringVariable(configValues, ConfigVariableVspherePassword)

	_, err = vcClient.Login(context.TODO(), vcUsername, vcPassword)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login to vSphere")
	}
	return vcClient, nil
}

func (w *TanzuKubernetesGrid) GetBOMByTKR(name string) (*BOMConfiguration, error) {
	tkrConfigMap, err := w.GetConfigMap("tkg-system", name)
	if err != nil {
		return nil, err
	}

	p, ok := tkrConfigMap.BinaryData["bomContent"]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve BOM content")
	}
	cfg := &BOMConfiguration{}
	if err := yaml.Unmarshal(p, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (w *TanzuKubernetesGrid) getCluster(namespace, name string) (*capiv1beta1.Cluster, error) {
	var cluster *capiv1beta1.Cluster
	if err := util.Retry(func() error {
		c, err := w.ClusterAPI.GetCluster(namespace, name)
		if err != nil {
			return err
		}
		cluster = c
		return nil
	}, 2, 500*time.Millisecond); err != nil {
		return nil, err
	}
	return cluster, nil
}

func (w *TanzuKubernetesGrid) getKubeadmControlPlaneForCluster(namespace, name string) (*kubeadmv1beta1.KubeadmControlPlane, error) {
	var kubeadmControlPlane *kubeadmv1beta1.KubeadmControlPlane
	if err := util.Retry(func() error {
		kcps, err := w.ClusterAPI.GetKubeadmControlPlanes(namespace, k8s.LabelEqual(capiv1beta1.ClusterLabelName, name))
		if err != nil {
			return err
		}
		if len(kcps.Items) < 1 {
			return fmt.Errorf("KubeadmControlPlane was not found for cluster %s", strconv.Quote(namespace+"/"+name))
		}
		kubeadmControlPlane = &kcps.Items[0]
		return nil
	}, 2, 500*time.Millisecond); err != nil {
		return nil, err
	}
	return kubeadmControlPlane, nil
}

func (w *TanzuKubernetesGrid) getMachineDeploymentsForCluster(namespace, name string) ([]capiv1beta1.MachineDeployment, error) {
	var machineDeployments []capiv1beta1.MachineDeployment
	if err := util.Retry(func() error {
		mds, err := w.ClusterAPI.GetMachineDeployments(namespace, k8s.LabelEqual(capiv1beta1.ClusterLabelName, name))
		if err != nil {
			return err
		}
		machineDeployments = mds.Items
		return nil
	}, 2, 500*time.Millisecond); err != nil {
		return nil, err
	}
	return machineDeployments, nil
}

func (w *TanzuKubernetesGrid) getMachinesForCluster(namespace, name string) ([]capiv1beta1.Machine, error) {
	var machines []capiv1beta1.Machine
	if err := util.Retry(func() error {
		ms, err := w.ClusterAPI.GetMachines(namespace, k8s.LabelEqual(capiv1beta1.ClusterLabelName, name))
		if err != nil {
			return err
		}
		machines = ms.Items
		return nil
	}, 2, 500*time.Millisecond); err != nil {
		return nil, err
	}
	return machines, nil
}

func (w *TanzuKubernetesGrid) EZUpdateClusterOnDemand(name, version string, spec *model.ClusterTKG2MSpec) (updated bool, err error) {
	requiresUpdate, err := w.DoesClusterRequireUpdate(name, version, spec)
	if err != nil {
		return updated, errors.Wrap(err, "error checking cluster state")
	}
	if requiresUpdate {
		return requiresUpdate, w.EZUpdateCluster(name, version, spec)
	}
	return updated, nil
}

func (w *TanzuKubernetesGrid) ClusterVersion(namespace, name string) (string, error) {
	if namespace == "" || name == "" {
		ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
		defer cancel()

		ver, err := w.VersionContext(ctx)
		if err != nil {
			return "", err
		}
		return ver.GitVersion, nil
	}

	k, err := w.GetClusterClient(namespace, name)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	ver, err := k.VersionContext(ctx)
	if err != nil {
		return "", err
	}
	return ver.GitVersion, nil
}

func (w *TanzuKubernetesGrid) DoesControlPlaneVersionMatch(namespace, name, version string) (desired, current bool, err error) {
	kcp, err := w.getKubeadmControlPlaneForCluster(namespace, name)
	if err != nil {
		return desired, current, err
	}
	v, err := w.ClusterVersion(namespace, name)
	if err != nil {
		return desired, current, err
	}
	return textutil.CompareMajorMinorPatchVersion(kcp.Spec.Version, version), textutil.CompareMajorMinorPatchVersion(v, version), nil
}

func (w *TanzuKubernetesGrid) DoesClusterRequireUpdate(name, version string, spec *model.ClusterTKG2MSpec) (bool, error) {
	cluster, err := w.getCluster(spec.Namespace, name)
	if err != nil {
		return false, err
	}

	// k8s version
	if !textutil.CompareMajorMinorPatchVersion(cluster.Spec.Topology.Version, version) ||
		(cluster.Spec.Topology.ControlPlane.Replicas == nil && spec.CtlPlaneReplicas > 0) ||
		int64(*cluster.Spec.Topology.ControlPlane.Replicas) != spec.CtlPlaneReplicas {
		return true, nil
	}

	// ctlPlane
	defaultVarMap := make(map[string]apiextensionsv1.JSON)
	workerNodePool0VarMap := make(map[string]apiextensionsv1.JSON)
	workerNodePool1VarMap := make(map[string]apiextensionsv1.JSON)
	workerNodePool2VarMap := make(map[string]apiextensionsv1.JSON)
	for _, each := range cluster.Spec.Topology.Variables {
		defaultVarMap[each.Name] = each.Value
		workerNodePool0VarMap[each.Name] = each.Value
		workerNodePool1VarMap[each.Name] = each.Value
		workerNodePool2VarMap[each.Name] = each.Value
	}

	v, ok := defaultVarMap["controlPlane"]
	if !ok {
		return true, nil
	}
	ctlPlane, err := parseClusterControlPlaneVariable(v)
	if err != nil {
		return false, errors.Wrap(err, "getting controlPlane")
	}
	if ctlPlane.Machine.NumCPUs != spec.CtlPlaneNumCpus ||
		ctlPlane.Machine.MemoryMiB != spec.CtlPlaneMemMiB ||
		ctlPlane.Machine.DiskGiB != spec.CtlPlaneDiskGiB {
		return true, nil
	}

	// workers
	if len(cluster.Spec.Topology.Workers.MachineDeployments) < 3 {
		return true, nil
	}

	var workerNodePool0, workerNodePool1, workerNodePool2 *capiv1beta1.MachineDeploymentTopology
	mds := cluster.Spec.Topology.Workers.MachineDeployments
	sort.Slice(mds, func(i, j int) bool { return mds[i].Name < mds[j].Name })
	workerNodePool0, workerNodePool1, workerNodePool2 = &mds[0], &mds[1], &mds[2]

	// md-0
	if (workerNodePool0.Replicas == nil && spec.WorkerNodePool0Replicas > 0) ||
		int64(*workerNodePool0.Replicas) != spec.WorkerNodePool0Replicas {
		return true, nil
	}

	// md-1
	if (workerNodePool1.Replicas == nil && spec.WorkerNodePool1Replicas > 0) ||
		int64(*workerNodePool1.Replicas) != spec.WorkerNodePool1Replicas {
		return true, nil
	}

	// md-2
	if (workerNodePool2.Replicas == nil && spec.WorkerNodePool2Replicas > 0) ||
		int64(*workerNodePool2.Replicas) != spec.WorkerNodePool2Replicas {
		return true, nil
	}

	v, ok = defaultVarMap["worker"]
	if !ok {
		return true, nil
	}
	worker, err := parseClusterWorkerVariable(v)
	if err != nil {
		return false, errors.Wrap(err, "getting worker")
	}
	if worker.Machine.NumCPUs != spec.WorkerNumCpus ||
		worker.Machine.MemoryMiB != spec.WorkerMemMiB ||
		worker.Machine.DiskGiB != spec.WorkerDiskGiB {
		return true, nil
	}

	return false, nil
}

func (w *TanzuKubernetesGrid) EZUpdateCluster(name, version string, spec *model.ClusterTKG2MSpec) error {
	cluster, err := w.getCluster(spec.Namespace, name)
	if err != nil {
		return err
	}

	patch := &capiv1beta1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: capiv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cluster.Name,
			Namespace:   cluster.Namespace,
			Annotations: cluster.Annotations,
			Labels:      cluster.Labels,
		},
		Spec: *cluster.Spec.DeepCopy(),
	}

	// version
	tkrs, err := w.SearchAvailableTKRsByHint(version)
	if err != nil {
		return err
	}
	if len(tkrs) < 1 {
		return fmt.Errorf("cannot find any TKR with version hint %s", version)
	}
	sort.Slice(tkrs, func(i, j int) bool { return tkrs[i].Name < tkrs[j].Name })
	tkr := tkrs[0]
	patch.Spec.Topology.Version = tkr.Spec.Kubernetes.Version

	_, err = w.PatchClusterBootstrap(spec.Namespace, name, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"annotations":{"tkg.tanzu.vmware.com/add-missing-fields-from-tkr":%s}}}`, strconv.Quote(tkr.Name))))
	if err != nil {
		return errors.Wrap(err, "error updating tkr annotation on ClusterBootstrap")
	}

	// ctlPlaneReplicas
	patch.Spec.Topology.ControlPlane.Replicas = util.Int32Ptr(int32(spec.CtlPlaneReplicas))
	patch.Spec.Topology.Variables, _ = popClusterVariables(patch.Spec.Topology.Variables, "controlPlane")
	p, err := json.Marshal(controlPlaneConfigVariable{
		Machine: machineSpec{
			NumCPUs:   spec.CtlPlaneNumCpus,
			MemoryMiB: spec.CtlPlaneMemMiB,
			DiskGiB:   spec.CtlPlaneDiskGiB,
		},
	})
	if err != nil {
		return errors.Wrap(err, "error updating controlPlane variable")
	}
	patch.Spec.Topology.Variables = append(patch.Spec.Topology.Variables, capiv1beta1.ClusterVariable{
		Name:  "controlPlane",
		Value: apiextensionsv1.JSON{Raw: p},
	})

	// workers
	var workerNodePool0, workerNodePool1, workerNodePool2 *capiv1beta1.MachineDeploymentTopology
	if len(patch.Spec.Topology.Workers.MachineDeployments) < 3 {
		return fmt.Errorf("worker MachineDeployments must be greater than 2")
	}
	mds := patch.Spec.Topology.Workers.MachineDeployments
	sort.Slice(mds, func(i, j int) bool { return mds[i].Name < mds[j].Name })
	workerNodePool0, workerNodePool1, workerNodePool2 = &mds[0], &mds[1], &mds[2]

	patch.Spec.Topology.Variables, _ = popClusterVariables(patch.Spec.Topology.Variables, "worker")
	p, err = json.Marshal(workerConfigVariable{
		Machine: machineSpec{
			NumCPUs:   spec.WorkerNumCpus,
			MemoryMiB: spec.WorkerMemMiB,
			DiskGiB:   spec.WorkerDiskGiB,
		},
	})
	if err != nil {
		return errors.Wrap(err, "error updating worker variable")
	}
	patch.Spec.Topology.Variables = append(patch.Spec.Topology.Variables, capiv1beta1.ClusterVariable{
		Name:  "worker",
		Value: apiextensionsv1.JSON{Raw: p},
	})

	// workerNodePool0Replicas
	workerNodePool0.Replicas = util.Int32Ptr(int32(spec.WorkerNodePool0Replicas))
	// workerNodePool0Zone
	if spec.WorkerNodePool0Zone != "" {
		workerNodePool0.FailureDomain = util.StringPtr(spec.WorkerNodePool0Zone)
	} else {
		workerNodePool0.FailureDomain = nil
	}

	// workerNodePool1Replicas
	workerNodePool1.Replicas = util.Int32Ptr(int32(spec.WorkerNodePool1Replicas))
	// workerNodePool1Zone
	if spec.WorkerNodePool1Zone != "" {
		workerNodePool1.FailureDomain = util.StringPtr(spec.WorkerNodePool1Zone)
	} else {
		workerNodePool1.FailureDomain = nil
	}

	// workerNodePool2Replicas
	workerNodePool2.Replicas = util.Int32Ptr(int32(spec.WorkerNodePool2Replicas))
	// workerNodePool2Zone
	if spec.WorkerNodePool2Zone != "" {
		workerNodePool2.FailureDomain = util.StringPtr(spec.WorkerNodePool2Zone)
	} else {
		workerNodePool2.FailureDomain = nil
	}

	var buf bytes.Buffer
	ser := yamlserializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	if err := ser.Encode(patch, &buf); err != nil {
		return errors.Wrap(err, "generating cluster patch")
	}

	// NOTE: this is a work around that resolve the issue that TKG2m cannot handle empty
	// failureDomain value.
	obj := &unstructured.Unstructured{Object: make(map[string]interface{})}
	if _, _, err = ser.Decode(buf.Bytes(), &schema.GroupVersionKind{Group: capiv1beta1.GroupVersion.Group, Version: capiv1beta1.GroupVersion.Version, Kind: "Cluster"}, obj); err != nil {
		return errors.Wrap(err, "evaluating cluster patch")
	}
	unstructured.RemoveNestedField(obj.Object, "spec", "topology", "workers")
	workerPatch := workersTopologyPatch{
		MachineDeployments: []machineDeploymentTopologyPatch{
			generatePatchFromMachineDeploymentTopology(workerNodePool0),
			generatePatchFromMachineDeploymentTopology(workerNodePool1),
			generatePatchFromMachineDeploymentTopology(workerNodePool2),
		},
	}
	o, err := workerPatch.ConvertToUnstructedObject()
	if err != nil {
		return errors.Wrap(err, "generating cluster patch")
	}
	unstructured.SetNestedField(obj.Object, o, "spec", "topology", "workers")
	buf.Reset()
	if err := ser.Encode(obj, &buf); err != nil {
		return errors.Wrap(err, "generating cluster patch")
	}

	err = w.Apply(buf.Bytes(), mapd.GlobalConfig.HTTPClientTimeout)
	if err != nil {
		return errors.Wrap(err, "applying cluster")
	}
	return nil
}

func NewTanzuKubernetesGrid(kubeconfig []byte) (*TanzuKubernetesGrid, error) {
	tkg, err := k8s.NewTanzuKubernetesGrid2FromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	clusterName, err := k8s.GetClusterNameFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	return &TanzuKubernetesGrid{TanzuKubernetesGrid2: tkg, clusterName: clusterName}, nil
}

func getBoolVariable(configValues map[string]any, key string) bool {
	if configValues == nil {
		return false
	}
	v, ok := configValues[key]
	if !ok {
		return false
	}
	switch v := v.(type) {
	case string:
		switch strings.ToLower(v) {
		case "1", "true":
			return true
		}
	case bool:
		return v
	}
	return false
}

func getIntVariable(configValues map[string]any, key string) int64 {
	if configValues == nil {
		return 0
	}
	v, ok := configValues[key]
	if !ok {
		return 0
	}
	switch v := v.(type) {
	case string:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	case int64:
		return v
	case int:
		return int64(v)
	}
	return 0
}

func getStringVariable(configValues map[string]any, key string) string {
	if configValues == nil {
		return ""
	}
	v, ok := configValues[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func doesVariableExists(configValues map[string]any, key string) bool {
	_, ok := configValues[key]
	return ok
}

func getAKOOAddonSecretValues(cli *k8s.Kubernetes, clusterName string) ([]byte, bool, error) {
	akoOperatorAddonName := fmt.Sprintf("%s-%s-addon", clusterName, AkoOperatorName)
	secret, err := cli.GetSecret(TkgNamespace, akoOperatorAddonName)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return []byte{}, false, nil
		}
		return []byte{}, false, err
	}
	p, ok := secret.Data["values.yaml"]
	if !ok {
		return []byte{}, false, nil
	}
	return p, true, nil
}

func mergeAKOOVariablesFromSecret(cli *k8s.Kubernetes, clusterName string, configValues map[string]interface{}) error {
	// retrieve the akoo variables from legacy addon secret.
	akooAddonSecretValues, found, err := getAKOOAddonSecretValues(cli, clusterName)
	if err != nil {
		return errors.Wrap(err, "unable to get akoo addon secret values")
	}

	if !found || len(akooAddonSecretValues) == 0 {
		return nil
	}

	akoOperatorPackage := &AkoOperatorPackage{}
	err = yaml.Unmarshal(akooAddonSecretValues, akoOperatorPackage)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal the akoo addon secret values.yaml")
	}

	configValues[ConfigVariableAviEnable] = akoOperatorPackage.AkoOperatorPackageValues.AviEnable
	configValues[ConfigVariableAviControllerAddress] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviControllerAddress
	configValues[ConfigVariableAviControllerUsername] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviControllerUsername
	configValues[ConfigVariableAviControllerPassword] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviControllerPassword
	configValues[ConfigVariableAviControllerCA] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviControllerCA
	configValues[ConfigVariableAviCloudName] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviCloudName
	configValues[ConfigVariableAviServiceEngineGroup] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviServiceEngineGroup
	configValues[ConfigVariableAviManagementClusterServiceEngineGroup] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviManagementClusterServiceEngineGroup
	configValues[ConfigVariableAviDataPlaneNetworkName] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviDataPlaneNetworkName
	configValues[ConfigVariableAviDataPlaneNetworkCIDR] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviDataPlaneNetworkCIDR
	configValues[ConfigVariableAviControlPlaneNetworkName] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviControlPlaneNetworkName
	configValues[ConfigVariableAviControlPlaneNetworkCIDR] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviControlPlaneNetworkCIDR
	configValues[ConfigVariableAviManagementClusterDataPlaneNetworkName] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviManagementClusterDataPlaneNetworkName
	configValues[ConfigVariableAviManagementClusterDataPlaneNetworkCIDR] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR
	configValues[ConfigVariableAviManagementClusterControlPlaneVipNetworkName] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviManagementClusterControlPlaneVipNetworkName
	configValues[ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviManagementClusterControlPlaneVipNetworkCIDR
	configValues[ConfigVariableVsphereHaProvider] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviControlPlaneHaProvider
	configValues[ConfigVariableAviIngressNodeNetworkList] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviIngressNodeNetworkList
	configValues[ConfigVariableClusterName] = clusterName
	configValues[ConfigVariableAviLabels] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviLabels
	configValues[ConfigVariableAviNSXTT1Router] = akoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig.AviNSXTT1Router

	return nil
}

func mergeExtraAKOOVariablesFromDeploymentConfig(cli *k8s.TanzuKubernetesGrid2, configValues map[string]interface{}) error {
	akoCfg, err := cli.GetAKODeploymentConfig("install-ako-for-all")
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil
		}
		return errors.Wrap(err, "unable to get global AKODeploymentConfig")
	}

	configValues[ConfigVariableAviNamespace] = akoCfg.Spec.AdminCredentialRef.Namespace
	configValues[ConfigVariableAviAdminCredentialName] = akoCfg.Spec.AdminCredentialRef.Name
	configValues[ConfigVariableAviCaName] = akoCfg.Spec.CertificateAuthorityRef.Name
	configValues[ConfigVariableAviDisableStaticRouteSync] = akoCfg.Spec.ExtraConfigs.DisableStaticRouteSync
	configValues[ConfigVariableAviIngressDefaultIngressController] = akoCfg.Spec.ExtraConfigs.IngressConfigs.DefaultIngressController
	configValues[ConfigVariableAviDisableIngressClass] = akoCfg.Spec.ExtraConfigs.IngressConfigs.DisableIngressClass
	configValues[ConfigVariableAviIngressShardVsSize] = akoCfg.Spec.ExtraConfigs.IngressConfigs.ShardVSSize
	configValues[ConfigVariableAviIngressServiceType] = akoCfg.Spec.ExtraConfigs.IngressConfigs.ServiceType
	return nil
}

// TKGPackageConfig defines TKG package configuration
type TKGPackageConfig struct {
	Metadata                     Metadata                     `yaml:"metadata"`
	ConfigValues                 map[string]interface{}       `yaml:"configvalues"`
	FrameworkPackage             FrameworkPackage             `yaml:"frameworkPackage"`
	ClusterClassPackage          ClusterClassPackage          `yaml:"clusterclassPackage"`
	TKRSourceControllerPackage   TKRSourceControllerPackage   `yaml:"tkrSourceControllerPackage"`
	CoreManagementPluginsPackage CoreManagementPluginsPackage `yaml:"coreManagementPluginsPackage"`
	AkoOperatorPackage           AkoOperatorPackage           `yaml:"akoOperatorPackage"`
}

// Metadata specifies metadata as part of TKG package config
type Metadata struct {
	InfraProvider string `yaml:"infraProvider"`
}

type TKRSourceControllerPackage struct {
	NamespaceForPackageInstallation  string                           `yaml:"namespaceForPackageInstallation,omitempty"`
	VersionConstraints               string                           `yaml:"versionConstraints,omitempty"`
	TKRSourceControllerPackageValues TKRSourceControllerPackageValues `yaml:"tkrSourceControllerPackageValues,omitempty"`
}

type TKRSourceControllerPackageValues struct {
	Namespace            string                                     `yaml:"namespace,omitempty"`
	CreateNamespace      string                                     `yaml:"createNamespace,omitempty"`
	VersionConstraints   string                                     `yaml:"versionConstraints,omitempty"`
	BomImagePath         string                                     `yaml:"bomImagePath,omitempty"`
	BomMetadataImagePath string                                     `yaml:"bomMetadataImagePath,omitempty"`
	TKRRepoImagePath     string                                     `yaml:"tkrRepoImagePath,omitempty"`
	DefaultCompatibleTKR string                                     `yaml:"defaultCompatibleTKR,omitempty"`
	CaCerts              string                                     `yaml:"caCerts,omitempty"`
	ImageRepo            string                                     `yaml:"imageRepository,omitempty"`
	Deployment           TKRSourceControllerPackageValuesDeployment `yaml:"deployment,omitempty"`
}

type TKRSourceControllerPackageValuesDeployment struct {
	HttpProxy  string `yaml:"httpProxy,omitempty"`
	HttpsProxy string `yaml:"httpsProxy,omitempty"`
	NoProxy    string `yaml:"noProxy,omitempty"`
}

type FrameworkPackage struct {
	NamespaceForPackageInstallation string                     `yaml:"namespaceForPackageInstallation,omitempty"`
	VersionConstraints              string                     `yaml:"versionConstraints,omitempty"`
	FeaturegatePackageValues        FeaturegatePackageValues   `yaml:"featureGatesPackageValues,omitempty"`
	TKRServicePackageValues         TKRServicePackageValues    `yaml:"tkrServicePackageValues,omitempty"`
	CLIPluginsPackageValues         CLIPluginsPackageValues    `yaml:"clipluginsPackageValues,omitempty"`
	AddonsManagerPackageValues      AddonsManagerPackageValues `yaml:"addonsManagerPackageValues,omitempty"`
	TanzuAuthPackageValues          TanzuAuthPackageValues     `yaml:"tanzuAuthPackageValues,omitempty"`
}

type ClusterClassPackage struct {
	NamespaceForPackageInstallation string                         `yaml:"namespaceForPackageInstallation,omitempty"`
	VersionConstraints              string                         `yaml:"versionConstraints,omitempty"`
	ClusterClassInfraPackageValues  ClusterClassInfraPackageValues `yaml:"clusterclassInfraPackageValues,omitempty"`
}

type AkoOperatorPackage struct {
	AkoOperatorPackageValues AkoOperatorPackageValues `yaml:"akoOperator,omitempty"`
}

type AddonsFeatureGates struct {
	ClusterBootstrapController bool `yaml:"clusterBootstrapController,omitempty"`
	PackageInstallStatus       bool `yaml:"packageInstallStatus,omitempty"`
}

type TanzuAddonsManager struct {
	FeatureGates AddonsFeatureGates `yaml:"featureGates,omitempty"`
}
type CoreManagementPluginsPackage struct {
	NamespaceForPackageInstallation string `yaml:"namespaceForPackageInstallation,omitempty"`
	VersionConstraints              string `yaml:"versionConstraints,omitempty"`
}

type AddonsManagerPackageValues struct {
	VersionConstraints string             `yaml:"versionConstraints,omitempty"`
	TanzuAddonsManager TanzuAddonsManager `yaml:"tanzuAddonsManager,omitempty"`
}

type FeaturegatePackageValues struct {
	Namespace          string `yaml:"namespace,omitempty"`
	CreateNamespace    string `yaml:"createNamespace,omitempty"`
	VersionConstraints string `yaml:"versionConstraints,omitempty"`
}

type TKRServicePackageValues struct {
	Namespace          string                            `yaml:"namespace,omitempty"`
	CreateNamespace    string                            `yaml:"createNamespace,omitempty"`
	VersionConstraints string                            `yaml:"versionConstraints,omitempty"`
	Deployment         TKRServicePackageValuesDeployment `yaml:"deployment,omitempty"`
}

type TKRServicePackageValuesDeployment struct {
	HttpProxy  string `yaml:"httpProxy,omitempty"`
	HttpsProxy string `yaml:"httpsProxy,omitempty"`
	NoProxy    string `yaml:"noProxy,omitempty"`
}

type CLIPluginsPackageValues struct {
	Namespace          string `yaml:"namespace,omitempty"`
	CreateNamespace    string `yaml:"createNamespace,omitempty"`
	VersionConstraints string `yaml:"versionConstraints,omitempty"`
	DeployCLIPluginCRD bool   `yaml:"deployCLIPluginCRD,omitempty"`
}

type ClusterClassInfraPackageValues struct {
	Namespace          string `yaml:"namespace,omitempty"`
	CreateNamespace    string `yaml:"createNamespace,omitempty"`
	VersionConstraints string `yaml:"versionConstraints,omitempty"`
}

type TanzuAuthPackageValues struct {
	Namespace          string `yaml:"namespace,omitempty"`
	CreateNamespace    string `yaml:"createNamespace,omitempty"`
	VersionConstraints string `yaml:"versionConstraints,omitempty"`
}

// AkoOperatorPackageValues
type AkoOperatorPackageValues struct {
	AviEnable          bool              `yaml:"avi_enable,omitempty"`
	ClusterName        string            `yaml:"cluster_name,omitempty"`
	OnBootstrapCluster bool              `yaml:"on_bootstrap_cluster,omitempty"`
	AkoOperatorConfig  AkoOperatorConfig `yaml:"config,omitempty"`
}

// NodeNetwork
type NodeNetwork struct {
	NetworkName string   `yaml:"networkName,omitempty" json:"networkName,omitempty"`
	Cidrs       []string `yaml:"cidrs,omitempty" json:"cidrs,omitempty"`
}

// AkoOperatorConfig
type AkoOperatorConfig struct {
	AviControllerAddress                           string `yaml:"avi_controller,omitempty"`
	AviControllerVersion                           string `yaml:"avi_controller_version,omitempty"`
	AviControllerUsername                          string `yaml:"avi_username,omitempty"`
	AviControllerPassword                          string `yaml:"avi_password,omitempty"`
	AviControllerCA                                string `yaml:"avi_ca_data_b64,omitempty"`
	AviCloudName                                   string `yaml:"avi_cloud_name,omitempty"`
	AviServiceEngineGroup                          string `yaml:"avi_service_engine_group,omitempty"`
	AviManagementClusterServiceEngineGroup         string `yaml:"avi_management_cluster_service_engine_group,omitempty"`
	AviDataPlaneNetworkName                        string `yaml:"avi_data_network,omitempty"`
	AviDataPlaneNetworkCIDR                        string `yaml:"avi_data_network_cidr,omitempty"`
	AviControlPlaneNetworkName                     string `yaml:"avi_control_plane_network,omitempty"`
	AviControlPlaneNetworkCIDR                     string `yaml:"avi_control_plane_network_cidr,omitempty"`
	AviManagementClusterDataPlaneNetworkName       string `yaml:"avi_management_cluster_vip_network_name,omitempty"`
	AviManagementClusterDataPlaneNetworkCIDR       string `yaml:"avi_management_cluster_vip_network_cidr,omitempty"`
	AviManagementClusterControlPlaneVipNetworkName string `yaml:"avi_management_cluster_control_plane_vip_network_name,omitempty"`
	AviManagementClusterControlPlaneVipNetworkCIDR string `yaml:"avi_management_cluster_control_plane_vip_network_cidr,omitempty"`
	AviLabels                                      string `yaml:"avi_labels,omitempty"`
	AviNSXTT1Router                                string `yaml:"avi_nsxt_t1_lr,omitempty"`
	AviControlPlaneHaProvider                      bool   `yaml:"avi_control_plane_ha_provider,omitempty"`
	AviIngressNodeNetworkList                      string `yaml:"avi_ingress_node_network_list,omitempty"`
}

const (
	vmOptionOsName    = "name"
	vmOptionOsVersion = "version"
	vmOptionOsArch    = "arch"
)

// GetUserProvidedOsOptions returns user provided os options
func GetUserProvidedOsOptions(configValues map[string]any) OSInfo {
	defaultOsInfo := GetDefaultUbuntuOsOptions(getStringVariable(configValues, ConfigVariableInfraProvider))

	osName := getStringVariable(configValues, ConfigVariableOSName)
	if osName == "" {
		osName = defaultOsInfo.Name
	}
	switch osName {
	case "ubuntu":
	case "photon":
		defaultOsInfo = GetDefaultPhotonOsOptions(getStringVariable(configValues, ConfigVariableInfraProvider))
	}
	osVersion := getStringVariable(configValues, ConfigVariableOSVersion)
	if osVersion == "" {
		osVersion = defaultOsInfo.Version
	}
	osArch := getStringVariable(configValues, ConfigVariableOSArch)
	if osArch == "" {
		osArch = defaultOsInfo.Arch
	}
	return OSInfo{
		Name:    osName,
		Version: osVersion,
		Arch:    osArch,
	}
}

// GetDefaultUbuntuOsOptions returns default OS option based on providerType
func GetDefaultUbuntuOsOptions(providerType string) OSInfo {
	switch providerType {
	case InfrastructureProviderVSphere:
		return OSInfo{Name: "ubuntu", Version: "20.04", Arch: "amd64"}

	case InfrastructureProviderAWS:
		return OSInfo{Name: "ubuntu", Version: "20.04", Arch: "amd64"}

	case InfrastructureProviderAzure:
		return OSInfo{Name: "ubuntu", Version: "20.04", Arch: "amd64"}
	}
	return OSInfo{}
}

// GetDefaultPhotonOsOptions returns default OS option based on providerType
func GetDefaultPhotonOsOptions(providerType string) OSInfo {
	switch providerType {
	case InfrastructureProviderVSphere:
		return OSInfo{Name: "photon", Version: "3", Arch: "amd64"}

	case InfrastructureProviderAWS:
		return OSInfo{Name: "photon", Version: "3", Arch: "amd64"}

	case InfrastructureProviderAzure:
		return OSInfo{Name: "photon", Version: "3", Arch: "amd64"}
	}
	return OSInfo{}
}

type controlPlaneConfigVariable struct {
	Machine machineSpec `json:"machine"`
}

type workerConfigVariable struct {
	Machine machineSpec `json:"machine"`
}

type machineSpec struct {
	DiskGiB   int64 `json:"diskGiB,omitempty"`
	MemoryMiB int64 `json:"memoryMiB,omitempty"`
	NumCPUs   int64 `json:"numCPUs,omitempty"`
}

func parseClusterControlPlaneVariable(val apiextensionsv1.JSON) (*controlPlaneConfigVariable, error) {
	if string(val.Raw) == "null" {
		return nil, nil
	}
	cfg := &controlPlaneConfigVariable{}
	if err := json.Unmarshal(val.Raw, cfg); err != nil {
		return nil, errors.Wrap(err, "error parsing variable value")
	}
	return cfg, nil
}

func parseClusterWorkerVariable(val apiextensionsv1.JSON) (*workerConfigVariable, error) {
	if string(val.Raw) == "null" {
		return nil, nil
	}
	cfg := &workerConfigVariable{}
	if err := json.Unmarshal(val.Raw, cfg); err != nil {
		return nil, errors.Wrap(err, "error parsing variable value")
	}
	return cfg, nil
}

// workersTopologyPatch represents the different sets of worker nodes in the cluster.
type workersTopologyPatch struct {
	// MachineDeployments is a list of machine deployments in the cluster.
	// +optional
	MachineDeployments []machineDeploymentTopologyPatch `json:"machineDeployments"`
}

func (p *workersTopologyPatch) ConvertToUnstructedObject() (map[string]interface{}, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]interface{})
	if err = json.Unmarshal(b, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

type machineDeploymentTopologyPatch struct {
	// Metadata is the metadata applied to the machines of the MachineDeployment.
	// At runtime this metadata is merged with the corresponding metadata from the ClusterClass.
	// +optional
	Metadata capiv1beta1.ObjectMeta `json:"metadata"`

	// Class is the name of the MachineDeploymentClass used to create the set of worker nodes.
	// This should match one of the deployment classes defined in the ClusterClass object
	// mentioned in the `Cluster.Spec.Class` field.
	Class string `json:"class"`

	// Name is the unique identifier for this MachineDeploymentTopology.
	// The value is used with other unique identifiers to create a MachineDeployment's Name
	// (e.g. cluster's name, etc). In case the name is greater than the allowed maximum length,
	// the values are hashed together.
	Name string `json:"name"`

	// FailureDomain is the failure domain the machines will be created in.
	// Must match a key in the FailureDomains map stored on the cluster object.
	// +optional
	FailureDomain *string `json:"failureDomain"`

	// Replicas is the number of worker nodes belonging to this set.
	// If the value is nil, the MachineDeployment is created without the number of Replicas (defaulting to zero)
	// and it's assumed that an external entity (like cluster autoscaler) is responsible for the management
	// of this value.
	// +optional
	Replicas *int32 `json:"replicas"`

	// NodeDrainTimeout is the total amount of time that the controller will spend on draining a node.
	// The default value is 0, meaning that the node can be drained without any time limitations.
	// NOTE: NodeDrainTimeout is different from `kubectl drain --timeout`
	// +optional
	NodeDrainTimeout *metav1.Duration `json:"nodeDrainTimeout"`

	// Variables can be used to customize the MachineDeployment through patches.
	// +optional
	Variables *capiv1beta1.MachineDeploymentVariables `json:"variables"`
}

func generatePatchFromMachineDeploymentTopology(v *capiv1beta1.MachineDeploymentTopology) machineDeploymentTopologyPatch {
	return machineDeploymentTopologyPatch{
		Metadata:         v.Metadata,
		Class:            v.Class,
		Name:             v.Name,
		FailureDomain:    v.FailureDomain,
		Replicas:         v.Replicas,
		NodeDrainTimeout: v.NodeDrainTimeout.DeepCopy(),
		Variables:        v.Variables.DeepCopy(),
	}
}
