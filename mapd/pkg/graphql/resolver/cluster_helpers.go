// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"bytes"
	"encoding/json"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

func convertClusterFromSQLModel(cluster *model1.Cluster) *model.Cluster {
	obj := &model.Cluster{
		ID:                    cluster.ID,
		Name:                  cluster.Name,
		Description:           cluster.Description,
		Type:                  model.ClusterType(cluster.Type.String()),
		Class:                 model.ClusterClass(cluster.Class.String()),
		CaCert:                string(cluster.CACert),
		InsecureSkipTLSVerify: cluster.InsecureTLS,
		Version:               abstract.MustParseVersion(cluster.Version),
		Status:                model.ClusterStatus(cluster.Status.String()),
		State:                 model.ClusterState(cluster.State.String()),
		CreatedAt:             cluster.CreatedAt,
		LastModified:          cluster.UpdatedAt,
	}
	if cluster.Endpoint != "" {
		obj.Endpoint = abstract.MustParseURL(cluster.Endpoint)
	}
	switch cluster.Class {
	case model1.ClusterClassTKGS:
		spec := &model1.ClusterTKGSSpec{}
		if err := json.Unmarshal(cluster.Spec, spec); err != nil {
			panic(err)
		}

		v := &model.TKGSClusterSpec{
			Namespace:             spec.Namespace,
			Cni:                   spec.CNI,
			SvcDomain:             spec.ServiceDomain,
			DefStorage:            spec.DefStorage,
			CtlPlaneClass:         spec.CtlPlaneClass,
			CtlPlaneCount:         spec.CtlPlaneCount,
			CtlPlaneStorage:       spec.CtlPlaneStorage,
			CtlPlaneStorageCapGi:  spec.CtlPlaneStorageCapGi,
			DataPlaneClass:        spec.DataPlaneClass,
			DataPlaneCount:        spec.DataPlaneCount,
			DataPlaneStorage:      spec.DataPlaneStorage,
			DataPlaneStorageCapGi: spec.DataPlaneStorageCapGi,
		}
		if spec.PodCIDR != "" {
			v.PodCidr = abstract.MustParseCIDR(spec.PodCIDR)
		}
		if spec.ServiceCIDR != "" {
			v.SvcCidr = abstract.MustParseCIDR(spec.ServiceCIDR)
		}
		obj.Spec = v
	case model1.ClusterClassTKGM:
		spec := &model1.ClusterTKGMSpec{}
		if err := json.Unmarshal(cluster.Spec, spec); err != nil {
			panic(err)
		}

		v := &model.TKGMClusterSpec{
			Namespace:             spec.Namespace,
			Cni:                   spec.CNI,
			CtlPlaneCount:         spec.CtlPlaneCount,
			CtlPlaneCPUs:          spec.CtlPlaneCPUs,
			CtlPlaneMemMi:         spec.CtlPlaneMemoryMi,
			CtlPlaneStorageCapGi:  spec.CtlPlaneStorageCapGi,
			DataPlaneCount:        spec.DataPlaneCount,
			DataPlaneCPUs:         spec.DataPlaneCPUs,
			DataPlaneMemMi:        spec.DataPlaneMemoryMi,
			DataPlaneStorageCapGi: spec.DataPlaneStorageCapGi,
			Env:                   spec.Env,
		}
		if spec.PodCIDR != "" {
			v.PodCidr = abstract.MustParseCIDR(spec.PodCIDR)
		}
		if spec.ServiceCIDR != "" {
			v.SvcCidr = abstract.MustParseCIDR(spec.ServiceCIDR)
		}
		obj.Spec = v
	case model1.ClusterClassTKG2S:
		spec := &model1.ClusterTKG2SSpec{}
		if err := json.Unmarshal(cluster.Spec, spec); err != nil {
			panic(err)
		}

		v := &model.TKG2SClusterSpec{
			Namespace:                    spec.Namespace,
			OsName:                       spec.OSName,
			Cni:                          spec.CNI,
			StorageClass:                 spec.StorageClass,
			CtlPlaneReplicas:             spec.CtlPlaneReplicas,
			CtlPlaneVMClass:              spec.CtlPlaneVmClass,
			CtlPlaneEtcdVolumeSizeGi:     spec.CtlPlaneEtcdVolumeSizeGi,
			WorkerNodePool0Replicas:      spec.WorkerNodePool0Replicas,
			WorkerNodePool0Zone:          spec.WorkerNodePool0Zone,
			WorkerNodePool1Replicas:      spec.WorkerNodePool1Replicas,
			WorkerNodePool1Zone:          spec.WorkerNodePool1Zone,
			WorkerNodePool2Replicas:      spec.WorkerNodePool2Replicas,
			WorkerNodePool2Zone:          spec.WorkerNodePool2Zone,
			WorkerVMClass:                spec.WorkerVmClass,
			WorkerContainerdVolumeSizeGi: spec.WorkerContainerdVolumeSizeGi,
		}
		if spec.ClusterCIDR != "" {
			v.ClusterCidr = abstract.MustParseCIDR(spec.ClusterCIDR)
		}
		if spec.ServiceCIDR != "" {
			v.ServiceCidr = abstract.MustParseCIDR(spec.ServiceCIDR)
		}
		obj.Spec = v
	case model1.ClusterClassTKG2M:
		spec := &model1.ClusterTKG2MSpec{}
		if err := json.Unmarshal(cluster.Spec, spec); err != nil {
			panic(err)
		}

		v := &model.TKG2MClusterSpec{
			Namespace:               spec.Namespace,
			OsName:                  spec.OSName,
			Cni:                     spec.CNI,
			CtlPlaneReplicas:        spec.CtlPlaneReplicas,
			CtlPlaneNumCpus:         spec.CtlPlaneNumCpus,
			CtlPlaneMemMiB:          spec.CtlPlaneMemMiB,
			CtlPlaneDiskGiB:         spec.CtlPlaneDiskGiB,
			WorkerNodePool0Replicas: spec.WorkerNodePool0Replicas,
			WorkerNodePool0Zone:     spec.WorkerNodePool0Zone,
			WorkerNodePool1Replicas: spec.WorkerNodePool1Replicas,
			WorkerNodePool1Zone:     spec.WorkerNodePool1Zone,
			WorkerNodePool2Replicas: spec.WorkerNodePool2Replicas,
			WorkerNodePool2Zone:     spec.WorkerNodePool2Zone,
			WorkerNumCpus:           spec.WorkerNumCpus,
			WorkerMemMiB:            spec.WorkerMemMiB,
			WorkerDiskGiB:           spec.WorkerDiskGiB,
		}
		if spec.ClusterCIDR != "" {
			v.ClusterCidr = abstract.MustParseCIDR(spec.ClusterCIDR)
		}
		if spec.ServiceCIDR != "" {
			v.ServiceCidr = abstract.MustParseCIDR(spec.ServiceCIDR)
		}
		obj.Spec = v
	}
	if kubeconfig := exportKubeConfig(cluster); kubeconfig != "" {
		obj.KubeConfig = &kubeconfig
	}
	if cluster.LoadBalancerID.Valid {
		obj.LoadBalancer = &model.LoadBalancer{ID: cluster.LoadBalancerID.UUID}
	}
	if cluster.ProviderID.Valid {
		obj.Provider = &model.ClusterProvider{ID: cluster.ProviderID.UUID}
	}
	if cluster.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: cluster.OrganizationID.UUID}
	}
	return obj
}

func finalizeClusterQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
	for _, each := range filters {
		switch each.Key {
		case "name":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`name` LIKE ?")
				args = append(args, "%"+textutil.EscapeSQLString(val)+"%")
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "type":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`type` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "class":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`class` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "status":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`status` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "state":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`state` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "related":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` IN (SELECT `guest_id` FROM `cluster_extension_related_cluster` WHERE `extension_id` = ?)")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "notRelated":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` NOT IN (SELECT `guest_id` FROM `cluster_extension_related_cluster` WHERE `extension_id` = ?)")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "relatedName":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` IN (SELECT `guest_id` FROM `cluster_extension_related_cluster` WHERE `extension_id` IN (SELECT `id` FROM `cluster_extension` WHERE `name` = ?))")
				args = append(args, strings.ToUpper(val))
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "noRelatedName":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` NOT IN (SELECT `guest_id` FROM `cluster_extension_related_cluster` WHERE `extension_id` IN (SELECT `id` FROM `cluster_extension` WHERE `name` = ?))")
				args = append(args, strings.ToUpper(val))
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "extension":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` IN (SELECT `cluster_id` FROM `cluster_extension` WHERE `name` = ?)")
				args = append(args, strings.ToUpper(val))
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "noExtension":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` NOT IN (SELECT `cluster_id` FROM `cluster_extension` WHERE `name` = ?)")
				args = append(args, strings.ToUpper(val))
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "imageRegistry":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				if val != "" {
					exprs = append(exprs, "`image_registry_id` = ?")
					args = append(args, val)
					continue
				}
				exprs = append(exprs, "`image_registry_id` IS NULL")
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "loadBalancer":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				if val != "" {
					exprs = append(exprs, "`load_balancer_id` = ?")
					args = append(args, val)
					continue
				}
				exprs = append(exprs, "`load_balancer_id` IS NULL")
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "organization":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`organization_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		}
	}
	return db
}

func convertClusterImageRegistryFromSQLModel(registry *model1.ClusterImageRegistry) *model.ClusterImageRegistry {
	obj := &model.ClusterImageRegistry{
		ID:      registry.ImageRegistryID.UUID,
		Default: registry.Default,
	}
	return obj
}

func finalizeClusterImageRegistryQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
	for _, each := range filters {
		switch each.Key {
		case "name":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`image_registry_id` IN (SELECT `id` FROM `image_registry` WHERE `name` LIKE ? AND `deleted_at` IS NULL)")
				args = append(args, "%"+textutil.EscapeSQLString(val)+"%")
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "default":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`default` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "pushable":
			db = db.Where("`image_registry_id` IN (SELECT `id` FROM `image_registry` WHERE `auth_user` != '' AND `deleted_at` IS NULL)")
		}
	}
	return db
}

func exportKubeConfig(cluster *model1.Cluster) string {
	if cluster.AuthToken == nil || len(cluster.AuthToken) == 0 {
		return ""
	}

	block, err := cryptoutil.NewAESGCMCipher(cluster.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	config := clientcmdapi.NewConfig()
	config.Clusters[cluster.Name] = &clientcmdapi.Cluster{
		Server:                cluster.Endpoint,
		InsecureSkipTLSVerify: true,
	}
	config.AuthInfos[cluster.Name] = &clientcmdapi.AuthInfo{
		Token: string(block.Decrypt(cluster.AuthToken)),
	}
	config.Contexts[cluster.Name] = &clientcmdapi.Context{
		Cluster:  cluster.Name,
		AuthInfo: cluster.Name,
	}
	config.CurrentContext = cluster.Name

	var buf bytes.Buffer
	if err := clientcmdlatest.Codec.Encode(config, &buf); err != nil {
		panic(err)
	}
	return buf.String()
}
