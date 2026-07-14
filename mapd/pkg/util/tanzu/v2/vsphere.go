// Copyright (c) 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/vc"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

type VSphereWithTanzu struct {
	*k8s.VSphereWithTanzu2
}

func (k *VSphereWithTanzu) GetUserVisibleNamespaces(selectors ...k8s.Selector) (*corev1.NamespaceList, error) {
	sels := append([]k8s.Selector{k8s.LabelExists("vSphereClusterID")}, selectors...)
	return k.GetNamespaces(sels...)
}

func (k *VSphereWithTanzu) GetUserVisibleNamespacesContext(ctx context.Context, selectors ...k8s.Selector) (*corev1.NamespaceList, error) {
	sels := append([]k8s.Selector{k8s.LabelExists("vSphereClusterID")}, selectors...)
	return k.GetNamespacesContext(ctx, sels...)
}

func (k *VSphereWithTanzu) GetAvailableTanzuKubernetesReleases(selectors ...k8s.Selector) (*runv1alpha3.TanzuKubernetesReleaseList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	return k.GetAvailableTanzuKubernetesReleasesContext(ctx, selectors...)
}

func (k *VSphereWithTanzu) GetAvailableTanzuKubernetesReleasesContext(ctx context.Context, selectors ...k8s.Selector) (*runv1alpha3.TanzuKubernetesReleaseList, error) {
	sels := append([]k8s.Selector{k8s.LabelNotExists("run.tanzu.vmware.com/legacy-tkr")}, selectors...)
	list, err := k.GetTanzuKubernetesReleasesContext(ctx, sels...)
	if err != nil {
		return nil, err
	}
	items := list.Items
	list.Items = []runv1alpha3.TanzuKubernetesRelease{}
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
		list.Items = append(list.Items, each)
	}
	return list, nil
}

func (w *VSphereWithTanzu) SearchAvailableTKRsByHint(version string) ([]runv1alpha3.TanzuKubernetesRelease, error) {
	tkrs, err := w.GetAvailableTanzuKubernetesReleases(k8s.LabelExists("v" + strings.TrimPrefix(version, "v")))
	if err != nil {
		return nil, errors.Wrapf(err, "error searching tkr")
	}
	return tkrs.Items, nil
}

func (w *VSphereWithTanzu) CreateCluster(c *model.Cluster, spec *model.ClusterTKG2SSpec, dryrun bool) ([]byte, error) {
	tkrs, err := w.SearchAvailableTKRsByHint(c.Version)
	if err != nil {
		return nil, err
	}
	if len(tkrs) < 1 {
		return nil, fmt.Errorf("cannot find any TKR with version hint %s", c.Version)
	}
	sort.Slice(tkrs, func(i, j int) bool { return tkrs[i].Name < tkrs[j].Name })
	cfg := k8s.VSphereWithTanzu2ClusterConfig{
		ID:                            c.ID,
		Version:                       tkrs[0].Name,
		StorageClass:                  spec.StorageClass,
		AdditionalTrustedCA:           string(bytes.Join(spec.TrustedCAs, []byte("\n"))),
		OSName:                        spec.OSName,
		CNI:                           spec.CNI,
		ControlPlaneReplicas:          spec.CtlPlaneReplicas,
		ControlPlaneVmClass:           spec.CtlPlaneVmClass,
		ControlPlaneEtcdVolumeSizeGiB: spec.CtlPlaneEtcdVolumeSizeGi,
		WorkerNodePool0Replicas:       spec.WorkerNodePool0Replicas,
		WorkerNodePool0Zone:           spec.WorkerNodePool0Zone,
		WorkerNodePool1Replicas:       spec.WorkerNodePool1Replicas,
		WorkerNodePool1Zone:           spec.WorkerNodePool1Zone,
		WorkerNodePool2Replicas:       spec.WorkerNodePool2Replicas,
		WorkerNodePool2Zone:           spec.WorkerNodePool2Zone,
		WorkerVmClass:                 spec.WorkerVmClass,
		WorkerContainerdVolumeSizeGiB: spec.WorkerContainerdVolumeSizeGi,
		ClusterCIDR:                   spec.ClusterCIDR,
		ServiceCIDR:                   spec.ServiceCIDR,
		NTP:                           spec.NTP,
	}
	return w.ApplyCluster(spec.Namespace, c.Name, cfg, false)
}

func (w *VSphereWithTanzu) GetClusterInfo(namespace, name string) (*ClusterInfo, error) {
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

func (w *VSphereWithTanzu) GetClusterObjects(namespace, name string) (*ClusterObjects, error) {
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

func (w *VSphereWithTanzu) getCluster(namespace, name string) (*capiv1beta1.Cluster, error) {
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

func (w *VSphereWithTanzu) getKubeadmControlPlaneForCluster(namespace, name string) (*kubeadmv1beta1.KubeadmControlPlane, error) {
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

func (w *VSphereWithTanzu) getMachineDeploymentsForCluster(namespace, name string) ([]capiv1beta1.MachineDeployment, error) {
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

func (w *VSphereWithTanzu) getMachinesForCluster(namespace, name string) ([]capiv1beta1.Machine, error) {
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

func (w *VSphereWithTanzu) EZUpdateClusterOnDemand(name, version string, spec *model.ClusterTKG2SSpec) (updated bool, err error) {
	requiresUpdate, err := w.DoesClusterRequireUpdate(name, version, spec)
	if err != nil {
		return updated, errors.Wrap(err, "error checking cluster state")
	}
	if requiresUpdate {
		return requiresUpdate, w.EZUpdateCluster(name, version, spec)
	}
	return updated, nil
}

func (w *VSphereWithTanzu) GetClusterClient(namespace, name string) (*k8s.Kubernetes, error) {
	b, err := w.GetClusterKubeConfig(namespace, name)
	if err != nil {
		return nil, err
	}
	return k8s.NewKubernetesFromKubeConfig(b)
}

func (w *VSphereWithTanzu) GetClusterKubeConfig(namespace, name string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	secret, err := w.GetSecretContext(ctx, namespace, fmt.Sprintf("%s-kubeconfig", name))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster kubeconfig")
	}
	v, ok := secret.Data["value"]
	if !ok {
		return nil, errors.New("failed to get kubeconfig")
	}
	return v, nil
}

func (w *VSphereWithTanzu) ClusterVersion(namespace, name string) (string, error) {
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

func (w *VSphereWithTanzu) DoesControlPlaneVersionMatch(namespace, name, version string) (desired, current bool, err error) {
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

func (w *VSphereWithTanzu) DoesClusterRequireUpdate(name, version string, spec *model.ClusterTKG2SSpec) (bool, error) {
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
	v, ok := defaultVarMap["vmClass"]
	if !ok {
		return true, nil
	}
	vmClass, err := parseClusterStringVariable(v)
	if err != nil {
		return false, errors.Wrap(err, "getting vmClass")
	}
	if vmClass != spec.CtlPlaneVmClass {
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
	for _, each := range workerNodePool0.Variables.Overrides {
		workerNodePool0VarMap[each.Name] = each.Value
	}
	v, ok = workerNodePool0VarMap["vmClass"]
	if !ok {
		return true, nil
	}
	vmClass, err = parseClusterStringVariable(v)
	if err != nil {
		return false, errors.Wrap(err, "getting vmClass")
	}
	if vmClass != spec.WorkerVmClass {
		return true, nil
	}

	// md-1
	if (workerNodePool1.Replicas == nil && spec.WorkerNodePool1Replicas > 0) ||
		int64(*workerNodePool1.Replicas) != spec.WorkerNodePool1Replicas {
		return true, nil
	}
	for _, each := range workerNodePool1.Variables.Overrides {
		workerNodePool1VarMap[each.Name] = each.Value
	}
	v, ok = workerNodePool1VarMap["vmClass"]
	if !ok {
		return true, nil
	}
	vmClass, err = parseClusterStringVariable(v)
	if err != nil {
		return false, errors.Wrap(err, "getting vmClass")
	}
	if vmClass != spec.WorkerVmClass {
		return true, nil
	}
	// md-2
	if (workerNodePool2.Replicas == nil && spec.WorkerNodePool2Replicas > 0) ||
		int64(*workerNodePool2.Replicas) != spec.WorkerNodePool2Replicas {
		return true, nil
	}
	for _, each := range workerNodePool2.Variables.Overrides {
		workerNodePool2VarMap[each.Name] = each.Value
	}
	v, ok = workerNodePool2VarMap["vmClass"]
	if !ok {
		return true, nil
	}
	vmClass, err = parseClusterStringVariable(v)
	if err != nil {
		return false, errors.Wrap(err, "getting vmClass")
	}
	if vmClass != spec.WorkerVmClass {
		return true, nil
	}

	return false, nil
}

func (w *VSphereWithTanzu) EZUpdateCluster(name, version string, spec *model.ClusterTKG2SSpec) error {
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
	patch.Spec.Topology.Version = tkr.Name

	_, err = w.PatchClusterBootstrap(spec.Namespace, name, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"annotations":{"tkg.tanzu.vmware.com/add-missing-fields-from-tkr":%s}}}`, strconv.Quote(tkr.Name))))
	if err != nil {
		return errors.Wrap(err, "error updating tkr annotation on ClusterBootstrap")
	}

	// ctlPlaneReplicas
	patch.Spec.Topology.ControlPlane.Replicas = util.Int32Ptr(int32(spec.CtlPlaneReplicas))

	// ctlPlaneVmClass
	patch.Spec.Topology.Variables, _ = popClusterVariables(patch.Spec.Topology.Variables, "vmClass")
	patch.Spec.Topology.Variables = append(patch.Spec.Topology.Variables, capiv1beta1.ClusterVariable{
		Name:  "vmClass",
		Value: apiextensionsv1.JSON{Raw: []byte(strconv.Quote(spec.CtlPlaneVmClass))},
	})

	// workers
	var workerNodePool0, workerNodePool1, workerNodePool2 *capiv1beta1.MachineDeploymentTopology
	if len(patch.Spec.Topology.Workers.MachineDeployments) < 3 {
		return fmt.Errorf("worker MachineDeployments must be greater than 2")
	}
	mds := patch.Spec.Topology.Workers.MachineDeployments
	sort.Slice(mds, func(i, j int) bool { return mds[i].Name < mds[j].Name })
	workerNodePool0, workerNodePool1, workerNodePool2 = &mds[0], &mds[1], &mds[2]

	// workerNodePool0Replicas
	workerNodePool0.Replicas = util.Int32Ptr(int32(spec.WorkerNodePool0Replicas))
	// workerNodePool0Zone
	workerNodePool0.FailureDomain = util.StringPtr(spec.WorkerNodePool0Zone)
	// workerVmClass
	workerNodePool0.Variables.Overrides, _ = popClusterVariables(workerNodePool0.Variables.Overrides, "vmClass")
	workerNodePool0.Variables.Overrides = append(workerNodePool0.Variables.Overrides, capiv1beta1.ClusterVariable{
		Name:  "vmClass",
		Value: apiextensionsv1.JSON{Raw: []byte(strconv.Quote(spec.WorkerVmClass))},
	})

	// workerNodePool1Replicas
	workerNodePool1.Replicas = util.Int32Ptr(int32(spec.WorkerNodePool1Replicas))
	// workerNodePool1Zone
	workerNodePool1.FailureDomain = util.StringPtr(spec.WorkerNodePool1Zone)
	// workerVmClass
	workerNodePool1.Variables.Overrides, _ = popClusterVariables(workerNodePool1.Variables.Overrides, "vmClass")
	workerNodePool1.Variables.Overrides = append(workerNodePool1.Variables.Overrides, capiv1beta1.ClusterVariable{
		Name:  "vmClass",
		Value: apiextensionsv1.JSON{Raw: []byte(strconv.Quote(spec.WorkerVmClass))},
	})

	// workerNodePool2Replicas
	workerNodePool2.Replicas = util.Int32Ptr(int32(spec.WorkerNodePool2Replicas))
	// workerNodePool2Zone
	workerNodePool2.FailureDomain = util.StringPtr(spec.WorkerNodePool2Zone)
	// workerVmClass
	workerNodePool2.Variables.Overrides, _ = popClusterVariables(workerNodePool2.Variables.Overrides, "vmClass")
	workerNodePool2.Variables.Overrides = append(workerNodePool2.Variables.Overrides, capiv1beta1.ClusterVariable{
		Name:  "vmClass",
		Value: apiextensionsv1.JSON{Raw: []byte(strconv.Quote(spec.WorkerVmClass))},
	})

	var buf bytes.Buffer
	ser := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	if err := ser.Encode(patch, &buf); err != nil {
		return errors.Wrap(err, "generating cluster patch")
	}

	err = w.Apply(buf.Bytes(), mapd.GlobalConfig.HTTPClientTimeout)
	if err != nil {
		return errors.Wrap(err, "applying cluster")
	}
	return nil
}

func NewVSphereWithTanzu(endpoint, user string, password []byte, clientCfg rest.TLSClientConfig) (*VSphereWithTanzu, error) {
	tkg, err := k8s.NewVSphereWithTanzu2(endpoint, user, password, clientCfg)
	if err != nil {
		return nil, err
	}
	return &VSphereWithTanzu{VSphereWithTanzu2: tkg}, nil
}

func popClusterVariables(vars []capiv1beta1.ClusterVariable, name ...string) (result, poped []capiv1beta1.ClusterVariable) {
	nameMap := make(map[string]bool)
	for _, s := range name {
		nameMap[s] = true
	}

	var ret, pop []capiv1beta1.ClusterVariable
	for _, each := range vars {
		if nameMap[each.Name] {
			pop = append(pop, each)
			continue
		}
		ret = append(ret, each)
	}
	return ret, pop
}

func getClusterReplicas(mds []capiv1beta1.MachineDeployment) (int32, int32, int32, int32) {
	var readyReplicas int32 = 0
	var specReplicas int32 = 0
	var replicas int32 = 0
	var updatedReplicas int32 = 0
	for i := range mds {
		readyReplicas += mds[i].Status.ReadyReplicas
		replicas += mds[i].Status.Replicas
		updatedReplicas += mds[i].Status.UpdatedReplicas
		if mds[i].Spec.Replicas != nil {
			specReplicas += *mds[i].Spec.Replicas
		}
	}
	return readyReplicas, specReplicas, replicas, updatedReplicas
}

func getClusterWorkerCount(clusterInfo *ClusterObjects) string {
	readyReplicas, specReplicas, _, _ := getClusterReplicas(clusterInfo.MachineDeployments)
	// return empty string if specReplicas are not yet set
	if specReplicas == 0 {
		return ""
	}
	cpReplicas := fmt.Sprintf("%v/%v", readyReplicas, specReplicas)
	return cpReplicas
}

func getClusterRoles(clusterLabels map[string]string) []string {
	clusterRoles := make([]string, 0)
	for labelKey := range clusterLabels {
		if !strings.HasPrefix(labelKey, TkgLabelClusterRolePrefix) {
			continue
		}

		clusterRole := strings.TrimPrefix(labelKey, TkgLabelClusterRolePrefix)
		clusterRoles = append(clusterRoles, clusterRole)
	}

	return clusterRoles
}

func getClusterControlPlaneCount(clusterInfo *ClusterObjects) string {
	if clusterInfo.KubeadmControlPlane.Spec.Replicas == nil {
		return ""
	}
	cpReplicas := fmt.Sprintf("%v/%v", clusterInfo.KubeadmControlPlane.Status.ReadyReplicas, *clusterInfo.KubeadmControlPlane.Spec.Replicas)
	return cpReplicas
}

func getClusterStatus(clusterInfo *ClusterObjects) TKGClusterPhase {
	if strings.EqualFold(clusterInfo.Cluster.Status.Phase, string(TKGClusterPhaseDeleting)) {
		return TKGClusterPhaseDeleting
	}

	readyReplicas, specReplicas, replicas, updatedReplicas := getClusterReplicas(clusterInfo.MachineDeployments)

	creationCompleteCondition := clusterInfo.Cluster.Status.InfrastructureReady &&
		clusterInfo.Cluster.Status.ControlPlaneReady &&
		clusterInfo.KubeadmControlPlane.Status.ReadyReplicas > 0 &&
		readyReplicas > 0

	if !creationCompleteCondition {
		return TKGClusterPhaseCreating
	}

	runningCondition := clusterInfo.Cluster.Status.ControlPlaneReady &&
		*clusterInfo.KubeadmControlPlane.Spec.Replicas == clusterInfo.KubeadmControlPlane.Status.Replicas &&
		*clusterInfo.KubeadmControlPlane.Spec.Replicas == clusterInfo.KubeadmControlPlane.Status.ReadyReplicas &&
		*clusterInfo.KubeadmControlPlane.Spec.Replicas == clusterInfo.KubeadmControlPlane.Status.UpdatedReplicas &&
		specReplicas == replicas &&
		specReplicas == readyReplicas &&
		specReplicas == updatedReplicas &&
		!isUpgradeInProgress(clusterInfo)

	if runningCondition {
		return TKGClusterPhaseRunning
	}

	return TKGClusterPhaseUpdating
}

func isUpgradeInProgress(clusterInfo *ClusterObjects) bool {
	// check k8s version of all machine objects version with kcp.Spec.Version
	// to determine this is scaling operation or upgrade operation
	for i := range clusterInfo.Machines {
		if clusterInfo.Machines[i].Spec.Version != nil && *clusterInfo.Machines[i].Spec.Version != clusterInfo.KubeadmControlPlane.Spec.Version {
			return true
		}
	}
	return false
}

func parseClusterStringVariable(val apiextensionsv1.JSON) (string, error) {
	if string(val.Raw) == "null" {
		return "", nil
	}
	var str string
	if err := json.Unmarshal(val.Raw, &str); err != nil {
		return "", errors.Wrap(err, "error parsing variable value")
	}
	return str, nil
}

// SelectTemplateForVsphereProviderBasedonOSOptions selects template among all for vsphere provider
func SelectTemplateForVsphereProviderBasedonOSOptions(vms []*vc.VSphereVirtualMachine, osName, osVersion, osArch string) *vc.VSphereVirtualMachine {
	if len(vms) == 0 { // no vms provided skipping selection
		return nil
	}

	if osName != "" {
		vms = filterVMTemplatesByOSOption(vms, vmOptionOsName, osName)
		if osVersion != "" {
			vms = filterVMTemplatesByOSOption(vms, vmOptionOsVersion, osVersion)
			if osArch != "" {
				vms = filterVMTemplatesByOSOption(vms, vmOptionOsArch, osArch)
			}
		}
		if len(vms) < 1 {
			return nil
		}
		return vms[0]
	}

	if len(vms) == 1 { // only one vm found, return that vm
		return vms[0]
	}

	vm := filterVMTemplatesByAllOsOptions(vms, osName, osVersion, osArch)
	if vm != nil {
		return vm
	}

	return vms[0]
}

// FindMatchingVirtualMachineTemplate finds a virtual machine template that matches the ova versions
func FindMatchingVirtualMachineTemplate(
	vms []*vc.VSphereVirtualMachine,
	ovaVersions []string,
) (matchedTemplateVMs []*vc.VSphereVirtualMachine, nonTemplateVMs []string) {

	for _, vm := range vms {
		for _, ovaVersion := range ovaVersions {
			if vm.OVAVersion == ovaVersion {
				if vm.IsTemplate {
					matchedTemplateVMs = append(matchedTemplateVMs, vm)
				} else {
					nonTemplateVMs = append(nonTemplateVMs, vm.Name)
				}
				break
			}
		}
	}

	return matchedTemplateVMs, nonTemplateVMs
}

// ValidateAndGetVirtualMachineTemplateForTKRVersion gets a virtual machine for the given TKR version and validates it
func ValidateAndGetVirtualMachineTemplateForTKRVersion(vcClient *vc.Client, tkrVersion string, ovaVersions []string, templateName, dc string, k8sVMTemplates []*vc.VSphereVirtualMachine) (*vc.VSphereVirtualMachine, error) {
	templateMOID, err := vcClient.FindVirtualMachine(context.Background(), templateName, dc)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid %s (%v)", ConfigVariableVsphereTemplate, templateName)
	}

	// check if VM template name and K8s version match
	for _, vm := range k8sVMTemplates {
		if vm.Moid == templateMOID {
			if !vm.IsTemplate {
				return nil, errors.Errorf("virtual machine %s is not a template, please convert it to a template and retry", templateName)
			}
			for _, ovaVersion := range ovaVersions {
				if vm.OVAVersion == ovaVersion {
					return vm, nil
				}
			}

			return nil, errors.Errorf("incorrect %s (%s) specified for the TanzuKubernetesRelease (%s). TKG CLI will autodetect the correct VM template to use, so %s should be removed unless required to disambiguate among multiple matching templates",
				ConfigVariableVsphereTemplate, templateName, tkrVersion, ConfigVariableVsphereTemplate)
		}
	}
	return nil, errors.Errorf("VM Template %s is not associated with TanzuKubernetesRelease %s. Please upload VM Template for this TanzuKubernetesRelease to continue", templateName, tkrVersion)
}

func filterVMTemplatesByOSOption(vms []*vc.VSphereVirtualMachine, osOption, osOptionValue string) []*vc.VSphereVirtualMachine {
	filteredVms := []*vc.VSphereVirtualMachine{}
	for _, vm := range vms {
		switch osOption {
		case vmOptionOsName:
			if vm.DistroName == osOptionValue {
				filteredVms = append(filteredVms, vm)
			}
		case vmOptionOsVersion:
			if vm.DistroVersion == osOptionValue {
				filteredVms = append(filteredVms, vm)
			}
		case vmOptionOsArch:
			if vm.DistroArch == osOptionValue {
				filteredVms = append(filteredVms, vm)
			}
		}
	}
	return filteredVms
}

func filterVMTemplatesByAllOsOptions(vms []*vc.VSphereVirtualMachine, osName, osVersion, osArch string) *vc.VSphereVirtualMachine {
	for _, vm := range vms {
		if (vm.DistroName == osName || osName == "") &&
			(vm.DistroVersion == osVersion || osVersion == "") &&
			(vm.DistroArch == osArch || osArch == "") {
			return vm
		}
	}
	return nil
}
