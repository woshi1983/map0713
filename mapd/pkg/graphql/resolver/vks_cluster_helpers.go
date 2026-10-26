package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/vks"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// prepareVKSCluster validates the selected Supervisor resources before MAP
// commits a pending cluster record and lets conductor create the remote CR.
func prepareVKSCluster(ctx context.Context, cp *model1.ClusterProvider, id uuid.UUID, name string, input *model.CreateVKSClusterSpec, version string, usage *model1.OrganizationQuotaUsage) (*model1.ClusterVKSSpec, error) {
	if input == nil {
		return nil, errInvalidRequest(ctx, "input.spec.vks", "value cannot be empty")
	}
	if input.PodCidr == nil || input.SvcCidr == nil || input.SvcDomain == nil || input.OsName == nil || input.CtlPlaneReplicas == nil || input.WorkerReplicas == nil || input.CtlPlaneEtcdVolumeSizeGi == nil || input.WorkerContainerdVolumeSizeGi == nil {
		return nil, errInvalidRequest(ctx, "input.spec.vks", "required VKS topology fields are missing")
	}
	provider := &model1.ClusterProviderVKSSpec{}
	if err := json.Unmarshal(cp.Spec, provider); err != nil {
		return nil, err
	}
	namespace := provider.DefaultNS
	if input.Namespace != nil && *input.Namespace != "" {
		namespace = *input.Namespace
	}
	if protectedVKSNamespace(namespace) {
		return nil, errInvalidRequest(ctx, "input.spec.vks.namespace", "invalid workload namespace", namespace)
	}
	client, _, err := vks.NewClientFromProvider(cp.ID, provider)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	classes, err := client.ListClusterClasses(ctx, namespace)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	selected := ""
	if input.ClusterClass != nil {
		selected = *input.ClusterClass
	}
	if selected == "" {
		if len(classes) != 1 {
			return nil, errInvalidRequest(ctx, "input.spec.vks.clusterClass", fmt.Sprintf("select one of %d available ClusterClasses", len(classes)))
		}
		selected = classes[0]
	}
	classNS := ""
	className := selected
	if parts := strings.SplitN(selected, "/", 2); len(parts) == 2 {
		classNS, className = parts[0], parts[1]
	}
	if input.ClassNamespace != nil && *input.ClassNamespace != "" {
		classNS = *input.ClassNamespace
	}
	if classNS == "" {
		classNS = namespace
	}
	qualified := classNS + "/" + className
	found := false
	for _, each := range classes {
		if each == qualified {
			found = true
			break
		}
	}
	if !found {
		return nil, errInvalidRequest(ctx, "input.spec.vks.clusterClass", "ClusterClass is unavailable in Supervisor", qualified)
	}
	releaseName, err := client.ResolveRelease(ctx, version)
	if err != nil {
		return nil, errInvalidRequest(ctx, "input.version", err.Error(), version)
	}
	ctl, err := client.GetVMClass(ctx, namespace, input.CtlPlaneVMClass)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	worker, err := client.GetVMClass(ctx, namespace, input.WorkerVMClass)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	ctlCPUs, ctlMemoryMi, err := vks.VMClassResources(ctl)
	if err != nil {
		return nil, errInvalidRequest(ctx, "input.spec.vks.ctlPlaneVmClass", err.Error())
	}
	workerCPUs, workerMemoryMi, err := vks.VMClassResources(worker)
	if err != nil {
		return nil, errInvalidRequest(ctx, "input.spec.vks.workerVmClass", err.Error())
	}
	osVersion, cni := "", "antrea"
	if input.OsVersion != nil {
		osVersion = *input.OsVersion
	}
	osVersion, err = client.ResolveOSVersion(ctx, releaseName, *input.OsName, osVersion)
	if err != nil {
		return nil, errInvalidRequest(ctx, "input.spec.vks.osVersion", err.Error())
	}
	if input.Cni != nil && *input.Cni != "" {
		cni = *input.Cni
	}
	spec := &model1.ClusterVKSSpec{
		Namespace: namespace, ClusterClass: className, ClassNamespace: classNS,
		OSName: *input.OsName, OSVersion: osVersion, CNI: cni,
		ClusterCIDR: input.PodCidr.String(), ServiceCIDR: input.SvcCidr.String(),
		ServiceDomain: *input.SvcDomain, StorageClass: input.StorageClass,
		CtlPlaneVmClass: input.CtlPlaneVMClass, CtlPlaneReplicas: *input.CtlPlaneReplicas,
		CtlPlaneEtcdVolumeSizeGi: *input.CtlPlaneEtcdVolumeSizeGi,
		WorkerVmClass:            input.WorkerVMClass, WorkerReplicas: *input.WorkerReplicas,
		WorkerContainerdVolumeSizeGi: *input.WorkerContainerdVolumeSizeGi,
	}
	if spec.CtlPlaneReplicas < 1 || spec.WorkerReplicas < 1 {
		return nil, errInvalidRequest(ctx, "input.spec.vks", "node counts must be positive")
	}
	for _, ca := range provider.TrustedCAs {
		spec.TrustedCAs = append(spec.TrustedCAs, ca.Data)
	}
	if err := client.ValidateCluster(ctx, id, name, releaseName, *spec); err != nil {
		return nil, errInvalidRequest(ctx, "input.spec.vks", err.Error())
	}
	usage.CPUs = spec.CtlPlaneReplicas*ctlCPUs + spec.WorkerReplicas*workerCPUs
	usage.MemoryMi = spec.CtlPlaneReplicas*ctlMemoryMi + spec.WorkerReplicas*workerMemoryMi
	usage.StorageGi = spec.CtlPlaneReplicas*spec.CtlPlaneEtcdVolumeSizeGi + spec.WorkerReplicas*spec.WorkerContainerdVolumeSizeGi
	return spec, nil
}

func deleteVKSCluster(ctx context.Context, provider *model1.ClusterProvider, cluster *model1.Cluster) error {
	providerSpec := &model1.ClusterProviderVKSSpec{}
	if err := json.Unmarshal(provider.Spec, providerSpec); err != nil {
		return err
	}
	spec := &model1.ClusterVKSSpec{}
	if err := json.Unmarshal(cluster.Spec, spec); err != nil {
		return err
	}
	client, _, err := vks.NewClientFromProvider(provider.ID, providerSpec)
	if err != nil {
		return err
	}
	if cluster.ID == uuid.Nil {
		return fmt.Errorf("cannot delete VKS cluster without MAP ID")
	}
	if err := client.DeleteCluster(ctx, cluster.ID, cluster.Name, *spec); err != nil {
		return err
	}
	_, err = client.GetCluster(ctx, spec.Namespace, cluster.Name)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("VKS cluster deletion is still in progress; retry after the remote object disappears")
}
