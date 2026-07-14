// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	runv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/run.tanzu/v1alpha1"
	vmoperatorv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/vmoperator/v1alpha1"
	tanzuv2 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/tanzu/v2"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func convertClusterProviderFromSQLModel(cp *model1.ClusterProvider, spec interface{}) *model.ClusterProvider {
	obj := &model.ClusterProvider{
		ID:           cp.ID,
		Name:         cp.Name,
		Description:  cp.Description,
		Class:        model.ClusterClass(cp.Class.String()),
		CreatedAt:    cp.CreatedAt,
		LastModified: cp.UpdatedAt,
	}
	switch v := spec.(type) {
	case *model1.ClusterProviderTKGSSpec:
		spec := &model.TKGSClusterProviderSpec{
			ID:                    cp.ID,
			Endpoint:              abstract.MustParseURL(v.Endpoint),
			CaCert:                string(v.CACert),
			InsecureSkipTLSVerify: v.InsecureSkipTLSVerify,
			AuthUser:              v.AuthUser,
			DefaultNamespace:      v.DefaultNS,
			Namespaces:            nil,
		}
		obj.Spec = spec
	case *model1.ClusterProviderTKGMSpec:
		block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}

		cli, err := k8s.NewKubernetesFromKubeConfig(block.Decrypt(v.KubeConfig))
		if err != nil {
			panic(err)
		}
		spec := &model.TKGMClusterProviderSpec{
			ID:               cp.ID,
			Endpoint:         abstract.MustParseURL(cli.Config.Host),
			DefaultNamespace: v.DefaultNS,
			DefaultEnv:       v.DefaultEnv,
			Namespaces:       nil,
		}
		for _, each := range v.TrustedCAs {
			spec.TrustedCAs = append(spec.TrustedCAs, &model.ClusterProviderSpecTrustedCa{
				Name: each.Name,
				Cert: string(each.Data),
			})
		}
		obj.Spec = spec
	case *model1.ClusterProviderTKG2SSpec:
		spec := &model.TKG2SClusterProviderSpec{
			ID:                    cp.ID,
			Endpoint:              abstract.MustParseURL(v.Endpoint),
			CaCert:                string(v.CACert),
			InsecureSkipTLSVerify: v.InsecureSkipTLSVerify,
			AuthUser:              v.AuthUser,
			DefaultNamespace:      v.DefaultNS,
			Namespaces:            nil,
		}
		for _, each := range v.TrustedCAs {
			spec.TrustedCAs = append(spec.TrustedCAs, &model.ClusterProviderSpecTrustedCa{
				Name: each.Name,
				Cert: string(each.Data),
			})
		}
		obj.Spec = spec
	case *model1.ClusterProviderTKG2MSpec:
		block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}

		cli, err := k8s.NewKubernetesFromKubeConfig(block.Decrypt(v.KubeConfig))
		if err != nil {
			panic(err)
		}
		spec := &model.TKG2MClusterProviderSpec{
			ID:                    cp.ID,
			Endpoint:              abstract.MustParseURL(cli.Config.Host),
			DefaultNamespace:      v.DefaultNS,
			Namespaces:            nil,
			VsphereNetwork:        v.VsphereNetwork,
			VsphereResourcePool:   v.VsphereResourcePool,
			VsphereFolder:         v.VsphereFolder,
			VsphereDatastore:      v.VsphereDatastore,
			AviServiceEngineGroup: v.AviServiceEngineGroup,
			AviDataNetwork:        v.AviDataNetwork,
			AviDataNetworkCidr:    v.AviDataNetworkCidr,
		}
		for _, each := range v.TrustedCAs {
			spec.TrustedCAs = append(spec.TrustedCAs, &model.ClusterProviderSpecTrustedCa{
				Name: each.Name,
				Cert: string(each.Data),
			})
		}
		obj.Spec = spec
	}
	if cp.LoadBalancerID.Valid {
		obj.LoadBalancer = &model.LoadBalancer{ID: cp.LoadBalancerID.UUID}
	}
	if cp.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: cp.OrganizationID.UUID}
	}
	return obj
}

func finalizeClusterProviderQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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
		case "imageRegistry":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`image_registry_id` = ?")
				args = append(args, val)
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
				exprs = append(exprs, "`load_balancer_id` = ?")
				args = append(args, val)
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

func convertClusterProviderImageRegistryFromSQLModel(registry *model1.ClusterProviderImageRegistry) *model.ClusterProviderImageRegistry {
	obj := &model.ClusterProviderImageRegistry{
		ID:      registry.ImageRegistryID.UUID,
		Default: registry.Default,
	}
	return obj
}

func finalizeClusterProviderImageRegistryQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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

func patchTkgServiceConfigurationAdditionalTrustedCAs(client *k8s.TanzuKubernetesGridService, name string, certs []runv1alpha1.TLSCertificate) (*runv1alpha1.TkgServiceConfiguration, error) {
	b, err := json.Marshal(certs)
	if err != nil {
		panic(err)
	}

	patchString := `{
	"spec": {
		"trust": {
			"additionalTrustedCAs": %s
		}
	}
}`

	return client.PatchTkgServiceConfiguration(name, types.MergePatchType, []byte(fmt.Sprintf(patchString, b)))
}

func getVirtualMachineClasses(client *k8s.TanzuKubernetesGridService) ([]vmoperatorv1alpha1.VirtualMachineClass, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetVirtualMachineClassesContext(ctx)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func getTKGSClusterProviderNamespaces(ctx context.Context, id uuid.UUID, endpoint string, insecure bool, caCert []byte, authUser string, authPassword []byte, filters []*model.QueryFilter) ([]*model.TKGSClusterProviderSpecNamespace, error) {
	client, err := k8s.NewTanzuKubernetesGridService(endpoint, authUser, authPassword, rest.TLSClientConfig{
		Insecure: insecure,
		CAData:   caCert,
	})
	if err != nil {
		panic(err)
	}

	var (
		namespaces []*model.TKGSClusterProviderSpecNamespace
		selectors  = []k8s.Selector{k8s.LabelExists("vSphereClusterID")}
	)
	for _, each := range filters {
		switch each.Key {
		case "name":
			if len(each.Value) == 1 {
				selectors = append(selectors, k8s.FieldEqual("metadata.name", each.Value[0]))
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetNamespacesContext(ctx, selectors...)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	for _, each := range list.Items {
		namespace := &model.TKGSClusterProviderSpecNamespace{
			ClusterProvider: id,
			Name:            each.Name,
			Labels:          each.Labels,
		}
		if !each.CreationTimestamp.IsZero() {
			namespace.CreatedAt = each.CreationTimestamp.Time
		} else {
			namespace.CreatedAt = time.Now()
		}
		if namespace.CreatedAt.IsZero() {
			namespace.CreatedAt = time.Unix(0, 0)
		}
		if namespace.LastModified.IsZero() {
			namespace.LastModified = namespace.CreatedAt
		}
		namespaces = append(namespaces, namespace)
	}
	return namespaces, nil
}

func getTKG2SClusterProviderNamespaces(ctx context.Context, id uuid.UUID, endpoint string, insecure bool, caCert []byte, authUser string, authPassword []byte, filters []*model.QueryFilter) ([]*model.TKG2SClusterProviderSpecNamespace, error) {
	client, err := tanzuv2.NewVSphereWithTanzu(endpoint, authUser, authPassword, rest.TLSClientConfig{
		Insecure: insecure,
		CAData:   caCert,
	})
	if err != nil {
		panic(err)
	}

	var (
		namespaces []*model.TKG2SClusterProviderSpecNamespace
		selectors  []k8s.Selector
	)
	for _, each := range filters {
		switch each.Key {
		case "name":
			if len(each.Value) == 1 {
				selectors = append(selectors, k8s.FieldEqual("metadata.name", each.Value[0]))
			}
		}
	}

	dctx, dcancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer dcancel()

	list, err := client.GetUserVisibleNamespacesContext(dctx, selectors...)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	for _, each := range list.Items {
		namespace := &model.TKG2SClusterProviderSpecNamespace{
			ClusterProvider: id,
			Name:            each.Name,
			Labels:          each.Labels,
		}
		if !each.CreationTimestamp.IsZero() {
			namespace.CreatedAt = each.CreationTimestamp.Time
		} else {
			namespace.CreatedAt = time.Now()
		}
		if namespace.CreatedAt.IsZero() {
			namespace.CreatedAt = time.Unix(0, 0)
		}
		if namespace.LastModified.IsZero() {
			namespace.LastModified = namespace.CreatedAt
		}
		namespaces = append(namespaces, namespace)
	}
	return namespaces, nil
}

func getTKG2MClusterProviderNamespaces(ctx context.Context, id uuid.UUID, kubeconfig []byte, filters []*model.QueryFilter) ([]*model.TKG2MClusterProviderSpecNamespace, error) {
	client, err := tanzuv2.NewTanzuKubernetesGrid(kubeconfig)
	if err != nil {
		panic(err)
	}

	var (
		namespaces []*model.TKG2MClusterProviderSpecNamespace
		selectors  []k8s.Selector
	)
	for _, each := range filters {
		switch each.Key {
		case "name":
			if len(each.Value) == 1 {
				selectors = append(selectors, k8s.FieldEqual("metadata.name", each.Value[0]))
			}
		}
	}

	list, err := client.GetNamespaces(selectors...)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	for _, each := range list.Items {
		if tkg2mProviderProtectedNamespaces[each.Name] {
			continue
		}

		namespace := &model.TKG2MClusterProviderSpecNamespace{
			ClusterProvider: id,
			Name:            each.Name,
			Labels:          each.Labels,
		}
		if !each.CreationTimestamp.IsZero() {
			namespace.CreatedAt = each.CreationTimestamp.Time
		} else {
			namespace.CreatedAt = time.Now()
		}
		if namespace.CreatedAt.IsZero() {
			namespace.CreatedAt = time.Unix(0, 0)
		}
		if namespace.LastModified.IsZero() {
			namespace.LastModified = namespace.CreatedAt
		}
		namespaces = append(namespaces, namespace)
	}
	return namespaces, nil
}

func validateTKG2ManagementCluster(ctx context.Context, kubeConfig []byte) error {
	client, err := tanzuv2.NewTanzuKubernetesGrid(kubeConfig)
	if err != nil {
		return errInvalidRequest(ctx, "input.spec.kubeConfig", "not a valid kubeconfig", string(kubeConfig))
	}

	ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	values, err := client.GetEnvironment()
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return errInvalidRequest(ctx, "input.spec.kubeConfig", "cluster is not a management cluster", string(kubeConfig))
		}
		return errGatewayError(ctx, err)
	}

	clusterRole, found, err := unstructured.NestedString(values, tanzuv2.ConfigVariableClusterRole)
	if !found || err != nil {
		clusterRole = ""
	}
	if strings.ToLower(clusterRole) != "management" {
		return errInvalidRequest(ctx, "input.spec.kubeConfig", "cluster is not a management cluster", string(kubeConfig))
	}
	return nil
}

var (
	tkg2sProviderProtectedNamespaces = map[string]bool{
		"default":           true,
		"kube-node-lease":   true,
		"kube-public":       true,
		"kube-system":       true,
		"svc-tmc-c8":        true,
		"tanzu-cli-system":  true,
		"vmware-system-ako": true,
		"vmware-system-appplatform-operator-system": true,
		"vmware-system-capw":                        true,
		"vmware-system-cert-manager":                true,
		"vmware-system-csi":                         true,
		"vmware-system-kubeimage":                   true,
		"vmware-system-license-operator":            true,
		"vmware-system-logging":                     true,
		"vmware-system-netop":                       true,
		"vmware-system-nsop":                        true,
		"vmware-system-pinniped":                    true,
		"vmware-system-pkgs":                        true,
		"vmware-system-registry":                    true,
		"vmware-system-supervisor-services":         true,
		"vmware-system-tkg":                         true,
		"vmware-system-ucs":                         true,
		"vmware-system-vmop":                        true,
	}
	tkg2mProviderProtectedNamespaces = map[string]bool{
		"avi-system":                        true,
		"capi-kubeadm-bootstrap-system":     true,
		"capi-kubeadm-control-plane-system": true,
		"capi-system":                       true,
		"capv-system":                       true,
		"cert-manager":                      true,
		"default":                           true,
		"kube-node-lease":                   true,
		"kube-public":                       true,
		"kube-system":                       true,
		"secretgen-controller":              true,
		"tanzu-auth":                        true,
		"tkg-system":                        true,
		"tkg-system-networking":             true,
		"tkg-system-public":                 true,
		"tkr-system":                        true,
	}
)
