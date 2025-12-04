package models

import (
	"gorm.io/gorm"
	"time"
)

type OciUser struct {
	ID               string     `gorm:"primaryKey;column:id" json:"id"`
	Username         string     `gorm:"column:username" json:"username"`
	TenantName       string     `gorm:"column:tenant_name" json:"tenantName"`
	TenantCreateTime *time.Time `gorm:"column:tenant_create_time" json:"tenantCreateTime"`
	OciTenantID      string     `gorm:"column:oci_tenant_id" json:"ociTenantId"`
	OciUserID        string     `gorm:"column:oci_user_id" json:"ociUserId"`
	OciFingerprint   string     `gorm:"column:oci_fingerprint" json:"ociFingerprint"`
	OciRegion        string     `gorm:"column:oci_region" json:"ociRegion"`
	OciKeyPath       string     `gorm:"column:oci_key_path" json:"ociKeyPath"`
	CreateTime       time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

// OciUserListResponse 配置列表响应
type OciUserListResponse struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	TenantName       string `json:"tenantName"`
	TenantCreateTime string `json:"tenantCreateTime"`
	OciTenantID      string `json:"ociTenantId"`
	OciRegion        string `json:"ociRegion"`
	CreateTime       string `json:"createTime"`
	InstanceCount    int    `json:"instanceCount"`
	RunningInstances int    `json:"runningInstances"`
}

// OciConfigDetails 配置详情响应
type OciConfigDetails struct {
	UserID      string         `json:"userId"`
	Username    string         `json:"username"`
	TenantID    string         `json:"tenantId"`
	TenantName  string         `json:"tenantName"`
	Fingerprint string         `json:"fingerprint"`
	KeyPath     string         `json:"keyPath"`
	Region      string         `json:"region"`
	CreateTime  string         `json:"createTime"`
	Instances   []InstanceInfo `json:"instances"`
	Volumes     []VolumeInfo   `json:"volumes"`
	VCNs        []VCNInfo      `json:"vcns"`
}

// InstanceInfo 实例信息
type InstanceInfo struct {
	ID                 string     `json:"id"`
	DisplayName        string     `json:"displayName"`
	State              string     `json:"state"`
	Shape              string     `json:"shape"`
	Ocpus              float32    `json:"ocpus"`
	Memory             float32    `json:"memory"`
	PublicIPs          []string   `json:"publicIps"`
	PrivateIPs         []string   `json:"privateIps"`
	IPv6               string     `json:"ipv6"`
	Region             string     `json:"region"`
	AvailabilityDomain string     `json:"availabilityDomain"`
	BootVolumeSize     int64      `json:"bootVolumeSize"`
	BootVolumeVpu      int64      `json:"bootVolumeVpu"`
	ImageName          string     `json:"imageName"`
	CreateTime         string     `json:"createTime"`
	VnicList           []VnicInfo `json:"vnicList"`
}

// VnicInfo VNIC信息
type VnicInfo struct {
	VnicID    string `json:"vnicId"`
	Name      string `json:"name"`
	PublicIP  string `json:"publicIp"`
	PrivateIP string `json:"privateIp"`
	SubnetID  string `json:"subnetId"`
}

// VolumeInfo 卷信息
type VolumeInfo struct {
	ID                 string `json:"id"`
	DisplayName        string `json:"displayName"`
	SizeInGBs          int64  `json:"sizeInGBs"`
	VpusPerGB          int64  `json:"vpusPerGB"`
	State              string `json:"state"`
	AvailabilityDomain string `json:"availabilityDomain"`
	InstanceName       string `json:"instanceName"`
	Attached           bool   `json:"attached"`
	CreateTime         string `json:"createTime"`
}

// VCNInfo VCN信息
type VCNInfo struct {
	ID          string       `json:"id"`
	DisplayName string       `json:"displayName"`
	CIDRBlock   string       `json:"cidrBlock"`
	State       string       `json:"state"`
	CreateTime  string       `json:"createTime"`
	Subnets     []SubnetInfo `json:"subnets"`
}

// SubnetInfo 子网信息
type SubnetInfo struct {
	ID                 string `json:"id"`
	DisplayName        string `json:"displayName"`
	CIDRBlock          string `json:"cidrBlock"`
	AvailabilityDomain string `json:"availabilityDomain"`
	State              string `json:"state"`
	IsPublic           bool   `json:"isPublic"`
}

// SecurityListInfo 安全列表信息
type SecurityListInfo struct {
	ID           string         `json:"id"`
	DisplayName  string         `json:"displayName"`
	VcnId        string         `json:"vcnId"`
	IngressRules []SecurityRule `json:"ingressRules"`
	EgressRules  []SecurityRule `json:"egressRules"`
}

// SecurityRule 安全规则
type SecurityRule struct {
	IsStateless  bool   `json:"isStateless"`
	Protocol     string `json:"protocol"`
	ProtocolName string `json:"protocolName"`
	Source       string `json:"source"`
	Destination  string `json:"destination"`
	PortRangeMin int    `json:"portRangeMin"`
	PortRangeMax int    `json:"portRangeMax"`
	IcmpType     *int   `json:"icmpType"`
	IcmpCode     *int   `json:"icmpCode"`
	Description  string `json:"description"`
}

// TenantInfo 租户详情
type TenantInfo struct {
	ID                   string           `json:"id"`
	Name                 string           `json:"name"`
	Description          string           `json:"description"`
	HomeRegionKey        string           `json:"homeRegionKey"`
	Regions              []string         `json:"regions"`
	CreateTime           string           `json:"createTime"`
	PasswordExpiresAfter int              `json:"passwordExpiresAfter"` // 密码过期天数，0表示永不过期
	UserList             []TenantUserInfo `json:"userList"`
}

// TenantUserInfo 租户用户信息
type TenantUserInfo struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	Email                   string `json:"email"`
	State                   string `json:"state"`
	EmailVerified           bool   `json:"emailVerified"`
	IsMfaActivated          bool   `json:"isMfaActivated"`
	CreateTime              string `json:"createTime"`
	LastSuccessfulLoginTime string `json:"lastSuccessfulLoginTime"`
}

// TrafficData 流量数据
type TrafficData struct {
	Time     []string `json:"time"`
	Inbound  []string `json:"inbound"`
	Outbound []string `json:"outbound"`
}

// TrafficCondition 流量查询条件
type TrafficCondition struct {
	Regions   []ValueLabel `json:"regions"`
	Instances []ValueLabel `json:"instances"`
}

// ValueLabel 值标签对
type ValueLabel struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

func (OciUser) TableName() string {
	return "oci_user"
}

type OciCreateTask struct {
	ID              string    `gorm:"primaryKey;column:id" json:"id"`
	UserID          string    `gorm:"column:user_id" json:"userId"`
	OciRegion       string    `gorm:"column:oci_region" json:"ociRegion"`
	Ocpus           float64   `gorm:"column:ocpus;default:1.0" json:"ocpus"`
	Memory          float64   `gorm:"column:memory;default:6.0" json:"memory"`
	Disk            int       `gorm:"column:disk;default:50" json:"disk"`
	Architecture    string    `gorm:"column:architecture;default:ARM" json:"architecture"`
	Interval        int       `gorm:"column:interval;default:60" json:"interval"`
	CreateNumbers   int       `gorm:"column:create_numbers;default:1" json:"createNumbers"`
	SSHKeyID        string    `gorm:"column:ssh_key_id" json:"sshKeyId"`
	OperationSystem string    `gorm:"column:operation_system;default:Ubuntu" json:"operationSystem"`
	CreateTime      time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (OciCreateTask) TableName() string {
	return "oci_create_task"
}

type OciKv struct {
	ID         string    `gorm:"primaryKey;column:id" json:"id"`
	Code       string    `gorm:"column:code;not null" json:"code"`
	Value      string    `gorm:"column:value;type:text" json:"value"`
	Type       string    `gorm:"column:type;not null" json:"type"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (OciKv) TableName() string {
	return "oci_kv"
}

type CfCfg struct {
	ID         string    `gorm:"primaryKey;column:id" json:"id"`
	Domain     string    `gorm:"column:domain;not null" json:"domain"`
	ZoneID     string    `gorm:"column:zone_id;not null" json:"zoneId"`
	APIToken   string    `gorm:"column:api_token;not null" json:"apiToken"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (CfCfg) TableName() string {
	return "cf_cfg"
}

type IpData struct {
	ID         string    `gorm:"primaryKey;column:id" json:"id"`
	IP         string    `gorm:"column:ip;not null" json:"ip"`
	Country    string    `gorm:"column:country" json:"country"`
	Area       string    `gorm:"column:area" json:"area"`
	City       string    `gorm:"column:city" json:"city"`
	Org        string    `gorm:"column:org" json:"org"`
	Asn        string    `gorm:"column:asn" json:"asn"`
	Type       string    `gorm:"column:type" json:"type"`
	Lat        float64   `gorm:"column:lat" json:"lat"`
	Lng        float64   `gorm:"column:lng" json:"lng"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (IpData) TableName() string {
	return "ip_data"
}

// SysSetting 系统设置表
type SysSetting struct {
	ID    string `gorm:"primaryKey;column:id" json:"id"`
	Key   string `gorm:"column:key;uniqueIndex;not null" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (SysSetting) TableName() string {
	return "sys_setting"
}

// OciConfigCache 配置缓存表
type OciConfigCache struct {
	ID               string    `gorm:"primaryKey;column:id" json:"id"`
	ConfigID         string    `gorm:"column:config_id;uniqueIndex;not null" json:"configId"`
	InstanceCount    int       `gorm:"column:instance_count;default:0" json:"instanceCount"`
	RunningInstances int       `gorm:"column:running_instances;default:0" json:"runningInstances"`
	InstancesData    string    `gorm:"column:instances_data;type:text" json:"instancesData"`
	VolumesData      string    `gorm:"column:volumes_data;type:text" json:"volumesData"`
	VcnsData         string    `gorm:"column:vcns_data;type:text" json:"vcnsData"`
	TenantData       string    `gorm:"column:tenant_data;type:text" json:"tenantData"`
	UpdateTime       time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (OciConfigCache) TableName() string {
	return "oci_config_cache"
}

// SSHKey SSH密钥表
type SSHKey struct {
	ID         string    `gorm:"primaryKey;column:id" json:"id"`
	Name       string    `gorm:"column:name;not null" json:"name"`
	PublicKey  string    `gorm:"column:public_key;type:text;not null" json:"publicKey"`
	PrivateKey string    `gorm:"column:private_key;type:text" json:"privateKey"`
	KeyType    string    `gorm:"column:key_type;not null" json:"keyType"` // config: 配置关联, standalone: 独立上传
	ConfigID   string    `gorm:"column:config_id" json:"configId"`        // 关联的配置ID，独立上传时为空
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (SSHKey) TableName() string {
	return "ssh_key"
}

// SSHKeyResponse SSH密钥响应
type SSHKeyResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PublicKey  string `json:"publicKey"`
	KeyType    string `json:"keyType"`
	ConfigID   string `json:"configId"`
	ConfigName string `json:"configName"`
	CreateTime string `json:"createTime"`
}

type ResponseData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func SuccessResponse(data interface{}, message string) ResponseData {
	if message == "" {
		message = "success"
	}
	return ResponseData{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

func ErrorResponse(code int, message string) ResponseData {
	return ResponseData{
		Code:    code,
		Message: message,
	}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&OciUser{},
		&OciCreateTask{},
		&OciKv{},
		&CfCfg{},
		&IpData{},
		&SysSetting{},
		&OciConfigCache{},
		&SSHKey{},
	)
}
