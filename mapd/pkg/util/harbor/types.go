// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package harbor

import "helm.sh/helm/v3/pkg/chart"

type SystemInfo struct {
	AuthMode                    string `json:"auth_mode"`
	ExternalURL                 string `json:"external_url"`
	HarborVersion               string `json:"harbor_version"`
	HasCARoot                   bool   `json:"has_ca_root"`
	NotificationEnable          bool   `json:"notification_enable"`
	ProjectCreationRestriction  string `json:"project_creation_restriction"`
	ReadOnly                    bool   `json:"read_only"`
	RegistryStorageProviderName string `json:"registry_storage_provider_name"`
	RegistryURL                 string `json:"registry_url"`
	SelfRegistration            string `json:"self_registration"`
	WithChartmuseum             string `json:"with_chartmuseum"`
	WithNotary                  string `json:"with_notary"`
}

type User struct {
	UserID          int64  `json:"user_id"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordVersion string `json:"password_version"`
	Realname        string `json:"realname"`
	Comment         string `json:"comment"`
	Deleted         bool   `json:"deleted"`
	RoleName        string `json:"role_name"`
	RoleID          int64  `json:"role_id"`
	SysAdminFlag    bool   `json:"sysadmin_flag"`
	AdminRoleInAuth bool   `json:"admin_role_in_auth"`
	ResetUUID       string `json:"reset_uuid"`
	CreationTime    string `json:"creation_time"`
	UpdateTime      string `json:"update_time"`
}

type CreateProject struct {
	ProjectName string `json:"project_name"`
	// The storage quota, -1 for unlimit.
	StorageLimit int64 `json:"storage_limit,omitempty"`
	Public       bool  `json:"public"`
}

type Project struct {
	// The metadata of the project.
	Metadata *ProjectMetadata `json:"metadata,omitempty"`

	// The total number of charts under this project.
	ChartCount int64 `json:"chart_count,omitempty"`

	// The creation time of the project.
	// Format: date-time
	CreationTime string `json:"creation_time,omitempty"`

	// The role ID with highest permission of the current user who triggered the API (for UI).  This attribute is deprecated and will be removed in future versions.
	CurrentUserRoleID int64 `json:"current_user_role_id,omitempty"`

	// The list of role ID of the current user who triggered the API (for UI)
	CurrentUserRoleIds []int64 `json:"current_user_role_ids"`

	// A deletion mark of the project.
	Deleted bool `json:"deleted,omitempty"`

	// The name of the project.
	Name string `json:"name,omitempty"`

	// The owner ID of the project always means the creator of the project.
	OwnerID int64 `json:"owner_id,omitempty"`

	// The owner name of the project.
	OwnerName string `json:"owner_name,omitempty"`

	// Project ID
	ProjectID int64 `json:"project_id,omitempty"`

	// The ID of referenced registry when the project is a proxy cache project.
	RegistryID int64 `json:"registry_id,omitempty"`

	// The number of the repositories under this project.
	RepoCount int64 `json:"repo_count,omitempty"`

	// Correspond to the UI about whether the project's publicity is  updatable (for UI)
	Togglable bool `json:"togglable,omitempty"`

	// The update time of the project.
	// Format: date-time
	UpdateTime string `json:"update_time,omitempty"`
}

type ProjectMetadata struct {
	// Whether scan images automatically when pushing. The valid values are "true", "false".
	AutoScan string `json:"auto_scan,omitempty"`

	// Whether content trust is enabled or not. If it is enabled, user can't pull unsigned images from this project. The valid values are "true", "false".
	EnableContentTrust string `json:"enable_content_trust,omitempty"`

	// Whether prevent the vulnerable images from running. The valid values are "true", "false".
	PreventVul string `json:"prevent_vul,omitempty"`

	// The public status of the project. The valid values are "true", "false".
	Public string `json:"public,omitempty"`

	// The ID of the tag retention policy for the project
	RetentionID string `json:"retention_id,omitempty"`

	// Whether this project reuse the system level CVE allowlist as the allowlist of its own.  The valid values are "true", "false". If it is set to "true" the actual allowlist associate with this project, if any, will be ignored.
	ReuseSysCveAllowlist string `json:"reuse_sys_cve_allowlist,omitempty"`

	// If the vulnerability is high than severity defined here, the images can't be pulled. The valid values are "none", "low", "medium", "high", "critical".
	Severity string `json:"severity,omitempty"`
}

type Repository struct {
	ArtifactCount int64 `json:"artifact_count,omitempty"`

	// The creation time of the project.
	// Format: date-time
	CreationTime string `json:"creation_time,omitempty"`

	ID int64 `json:"id,omitempty"`

	// The name of the project.
	Name string `json:"name,omitempty"`

	// Project ID
	ProjectID int64 `json:"project_id,omitempty"`

	// The update time of the project.
	// Format: date-time
	UpdateTime string `json:"update_time,omitempty"`
}

type Artifact struct {
	ID           int64          `json:"id,omitempty"`
	ProjectID    int64          `json:"project_id,omitempty"`
	RepositoryID int64          `json:"repository_id,omitempty"`
	Size         int64          `json:"size,omitempty"`
	Tags         []*ArtifactTag `json:"tags,omitempty"`
	Type         string         `json:"type,omitempty"`
}

type ArtifactTag struct {
	ArtifactID   int64  `json:"artifact_id,omitempty"`
	ID           int64  `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	RepositoryID int64  `json:"repository_id,omitempty"`
}

type ChartRepo struct {
	Name          string `json:"name,omitempty"`
	LatestVersion string `json:"latest_version,omitempty"`
	TotalVersions int64  `json:"total_versions,omitempty"`
	Icon          string `json:"icon,omitempty"`
	Home          string `json:"home,omitempty"`
	Deprecated    bool   `json:"deprecated,omitempty"`
	Created       string `json:"created,omitempty"`
	Updated       string `json:"updated,omitempty"`
}

type Chart struct {
	// Dependencies are a list of dependencies for a chart.
	Dependencies []*chart.Dependency `json:"dependencies,omitempty"`
	// Metadata is the contents of the Chartfile.
	Metadata ChartMeta `json:"metadata"`
	// Values are default config for this chart.
	Values map[string]interface{} `json:"values"`
	// Schema is an optional JSON schema for imposing structure on Values
	Schema []byte `json:"schema"`
	// Files are miscellaneous files in a chart archive,
	// e.g. README, LICENSE, etc.
	Files map[string]string `json:"files"`
}

// ChartMeta is metadata for a Chart file. This models the structure of a Chart.yaml file.
type ChartMeta struct {
	Annotations  map[string]string   `json:"annotations,omitempty"`
	APIVersion   string              `json:"apiVersion,omitempty"`
	AppVersion   string              `json:"appVersion,omitempty"`
	Dependencies []*chart.Dependency `json:"dependencies,omitempty"`
	Description  string              `json:"description,omitempty"`
	Digest       string              `json:"digest"`
	Home         string              `json:"home,omitempty"`
	Icon         string              `json:"icon,omitempty"`
	Keywords     []string            `json:"keywords,omitempty"`
	Maintainers  []*chart.Maintainer `json:"maintainers,omitempty"`
	Name         string              `json:"name,omitempty"`
	Sources      []string            `json:"sources,omitempty"`
	URLs         []string            `json:"urls"`
	Version      string              `json:"version,omitempty"`
}
