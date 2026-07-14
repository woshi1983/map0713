// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package avi

type Login struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Pool struct {
	UUID                            string                  `json:"uuid,omitempty"`
	Name                            string                  `json:"name,omitempty"`
	URL                             string                  `json:"url,omitempty"`
	DefaultServerPort               int64                   `json:"default_server_port,omitempty"`
	LBAlgorithm                     string                  `json:"lb_algorithm,omitempty"`
	LBAlgorithmConsistentHashHeader string                  `json:"lb_algorithm_consistent_hash_hdr,omitempty"`
	LBAlgorithmCoreNonaffinity      int64                   `json:"lb_algorithm_core_nonaffinity,omitempty"`
	LBAlgorithmHash                 string                  `json:"lb_algorithm_hash,omitempty"`
	MinHealthMonitorsUp             int64                   `json:"min_health_monitors_up,omitempty"`
	MinServersUp                    int64                   `json:"min_servers_up,omitempty"`
	Servers                         []*Server               `json:"servers,omitempty"`
	ConnPoolProperties              *PoolConnPoolProperties `json:"conn_pool_properties,omitempty"`
	CloudRef                        string                  `json:"cloud_ref,omitempty"`
}

type Server struct {
	Enabled bool    `json:"enabled,omitempty"`
	IP      *IPAddr `json:"ip,omitempty"`
	Port    int64   `json:"port,omitempty"`
	Ratio   int64   `json:"ratio,omitempty"`
}

type PoolConnPoolProperties struct {
	UpstreamConnpoolConnIdleTmo    int64 `json:"upstream_connpool_conn_idle_tmo,omitempty"`
	UpstreamConnpoolConnLifeTmo    int64 `json:"upstream_connpool_conn_life_tmo,omitempty"`
	UpstreamConnpoolConnMaxReuse   int64 `json:"upstream_connpool_conn_max_reuse,omitempty"`
	UpstreamConnpoolServerMaxCache int64 `json:"upstream_connpool_server_max_cache,omitempty"`
}

type PoolGroup struct {
	UUID       string             `json:"uuid,omitempty"`
	Name       string             `json:"name,omitempty"`
	URL        string             `json:"url,omitempty"`
	Members    []*PoolGroupMember `json:"members,omitempty"`
	MinServers int64              `json:"min_servers,omitempty"`
	CloudRef   string             `json:"cloud_ref,omitempty"`
}

type PoolGroupMember struct {
	DeploymentState string `json:"deployment_state,omitempty"`
	PoolRef         string `json:"pool_ref,omitempty"`
	Ratio           int64  `json:"ratio,omitempty"`
}

type VirtualService struct {
	UUID                  string     `json:"uuid,omitempty"`
	Name                  string     `json:"name,omitempty"`
	URL                   string     `json:"url,omitempty"`
	Enabled               bool       `json:"enabled,omitempty"`
	Type                  string     `json:"type,omitempty"`
	ApplicationProfileRef string     `json:"application_profile_ref,omitempty"`
	NetworkProfileRef     string     `json:"network_profile_ref,omitempty"`
	NetworkRef            string     `json:"network_ref,omitempty"`
	PoolGroupRef          string     `json:"pool_group_ref,omitempty"`
	PoolRef               string     `json:"pool_ref,omitempty"`
	Services              []*Service `json:"services,omitempty"`
	VSVipRef              string     `json:"vsvip_ref,omitempty"`
	SEGroupRef            string     `json:"se_group_ref,omitempty"`
	CloudRef              string     `json:"cloud_ref,omitempty"`
}

type Service struct {
	Port        int32 `json:"port,omitempty"`
	EnableSSL   bool  `json:"enable_ssl,omitempty"`
	EnableHTTP2 bool  `json:"enable_http2,omitempty"`
}

type VSVIP struct {
	UUID     string `json:"uuid,omitempty"`
	Name     string `json:"name,omitempty"`
	URL      string `json:"url,omitempty"`
	VIP      []*VIP `json:"vip,omitempty"`
	CloudRef string `json:"cloud_ref,omitempty"`
}

type VSVIPList struct {
	Count   int64    `json:"count,omitempty"`
	Results []*VSVIP `json:"results,omitempty"`
}

type VIP struct {
	AutoAllocateFloatingIP bool                   `json:"auto_allocate_floating_ip,omitempty"`
	AutoAllocateIP         bool                   `json:"auto_allocate_ip,omitempty"`
	AutoAllocateIPType     string                 `json:"auto_allocate_ip_type,omitempty"`
	AVIAllocatedFIP        bool                   `json:"avi_allocated_fip,omitempty"`
	AVIAllocatedVIP        bool                   `json:"avi_allocated_vip,omitempty"`
	Enabled                bool                   `json:"enabled,omitempty"`
	IPAddress              *IPAddr                `json:"ip_address,omitempty"`
	IP6Address             *IPAddr                `json:"ip6_address,omitempty"`
	IPAMNetworkSubnet      *IPNetworkSubnet       `json:"ipam_network_subnet,omitempty"`
	NetworkRef             string                 `json:"network_ref,omitempty"`
	PlacementNetworks      []*VIPPlacementNetwork `json:"placement_networks,omitempty"`
	PrefixLength           int64                  `json:"prefix_length,omitempty"`
	Subnet                 *IPAddrPrefix          `json:"subnet,omitempty"`
	Subnet6                *IPAddrPrefix          `json:"subnet6,omitempty"`
}

type IPNetworkSubnet struct {
	NetworkRef string        `json:"network_ref,omitempty"`
	Subnet     *IPAddrPrefix `json:"subnet,omitempty"`
	Subnet6    *IPAddrPrefix `json:"subnet6,omitempty"`
}

type VIPPlacementNetwork struct {
	NetworkRef string        `json:"network_ref,omitempty"`
	Subnet     *IPAddrPrefix `json:"subnet,omitempty"`
	Subnet6    *IPAddrPrefix `json:"subnet6,omitempty"`
}

type IPAddrPrefix struct {
	IPAddr IPAddr `json:"ip_addr,omitempty"`
	Mask   int32  `json:"mask,omitempty"`
}

type IPAddr struct {
	Addr string `json:"addr,omitempty"`
	Type string `json:"type,omitempty"`
}

type ListVSVIPOptions struct {
	Name      string
	CloudName string
}

type ApplicationProfile struct {
	UUID               string `json:"uuid,omitempty"`
	Name               string `json:"name,omitempty"`
	URL                string `json:"url,omitempty"`
	Type               string `json:"type,omitempty"`
	TenantRef          string `json:"tenant_ref,omitempty"`
	PreserveClientIP   bool   `json:"preserve_client_ip,omitempty"`
	PreserveClientPort bool   `json:"preserve_client_port,omitempty"`
	PreserveDestIPPort bool   `json:"preserve_dest_ip_port,omitempty"`
}

type ApplicationProfileList struct {
	Count   int64                 `json:"count,omitempty"`
	Results []*ApplicationProfile `json:"results,omitempty"`
}

type ListApplicationProfileOptions struct {
	Name string
}

type ServiceEngineGroup struct {
	UUID string `json:"uuid,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
	// `PLACEMENT_ALGO_PACKED` (default) or `PLACEMENT_ALGO_DISTRIBUTED`
	// Reference: https://avinetworks.com/docs/20.1/elastic-ha-for-avi-service-engines/#elastic-ha-activeactive
	Algo               string `json:"algo,omitempty"`
	MinScaleoutPerVS   int64  `json:"min_scaleout_per_vs,omitempty"`
	MaxScaleoutPerVS   int64  `json:"max_scaleout_per_vs,omitempty"`
	MinSE              int64  `json:"min_se,omitempty"`
	MaxSE              int64  `json:"max_se,omitempty"`
	MaxVSPerSE         int64  `json:"max_vs_per_se,omitempty"`
	VCPUsPerSE         int64  `json:"vcpus_per_se,omitempty"`
	MemoryPerSE        int64  `json:"memory_per_se,omitempty"`
	HAMode             string `json:"ha_mode,omitempty"`
	SEDeprovisionDelay int64  `json:"se_deprovision_delay"`
	LicenseTier        string `json:"license_tier,omitempty"`
	CloudRef           string `json:"cloud_ref,omitempty"`
}

type ServiceEngineGroupList struct {
	Count   int64                 `json:"count,omitempty"`
	Results []*ServiceEngineGroup `json:"results,omitempty"`
}

type ListSEGroupOptions struct {
	Name      string
	CloudName string
}

type Network struct {
	UUID              string    `json:"uuid,omitempty"`
	Name              string    `json:"name,omitempty"`
	URL               string    `json:"url,omitempty"`
	ConfiguredSubnets []*Subnet `json:"configured_subnets,omitempty"`
	CloudRef          string    `json:"cloud_ref,omitempty"`
}

type NetworkList struct {
	Count   int64      `json:"count,omitempty"`
	Results []*Network `json:"results,omitempty"`
}

type NetworkProfile struct {
	UUID             string               `json:"uuid,omitempty"`
	Name             string               `json:"name,omitempty"`
	URL              string               `json:"url,omitempty"`
	Profile          *NetworkProfileUnion `json:"profile,omitempty"`
	ConnectionMirror bool                 `json:"connection_mirror,omitempty"`
	Description      string               `json:"description,omitempty"`
}

type NetworkProfileList struct {
	Count   int64             `json:"count,omitempty"`
	Results []*NetworkProfile `json:"results,omitempty"`
}

type ListNetworkProfileOptions struct {
	Name string
}

type NetworkProfileUnion struct {
	Type               string              `json:"type,omitempty"`
	TCPFastPathProfile *TCPFastPathProfile `json:"tcp_fast_path_profile,omitempty"`
	TCPProxyProfile    *TCPProxyProfile    `json:"tcp_proxy_profile,omitempty"`
	UDPFastPathProfile *UDPFastPathProfile `json:"udp_fast_path_profile,omitempty"`
	UDPProxyProfile    *UDPProxyProfile    `json:"udp_proxy_profile,omitempty"`
}

type TCPFastPathProfile struct {
	DSRProfile          *DSRProfile `json:"dsr_profile,omitempty"`
	EnableSYNProtection bool        `json:"enable_syn_protection,omitempty"`
	SessionIdleTimeout  int32       `json:"session_idle_timeout,omitempty"`
}

type TCPProxyProfile struct {
	AggressiveCongestionAvoidance   bool   `json:"aggressive_congestion_avoidance,omitempty"`
	AutoWindowGrowth                bool   `json:"auto_window_growth,omitempty"`
	Automatic                       bool   `json:"automatic,omitempty"`
	CCAlgo                          string `json:"cc_algo,omitempty"`
	CongestionRecoveryScalingFactor int32  `json:"congestion_recovery_scaling_factor,omitempty"`
	IdleConnectionTimeout           int32  `json:"idle_connection_timeout,omitempty"`
	IdleConnectionType              string `json:"idle_connection_type,omitempty"`
	IgnoreTimeWait                  bool   `json:"ignore_time_wait,omitempty"`
	IpDscp                          int32  `json:"ip_dscp,omitempty"`
	KeepaliveInHalfcloseState       bool   `json:"keepalive_in_halfclose_state,omitempty"`
	MaxRetransmissions              int32  `json:"max_retransmissions,omitempty"`
	MaxSegmentSize                  int32  `json:"max_segment_size,omitempty"`
	MaxSynRetransmissions           int32  `json:"max_syn_retransmissions,omitempty"`
	MinRexmtTimeout                 int32  `json:"min_rexmt_timeout,omitempty"`
	NaglesAlgorithm                 bool   `json:"nagles_algorithm,omitempty"`
	ReassemblyQueueSize             int32  `json:"reassembly_queue_size,omitempty"`
	ReceiveWindow                   int32  `json:"receive_window,omitempty"`
	ReorderThreshold                int32  `json:"reorder_threshold,omitempty"`
	SlowStartScalingFactor          int32  `json:"slow_start_scaling_factor,omitempty"`
	TimeWaitDelay                   int32  `json:"time_wait_delay,omitempty"`
	UseInterfaceMTU                 bool   `json:"use_interface_mtu,omitempty"`
}

type UDPFastPathProfile struct {
	DSRProfile         *DSRProfile `json:"dsr_profile,omitempty"`
	PerPktLoadbalance  bool        `json:"per_pkt_loadbalance,omitempty"`
	SessionIdleTimeout int32       `json:"session_idle_timeout,omitempty"`
	SNAT               bool        `json:"snat,omitempty"`
}

type UDPProxyProfile struct {
	SessionIdleTimeout int32 `json:"session_idle_timeout,omitempty"`
}

type DSRProfile struct {
	DSREncapType string `json:"dsr_encap_type,omitempty"`
	DSRType      string `json:"dsr_type,omitempty"`
}

type Subnet struct {
	Prefix         IPAddrPrefix     `json:"prefix,omitempty"`
	StaticIPRanges []*StaticIPRange `json:"static_ip_ranges,omitempty"`
	StaticIPs      []*IPAddr        `json:"static_ips,omitempty"`
	StaticRanges   []*IPAddrRange   `json:"static_ranges,omitempty"`
}

type StaticIPRange struct {
	Range IPAddrRange `json:"range,omitempty"`
	Type  string      `json:"type,omitempty"`
}

type IPAddrRange struct {
	Begin IPAddr `json:"begin,omitempty"`
	End   IPAddr `json:"end,omitempty"`
}

type ListNetworkOptions struct {
	Name      string
	CloudName string
}

type NetworkRuntime struct {
	UUID          string           `json:"uuid,omitempty"`
	Name          string           `json:"name,omitempty"`
	URL           string           `json:"url,omitempty"`
	SubnetRuntime []*SubnetRuntime `json:"subnet_runtime,omitempty"`
}

type SubnetRuntime struct {
	Prefix          IPAddrPrefix            `json:"prefix,omitempty"`
	IPRangeRuntimes []*StaticIPRangeRuntime `json:"ip_range_runtimes,omitempty"`
}

type StaticIPRangeRuntime struct {
	FreeIPCount  int32  `json:"free_ip_count,omitempty"`
	TotalIPCount int32  `json:"total_ip_count,omitempty"`
	Type         string `json:"type,omitempty"`
}

type LicenseLedgerDetails struct {
	UUID       string              `json:"uuid,omitempty"`
	TierUsages []*LicenseTierUsage `json:"tier_usages,omitempty"`
	SEInfos    []*LicenseInfo      `json:"se_infos,omitempty"`
}

type LicenseTierUsage struct {
	Tier  string        `json:"tier,omitempty"`
	Usage *LicenseUsage `json:"usage,omitempty"`
}

type LicenseUsage struct {
	Available float64 `json:"available,omitempty"`
	Consumed  float64 `json:"consumed,omitempty"`
	Escrow    float64 `json:"escrow,omitempty"`
	Remaining float64 `json:"remaining,omitempty"`
}

type LicenseInfo struct {
	UUID         string  `json:"uuid,omitempty"`
	TenantUUID   string  `json:"tenant_uuid,omitempty"`
	Tier         string  `json:"tier,omitempty"`
	ServiceCores float64 `json:"service_cores,omitempty"`
	LastUpdated  float64 `json:"last_updated,omitempty"`
}

type Cloud struct {
	UUID  string `json:"uuid,omitempty"`
	Name  string `json:"name,omitempty"`
	VType string `json:"vtype,omitempty"`
}

type CloudList struct {
	Count   int64    `json:"count,omitempty"`
	Results []*Cloud `json:"results,omitempty"`
}

type ListCloudOptions struct {
	Name string
}
