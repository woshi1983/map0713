// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gorm.io/gorm"
)

type Cluster struct {
	ID             uuid.UUID    `gorm:"type:varchar(255);primaryKey"` // write-protected
	Name           string       `gorm:"type:varchar(128);index"`
	Description    string       `gorm:"type:mediumtext"`
	Type           ClusterType  `gorm:"type:varchar(128);index"`
	Class          ClusterClass `gorm:"type:varchar(128);index"`
	Endpoint       string       `gorm:"type:varchar(255)"`
	CACert         []byte       `gorm:"type:mediumblob"`
	InsecureTLS    bool
	AuthToken      []byte `gorm:"type:mediumblob"`
	Version        string `gorm:"type:varchar(255);index"`
	ProvisionPhase ClusterProvision
	Provisioned    ClusterProvision
	Status         ClusterStatus  `gorm:"type:varchar(128);index"`
	State          ClusterState   `gorm:"type:varchar(128);index"`
	Spec           []byte         `gorm:"type:mediumblob"`
	LoadBalancerID NullUUID       `gorm:"type:varchar(255)"`
	ProviderID     NullUUID       `gorm:"type:varchar(255)"`
	OrganizationID NullUUID       `gorm:"type:varchar(255)"`
	CreatedAt      time.Time      `gorm:"type:datetime(6)"`
	UpdatedAt      time.Time      `gorm:"type:datetime(6)"`
	DeletedAt      gorm.DeletedAt `gorm:"type:datetime(6);index"`

	LoadBalancer *LoadBalancer      `gorm:"foreignKey:LoadBalancerID;references:ID;constraint:onDelete:SET NULL,onUpdate:CASCADE"`
	Provider     *ClusterProvider   `gorm:"foreignKey:ProviderID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Organization *Organization      `gorm:"foreignKey:OrganizationID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Extensions   []ClusterExtension `gorm:"-"`
}

func (typedef Cluster) TableName() string {
	return "cluster"
}

func ModelCluster(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &Cluster{}
		migrator := db.Set("gorm:table_options", " CHARSET=utf8mb4").Migrator()
		if err := migrator.AutoMigrate(typedef); err != nil {
			return err
		}

		if err := next(db); err != nil {
			return err
		}

		// additional migration task here.
		return nil
	}
}

type ClusterImageRegistry struct {
	ID              uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	Default         bool
	ClusterID       NullUUID `gorm:"type:varchar(255)"`
	ImageRegistryID NullUUID `gorm:"type:varchar(255)"`

	Cluster       *Cluster       `gorm:"foreignKey:ClusterID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	ImageRegistry *ImageRegistry `gorm:"foreignKey:ImageRegistryID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ClusterImageRegistry) TableName() string {
	return "cluster_image_registry"
}

func ModelClusterImageRegistry(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterImageRegistry{}
		migrator := db.Set("gorm:table_options", " CHARSET=utf8mb4").Migrator()
		if err := migrator.AutoMigrate(typedef); err != nil {
			return err
		}

		if err := next(db); err != nil {
			return err
		}

		// additional migration task here.
		return nil
	}
}

type ClusterQuotaUsage struct {
	ID        uuid.UUID `gorm:"type:varchar(255);primaryKey"` // write-protected
	ClusterID NullUUID  `gorm:"type:varchar(255)"`
	UsageID   NullUUID  `gorm:"type:varchar(255)"`

	Cluster *Cluster                `gorm:"foreignKey:ClusterID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
	Usage   *OrganizationQuotaUsage `gorm:"foreignKey:UsageID;references:ID;constraint:onDelete:CASCADE,onUpdate:CASCADE"`
}

func (typedef ClusterQuotaUsage) TableName() string {
	return "cluster_quota_usage"
}

func ModelClusterQuotaUsage(next mapd.RegisterHandler) mapd.RegisterHandler {
	return func(db *gorm.DB) error {
		typedef := &ClusterQuotaUsage{}
		migrator := db.Set("gorm:table_options", " CHARSET=utf8mb4").Migrator()
		if err := migrator.AutoMigrate(typedef); err != nil {
			return err
		}

		if err := next(db); err != nil {
			return err
		}

		// additional migration task here.
		return nil
	}
}

type ClusterTKGSSpec struct {
	Namespace             string `json:"namespace,omitempty"`
	CNI                   string `json:"cni,omitempty"`
	PodCIDR               string `json:"podCIDR,omitempty"`
	ServiceCIDR           string `json:"serviceCIDR,omitempty"`
	ServiceDomain         string `json:"serviceDomain,omitempty"`
	DefStorage            string `json:"defStorage,omitempty"`
	CtlPlaneClass         string `json:"ctlPlaneClass,omitempty"`
	CtlPlaneCount         int64  `json:"ctlPlaneCount,omitempty"`
	CtlPlaneStorage       string `json:"ctlPlaneStorage,omitempty"`
	CtlPlaneStorageCapGi  int64  `json:"ctlPlaneStorageCapGi,omitempty"`
	DataPlaneClass        string `json:"dataPlaneClass,omitempty"`
	DataPlaneCount        int64  `json:"dataPlaneCount,omitempty"`
	DataPlaneStorage      string `json:"dataPlaneStorage,omitempty"`
	DataPlaneStorageCapGi int64  `json:"dataPlaneStorageCapGi,omitempty"`
}

type ClusterTKGMSpec struct {
	Namespace             string            `json:"namespace,omitempty"`
	CNI                   string            `json:"cni,omitempty"`
	PodCIDR               string            `json:"podCIDR,omitempty"`
	ServiceCIDR           string            `json:"serviceCIDR,omitempty"`
	CtlPlaneCount         int64             `json:"ctlPlaneCount,omitempty"`
	CtlPlaneCPUs          int64             `json:"ctlPlaneCPUs,omitempty"`
	CtlPlaneMemoryMi      int64             `json:"ctlPlaneMemoryMi,omitempty"`
	CtlPlaneStorageCapGi  int64             `json:"ctlPlaneStorageCapGi,omitempty"`
	DataPlaneCount        int64             `json:"dataPlaneCount,omitempty"`
	DataPlaneCPUs         int64             `json:"dataPlaneCPUs,omitempty"`
	DataPlaneMemoryMi     int64             `json:"dataPlaneMemoryMi,omitempty"`
	DataPlaneStorageCapGi int64             `json:"dataPlaneStorageCapGi,omitempty"`
	SSHPrivateKey         []byte            `json:"sshPrivateKey,omitempty"`
	Env                   map[string]string `json:"env,omitempty"`
	TrustedCAs            [][]byte          `json:"trustedCAs,omitempty"`
}

type ClusterTKG2SSpec struct {
	Namespace                    string   `json:"namespace,omitempty"`
	OSName                       string   `json:"osName,omitempty"`
	CNI                          string   `json:"cni,omitempty"`
	ClusterCIDR                  string   `json:"clusterCIDR,omitempty"`
	ServiceCIDR                  string   `json:"serviceCIDR,omitempty"`
	StorageClass                 string   `json:"storageClass,omitempty"`
	CtlPlaneReplicas             int64    `json:"ctlPlaneReplicas,omitempty"`
	CtlPlaneVmClass              string   `json:"ctlPlaneVmClass,omitempty"`
	CtlPlaneEtcdVolumeSizeGi     int64    `json:"ctlPlaneEtcdVolumeSizeGi,omitempty"`
	WorkerNodePool0Replicas      int64    `json:"workerNodePool0Replicas,omitempty"`
	WorkerNodePool0Zone          string   `json:"workerNodePool0Zone,omitempty"`
	WorkerNodePool1Replicas      int64    `json:"workerNodePool1Replicas,omitempty"`
	WorkerNodePool1Zone          string   `json:"workerNodePool1Zone,omitempty"`
	WorkerNodePool2Replicas      int64    `json:"workerNodePool2Replicas,omitempty"`
	WorkerNodePool2Zone          string   `json:"workerNodePool2Zone,omitempty"`
	WorkerVmClass                string   `json:"workerVmClass,omitempty"`
	WorkerContainerdVolumeSizeGi int64    `json:"workerContainerdVolumeSizeGi,omitempty"`
	TrustedCAs                   [][]byte `json:"trustedCAs,omitempty"`
	NTP                          string   `json:"ntp,omitempty"`
}

type ClusterTKG2MSpec struct {
	Namespace               string   `json:"namespace,omitempty"`
	OSName                  string   `json:"osName,omitempty"`
	CNI                     string   `json:"cni,omitempty"`
	ClusterCIDR             string   `json:"clusterCIDR,omitempty"`
	ServiceCIDR             string   `json:"serviceCIDR,omitempty"`
	CtlPlaneReplicas        int64    `json:"ctlPlaneReplicas,omitempty"`
	CtlPlaneNumCpus         int64    `json:"ctlPlaneNumCpus,omitempty"`
	CtlPlaneMemMiB          int64    `json:"ctlPlaneMemMiB,omitempty"`
	CtlPlaneDiskGiB         int64    `json:"ctlPlaneDiskGiB,omitempty"`
	WorkerNodePool0Replicas int64    `json:"workerNodePool0Replicas,omitempty"`
	WorkerNodePool0Zone     string   `json:"workerNodePool0Zone,omitempty"`
	WorkerNodePool1Replicas int64    `json:"workerNodePool1Replicas,omitempty"`
	WorkerNodePool1Zone     string   `json:"workerNodePool1Zone,omitempty"`
	WorkerNodePool2Replicas int64    `json:"workerNodePool2Replicas,omitempty"`
	WorkerNodePool2Zone     string   `json:"workerNodePool2Zone,omitempty"`
	WorkerNumCpus           int64    `json:"workerNumCpus,omitempty"`
	WorkerMemMiB            int64    `json:"workerMemMiB,omitempty"`
	WorkerDiskGiB           int64    `json:"workerDiskGiB,omitempty"`
	SSHPrivateKey           []byte   `json:"sshPrivateKey,omitempty"`
	TrustedCAs              [][]byte `json:"trustedCAs,omitempty"`
	NTP                     string   `json:"ntp,omitempty"`
	VsphereNetwork          string   `json:"vsphereNetwork,omitempty"`
	VsphereResourcePool     string   `json:"vsphereResourcePool,omitempty"`
	VsphereFolder           string   `json:"vsphereFolder,omitempty"`
	VsphereDatastore        string   `json:"vsphereDatastore,omitempty"`
	AviServiceEngineGroup   string   `json:"aviServiceEngineGroup,omitempty"`
	AviDataNetwork          string   `json:"aviDataNetwork,omitempty"`
	AviDataNetworkCidr      string   `json:"aviDataNetworkCidr,omitempty"`
}

type ClusterExternalSpec struct {
	RootCAs []string `json:"rootCAs,omitempty"`
}

type ClusterType string

func (s ClusterType) String() string {
	if s == ClusterTypeUnknown {
		return "UNKNOWN"
	}
	return string(s)
}

const (
	ClusterTypeUnknown     ClusterType = ""
	ClusterTypeAttached    ClusterType = "ATTACHED"
	ClusterTypeProvisioned ClusterType = "PROVISIONED"
)

type ClusterClass string

func (s ClusterClass) String() string {
	if s == ClusterClassUnknown {
		return "UNKNOWN"
	}
	return string(s)
}

const (
	ClusterClassUnknown ClusterClass = ""
	ClusterClassTKGS    ClusterClass = "TKGS"
	ClusterClassTKGM    ClusterClass = "TKGM"
	ClusterClassTKG2S   ClusterClass = "TKG2S"
	ClusterClassTKG2M   ClusterClass = "TKG2M"
)

type ClusterProvision int64

const (
	ClusterProvisionNone ClusterProvision = iota
	ClusterProvisionPrepare
	ClusterProvisionKubernetes
	ClusterProvisionNamespace
	ClusterProvisionServiceAccount
	ClusterProvisionClusterRoleBinding
	ClusterProvisionValidation
	ClusterProvisionPostScripts
	ClusterProvisionFinish
)

type ClusterState string

func (s ClusterState) String() string {
	if s == ClusterStateUnknown {
		return "UNKNOWN"
	}
	return string(s)
}

const (
	ClusterStateUnknown  ClusterState = ""
	ClusterStateCreating ClusterState = "CREATING" // creating
	ClusterStateRunning  ClusterState = "RUNNING"  // running
	ClusterStateUpdating ClusterState = "UPDATING" // updating
	ClusterStateFailed   ClusterState = "FAILED"   // failed
	ClusterStateDeleting ClusterState = "DELETING" // deleting
)

type ClusterStatus string

func (s ClusterStatus) String() string {
	if s == ClusterStatusPending {
		return "PENDING"
	}
	return string(s)
}

const (
	ClusterStatusPending      ClusterStatus = ""
	ClusterStatusInitializing ClusterStatus = "INITIALIZING"
	ClusterStatusHealthy      ClusterStatus = "HEALTHY"
	ClusterStatusUnhealthy    ClusterStatus = "UNHEALTHY"
	ClusterStatusMaintenance  ClusterStatus = "MAINTENANCE"
)
