// Copyright (c) 2026 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"encoding/json"
	"fmt"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

// requireRelatedExtensionReady gates an installer before it creates a Kubernetes
// job. A pending dependency is retryable; a failed or disabled one cannot
// become ready without a new user operation.
func requireRelatedExtensionReady(clusterName, dependencyName string, dependency *model.ClusterExtension) error {
	if dependency.Status == model.ClusterExtensionStatusFailed || dependency.State == model.ClusterExtensionStateDisabling {
		return &ErrDependency{Reason: fmt.Sprintf("%s on cluster %s failed or is being disabled", dependencyName, clusterName)}
	}
	if dependency.State != model.ClusterExtensionStateEnabling && dependency.State != model.ClusterExtensionStateEnabled {
		return &ErrDependency{Reason: fmt.Sprintf("%s on cluster %s is not enabled", dependencyName, clusterName)}
	}
	if dependency.State != model.ClusterExtensionStateEnabled || dependency.Status != model.ClusterExtensionStatusProvisioned {
		return &ErrDependency{Reason: fmt.Sprintf("waiting for %s on cluster %s to finish installing", dependencyName, clusterName), Pending: true}
	}
	var endpoints []string
	if err := json.Unmarshal(dependency.Endpoints, &endpoints); err != nil || len(endpoints) == 0 {
		return &ErrDependency{Reason: fmt.Sprintf("%s on cluster %s is provisioned but has no endpoint", dependencyName, clusterName)}
	}
	return nil
}
