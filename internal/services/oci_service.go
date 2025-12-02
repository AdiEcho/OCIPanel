package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/adiecho/oci-panel/internal/config"
	"github.com/adiecho/oci-panel/internal/models"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/identitydomains"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
)

type OCIService struct {
	cfg *config.Config
}

func NewOCIService(cfg *config.Config) *OCIService {
	return &OCIService{cfg: cfg}
}

// 辅助函数：创建字符串指针
func stringPtr(s string) *string {
	return &s
}

// 辅助函数：创建布尔指针
func boolPtr(b bool) *bool {
	return &b
}

func (s *OCIService) GetConfigProvider(user *models.OciUser) (common.ConfigurationProvider, error) {
	keyPath := "keys/" + user.OciKeyPath
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	return common.NewRawConfigurationProvider(
		user.OciTenantID,
		user.OciUserID,
		user.OciRegion,
		user.OciFingerprint,
		string(privateKey),
		nil,
	), nil
}

func (s *OCIService) GetComputeClient(user *models.OciUser) (core.ComputeClient, error) {
	configProvider, err := s.GetConfigProvider(user)
	if err != nil {
		return core.ComputeClient{}, err
	}

	client, err := core.NewComputeClientWithConfigurationProvider(configProvider)
	if err != nil {
		return core.ComputeClient{}, err
	}

	return client, nil
}

func (s *OCIService) GetVirtualNetworkClient(user *models.OciUser) (core.VirtualNetworkClient, error) {
	configProvider, err := s.GetConfigProvider(user)
	if err != nil {
		return core.VirtualNetworkClient{}, err
	}

	client, err := core.NewVirtualNetworkClientWithConfigurationProvider(configProvider)
	if err != nil {
		return core.VirtualNetworkClient{}, err
	}

	return client, nil
}

func (s *OCIService) GetBlockstorageClient(user *models.OciUser) (core.BlockstorageClient, error) {
	configProvider, err := s.GetConfigProvider(user)
	if err != nil {
		return core.BlockstorageClient{}, err
	}

	client, err := core.NewBlockstorageClientWithConfigurationProvider(configProvider)
	if err != nil {
		return core.BlockstorageClient{}, err
	}

	return client, nil
}

func (s *OCIService) GetIdentityClient(user *models.OciUser) (identity.IdentityClient, error) {
	configProvider, err := s.GetConfigProvider(user)
	if err != nil {
		return identity.IdentityClient{}, err
	}

	client, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
	if err != nil {
		return identity.IdentityClient{}, err
	}

	return client, nil
}

func (s *OCIService) GetIdentityDomainsClient(user *models.OciUser, endpoint string) (identitydomains.IdentityDomainsClient, error) {
	configProvider, err := s.GetConfigProvider(user)
	if err != nil {
		return identitydomains.IdentityDomainsClient{}, err
	}

	client, err := identitydomains.NewIdentityDomainsClientWithConfigurationProvider(configProvider, endpoint)
	if err != nil {
		return identitydomains.IdentityDomainsClient{}, err
	}

	return client, nil
}

func (s *OCIService) ListInstances(ctx context.Context, user *models.OciUser, compartmentId string) ([]core.Instance, error) {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return nil, err
	}

	req := core.ListInstancesRequest{
		CompartmentId: &compartmentId,
	}

	resp, err := client.ListInstances(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}

func (s *OCIService) GetInstance(ctx context.Context, user *models.OciUser, instanceId string) (*core.Instance, error) {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return nil, err
	}

	req := core.GetInstanceRequest{
		InstanceId: &instanceId,
	}

	resp, err := client.GetInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	return &resp.Instance, nil
}

func (s *OCIService) InstanceAction(ctx context.Context, user *models.OciUser, instanceId string, action string) error {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return err
	}

	req := core.InstanceActionRequest{
		InstanceId: &instanceId,
		Action:     core.InstanceActionActionEnum(action),
	}

	_, err = client.InstanceAction(ctx, req)
	return err
}

func (s *OCIService) TerminateInstance(ctx context.Context, user *models.OciUser, instanceId string) error {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return err
	}

	req := core.TerminateInstanceRequest{
		InstanceId: &instanceId,
	}

	_, err = client.TerminateInstance(ctx, req)
	return err
}

func (s *OCIService) UpdateInstance(ctx context.Context, user *models.OciUser, instanceId string, displayName string) error {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return err
	}

	req := core.UpdateInstanceRequest{
		InstanceId: &instanceId,
		UpdateInstanceDetails: core.UpdateInstanceDetails{
			DisplayName: &displayName,
		},
	}

	_, err = client.UpdateInstance(ctx, req)
	return err
}

type LaunchInstanceParams struct {
	CompartmentId      string
	AvailabilityDomain string
	DisplayName        string
	ImageId            string
	Shape              string
	SubnetId           string
	Ocpus              float32
	MemoryInGBs        float32
	SshPublicKey       string
}

func (s *OCIService) LaunchInstance(ctx context.Context, user *models.OciUser, params LaunchInstanceParams) (*core.Instance, error) {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return nil, err
	}

	req := core.LaunchInstanceRequest{
		LaunchInstanceDetails: core.LaunchInstanceDetails{
			CompartmentId:      &params.CompartmentId,
			AvailabilityDomain: &params.AvailabilityDomain,
			DisplayName:        &params.DisplayName,
			SourceDetails: &core.InstanceSourceViaImageDetails{
				ImageId: &params.ImageId,
			},
			Shape: &params.Shape,
			CreateVnicDetails: &core.CreateVnicDetails{
				SubnetId: &params.SubnetId,
			},
			ShapeConfig: &core.LaunchInstanceShapeConfigDetails{
				Ocpus:       &params.Ocpus,
				MemoryInGBs: &params.MemoryInGBs,
			},
			Metadata: map[string]string{
				"ssh_authorized_keys": params.SshPublicKey,
			},
		},
	}

	resp, err := client.LaunchInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	return &resp.Instance, nil
}

// GetInstanceDetails 获取实例详细信息包括VNICs
func (s *OCIService) GetInstanceDetails(ctx context.Context, user *models.OciUser, instanceId string) (*models.InstanceInfo, error) {
	computeClient, err := s.GetComputeClient(user)
	if err != nil {
		return nil, err
	}

	// 获取实例基本信息
	instance, err := s.GetInstance(ctx, user, instanceId)
	if err != nil {
		return nil, err
	}

	info := &models.InstanceInfo{
		ID:                 *instance.Id,
		DisplayName:        *instance.DisplayName,
		State:              string(instance.LifecycleState),
		Shape:              *instance.Shape,
		Region:             user.OciRegion,
		AvailabilityDomain: *instance.AvailabilityDomain,
		CreateTime:         instance.TimeCreated.Format("2006-01-02 15:04:05"),
		PublicIPs:          []string{},
		PrivateIPs:         []string{},
		VnicList:           []models.VnicInfo{},
	}

	// 获取实例规格配置
	if instance.ShapeConfig != nil {
		if instance.ShapeConfig.Ocpus != nil {
			info.Ocpus = *instance.ShapeConfig.Ocpus
		}
		if instance.ShapeConfig.MemoryInGBs != nil {
			info.Memory = *instance.ShapeConfig.MemoryInGBs
		}
	}

	// 获取镜像信息
	if instance.SourceDetails != nil {
		switch sourceDetails := instance.SourceDetails.(type) {
		case core.InstanceSourceViaImageDetails:
			if sourceDetails.ImageId != nil {
				imageReq := core.GetImageRequest{ImageId: sourceDetails.ImageId}
				imageResp, err := computeClient.GetImage(ctx, imageReq)
				if err == nil && imageResp.DisplayName != nil {
					info.ImageName = *imageResp.DisplayName
				}
			}
		}
	}

	// 获取引导卷信息
	bootVolumeClient, err := s.GetBlockstorageClient(user)
	if err == nil {
		bootVolumeReq := core.ListBootVolumeAttachmentsRequest{
			CompartmentId:      instance.CompartmentId,
			InstanceId:         instance.Id,
			AvailabilityDomain: instance.AvailabilityDomain,
		}
		bootVolumeResp, err := computeClient.ListBootVolumeAttachments(ctx, bootVolumeReq)
		if err == nil && len(bootVolumeResp.Items) > 0 {
			bvId := bootVolumeResp.Items[0].BootVolumeId
			if bvId != nil {
				bvReq := core.GetBootVolumeRequest{BootVolumeId: bvId}
				bvResp, err := bootVolumeClient.GetBootVolume(ctx, bvReq)
				if err == nil {
					if bvResp.SizeInGBs != nil {
						info.BootVolumeSize = *bvResp.SizeInGBs
					}
					if bvResp.VpusPerGB != nil {
						info.BootVolumeVpu = *bvResp.VpusPerGB
					}
				}
			}
		}
	}

	// 获取VNIC附件
	listVnicReq := core.ListVnicAttachmentsRequest{
		CompartmentId: instance.CompartmentId,
		InstanceId:    instance.Id,
	}
	vnicResp, err := computeClient.ListVnicAttachments(ctx, listVnicReq)
	if err == nil {
		vnClient, err := s.GetVirtualNetworkClient(user)
		if err == nil {
			for _, vnicAttachment := range vnicResp.Items {
				if vnicAttachment.VnicId != nil {
					vnicReq := core.GetVnicRequest{VnicId: vnicAttachment.VnicId}
					vnic, err := vnClient.GetVnic(ctx, vnicReq)
					if err == nil {
						vnicInfo := models.VnicInfo{
							VnicID: *vnicAttachment.VnicId,
						}
						if vnic.DisplayName != nil {
							vnicInfo.Name = *vnic.DisplayName
						}
						if vnic.SubnetId != nil {
							vnicInfo.SubnetID = *vnic.SubnetId
						}
						if vnic.PublicIp != nil && *vnic.PublicIp != "" {
							vnicInfo.PublicIP = *vnic.PublicIp
							info.PublicIPs = append(info.PublicIPs, *vnic.PublicIp)
						}
						if vnic.PrivateIp != nil && *vnic.PrivateIp != "" {
							vnicInfo.PrivateIP = *vnic.PrivateIp
							info.PrivateIPs = append(info.PrivateIPs, *vnic.PrivateIp)
						}
						info.VnicList = append(info.VnicList, vnicInfo)

						// 获取IPv6地址
						ipv6Req := core.ListIpv6sRequest{VnicId: vnicAttachment.VnicId}
						ipv6Resp, err := vnClient.ListIpv6s(ctx, ipv6Req)
						if err == nil && len(ipv6Resp.Items) > 0 {
							if ipv6Resp.Items[0].IpAddress != nil {
								info.IPv6 = *ipv6Resp.Items[0].IpAddress
							}
						}
					}
				}
			}
		}
	}

	return info, nil
}

// ListBootVolumes 列出引导卷
func (s *OCIService) ListBootVolumes(ctx context.Context, user *models.OciUser, compartmentId string) ([]models.VolumeInfo, error) {
	client, err := s.GetBlockstorageClient(user)
	if err != nil {
		return nil, err
	}

	computeClient, err := s.GetComputeClient(user)
	if err != nil {
		return nil, err
	}

	req := core.ListBootVolumesRequest{
		CompartmentId: &compartmentId,
	}

	resp, err := client.ListBootVolumes(ctx, req)
	if err != nil {
		return nil, err
	}

	volumes := make([]models.VolumeInfo, 0)
	for _, bv := range resp.Items {
		volume := models.VolumeInfo{
			ID:          *bv.Id,
			DisplayName: *bv.DisplayName,
			State:       string(bv.LifecycleState),
		}
		if bv.SizeInGBs != nil {
			volume.SizeInGBs = *bv.SizeInGBs
		}
		if bv.VpusPerGB != nil {
			volume.VpusPerGB = *bv.VpusPerGB
		}
		if bv.AvailabilityDomain != nil {
			volume.AvailabilityDomain = *bv.AvailabilityDomain
		}
		if bv.TimeCreated != nil {
			volume.CreateTime = bv.TimeCreated.Format("2006-01-02 15:04:05")
		}

		// 检查是否已附加到实例
		attachReq := core.ListBootVolumeAttachmentsRequest{
			CompartmentId:      &compartmentId,
			BootVolumeId:       bv.Id,
			AvailabilityDomain: bv.AvailabilityDomain,
		}
		attachResp, err := computeClient.ListBootVolumeAttachments(ctx, attachReq)
		if err == nil && len(attachResp.Items) > 0 {
			volume.Attached = true
			if attachResp.Items[0].InstanceId != nil {
				// 获取实例名称
				instReq := core.GetInstanceRequest{InstanceId: attachResp.Items[0].InstanceId}
				instResp, err := computeClient.GetInstance(ctx, instReq)
				if err == nil && instResp.DisplayName != nil {
					volume.InstanceName = *instResp.DisplayName
				}
			}
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

// ListVCNs 列出虚拟云网络
func (s *OCIService) ListVCNs(ctx context.Context, user *models.OciUser, compartmentId string) ([]models.VCNInfo, error) {
	client, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return nil, err
	}

	req := core.ListVcnsRequest{
		CompartmentId: &compartmentId,
	}

	resp, err := client.ListVcns(ctx, req)
	if err != nil {
		return nil, err
	}

	vcns := make([]models.VCNInfo, 0)
	for _, vcn := range resp.Items {
		vcnInfo := models.VCNInfo{
			ID:          *vcn.Id,
			DisplayName: *vcn.DisplayName,
			State:       string(vcn.LifecycleState),
			Subnets:     []models.SubnetInfo{},
		}
		// 使用 CidrBlocks 替代已弃用的 CidrBlock
		if len(vcn.CidrBlocks) > 0 {
			vcnInfo.CIDRBlock = vcn.CidrBlocks[0]
		}
		if vcn.TimeCreated != nil {
			vcnInfo.CreateTime = vcn.TimeCreated.Format("2006-01-02 15:04:05")
		}

		// 获取子网列表
		subnetReq := core.ListSubnetsRequest{
			CompartmentId: &compartmentId,
			VcnId:         vcn.Id,
		}
		subnetResp, err := client.ListSubnets(ctx, subnetReq)
		if err == nil {
			for _, subnet := range subnetResp.Items {
				subnetInfo := models.SubnetInfo{
					ID:    *subnet.Id,
					State: string(subnet.LifecycleState),
				}
				if subnet.DisplayName != nil {
					subnetInfo.DisplayName = *subnet.DisplayName
				}
				// 使用 Ipv4CidrBlocks 替代已弃用的 CidrBlock
				if len(subnet.Ipv4CidrBlocks) > 0 {
					subnetInfo.CIDRBlock = subnet.Ipv4CidrBlocks[0]
				}
				if subnet.AvailabilityDomain != nil {
					subnetInfo.AvailabilityDomain = *subnet.AvailabilityDomain
				}
				if subnet.ProhibitPublicIpOnVnic != nil {
					subnetInfo.IsPublic = !*subnet.ProhibitPublicIpOnVnic
				}
				vcnInfo.Subnets = append(vcnInfo.Subnets, subnetInfo)
			}
		}

		vcns = append(vcns, vcnInfo)
	}

	return vcns, nil
}

// ChangePublicIP 更改实例公网IP
func (s *OCIService) ChangePublicIP(ctx context.Context, user *models.OciUser, vnicId string) (string, error) {
	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return "", err
	}

	// 获取当前VNIC信息
	vnicReq := core.GetVnicRequest{VnicId: &vnicId}
	vnicResp, err := vnClient.GetVnic(ctx, vnicReq)
	if err != nil {
		return "", fmt.Errorf("failed to get VNIC: %w", err)
	}

	// 检查是否有公网IP
	if vnicResp.PublicIp == nil || *vnicResp.PublicIp == "" {
		return "", fmt.Errorf("VNIC does not have a public IP")
	}

	// 获取现有的public IP对象
	listPublicIpsReq := core.ListPublicIpsRequest{
		Scope:         core.ListPublicIpsScopeRegion,
		CompartmentId: &user.OciTenantID,
		Lifetime:      core.ListPublicIpsLifetimeEphemeral,
	}
	publicIpsResp, err := vnClient.ListPublicIps(ctx, listPublicIpsReq)
	if err != nil {
		return "", fmt.Errorf("failed to list public IPs: %w", err)
	}

	// 找到对应的publicIP对象
	var publicIpId *string
	for _, pip := range publicIpsResp.Items {
		if pip.AssignedEntityId != nil && *pip.AssignedEntityId == vnicId {
			publicIpId = pip.Id
			break
		}
	}

	if publicIpId == nil {
		return "", fmt.Errorf("could not find public IP object for VNIC")
	}

	// 删除现有公网IP
	deleteReq := core.DeletePublicIpRequest{PublicIpId: publicIpId}
	_, err = vnClient.DeletePublicIp(ctx, deleteReq)
	if err != nil {
		return "", fmt.Errorf("failed to delete public IP: %w", err)
	}

	// 等待一下确保删除完成
	time.Sleep(2 * time.Second)

	// 创建新的临时公网IP
	createReq := core.CreatePublicIpRequest{
		CreatePublicIpDetails: core.CreatePublicIpDetails{
			CompartmentId: &user.OciTenantID,
			Lifetime:      core.CreatePublicIpDetailsLifetimeEphemeral,
			PrivateIpId:   vnicResp.PrivateIp,
		},
	}
	createResp, err := vnClient.CreatePublicIp(ctx, createReq)
	if err != nil {
		return "", fmt.Errorf("failed to create new public IP: %w", err)
	}

	if createResp.IpAddress == nil {
		return "", fmt.Errorf("new public IP address is nil")
	}

	return *createResp.IpAddress, nil
}

// UpdateInstanceShape 更新实例配置（CPU和内存）
func (s *OCIService) UpdateInstanceShape(ctx context.Context, user *models.OciUser, instanceId string, ocpus float32, memoryInGBs float32) error {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return err
	}

	// 获取当前实例信息
	instance, err := s.GetInstance(ctx, user, instanceId)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// 确保实例已停止
	if instance.LifecycleState != core.InstanceLifecycleStateStopped {
		return fmt.Errorf("instance must be stopped before updating shape config")
	}

	req := core.UpdateInstanceRequest{
		InstanceId: &instanceId,
		UpdateInstanceDetails: core.UpdateInstanceDetails{
			ShapeConfig: &core.UpdateInstanceShapeConfigDetails{
				Ocpus:       &ocpus,
				MemoryInGBs: &memoryInGBs,
			},
		},
	}

	_, err = client.UpdateInstance(ctx, req)
	return err
}

// UpdateBootVolume 更新引导卷配置
func (s *OCIService) UpdateBootVolume(ctx context.Context, user *models.OciUser, bootVolumeId string, sizeInGBs int64, vpusPerGB int64) error {
	client, err := s.GetBlockstorageClient(user)
	if err != nil {
		return err
	}

	req := core.UpdateBootVolumeRequest{
		BootVolumeId: &bootVolumeId,
		UpdateBootVolumeDetails: core.UpdateBootVolumeDetails{
			SizeInGBs: &sizeInGBs,
			VpusPerGB: &vpusPerGB,
		},
	}

	_, err = client.UpdateBootVolume(ctx, req)
	return err
}

// CreateConsoleConnection 创建控制台连接
func (s *OCIService) CreateConsoleConnection(ctx context.Context, user *models.OciUser, instanceId string, publicKey string) (string, error) {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return "", err
	}

	req := core.CreateInstanceConsoleConnectionRequest{
		CreateInstanceConsoleConnectionDetails: core.CreateInstanceConsoleConnectionDetails{
			InstanceId: &instanceId,
			PublicKey:  &publicKey,
		},
	}

	resp, err := client.CreateInstanceConsoleConnection(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create console connection: %w", err)
	}

	if resp.Id == nil {
		return "", fmt.Errorf("console connection ID is nil")
	}

	return *resp.Id, nil
}

// GetConsoleConnectionString 获取控制台连接字符串
func (s *OCIService) GetConsoleConnectionString(ctx context.Context, user *models.OciUser, connectionId string) (string, error) {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return "", err
	}

	// 等待连接激活
	var connectionString string
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		req := core.GetInstanceConsoleConnectionRequest{
			InstanceConsoleConnectionId: &connectionId,
		}
		resp, err := client.GetInstanceConsoleConnection(ctx, req)
		if err != nil {
			return "", fmt.Errorf("failed to get console connection: %w", err)
		}

		if resp.LifecycleState == core.InstanceConsoleConnectionLifecycleStateActive {
			if resp.ConnectionString != nil {
				connectionString = *resp.ConnectionString
				break
			}
		}

		time.Sleep(2 * time.Second)
	}

	if connectionString == "" {
		return "", fmt.Errorf("console connection did not become active within timeout")
	}

	return connectionString, nil
}

// GetTenantInfo 获取租户详情
func (s *OCIService) GetTenantInfo(ctx context.Context, user *models.OciUser) (*models.TenantInfo, error) {
	identityClient, err := s.GetIdentityClient(user)
	if err != nil {
		return nil, err
	}

	// 获取租户信息
	tenantReq := identity.GetTenancyRequest{TenancyId: &user.OciTenantID}
	tenantResp, err := identityClient.GetTenancy(ctx, tenantReq)
	if err != nil {
		return nil, err
	}

	tenantInfo := &models.TenantInfo{
		ID:       *tenantResp.Id,
		UserList: []models.TenantUserInfo{},
	}
	if tenantResp.Name != nil {
		tenantInfo.Name = *tenantResp.Name
	}
	if tenantResp.Description != nil {
		tenantInfo.Description = *tenantResp.Description
	}
	if tenantResp.HomeRegionKey != nil {
		tenantInfo.HomeRegionKey = *tenantResp.HomeRegionKey
	}

	// 获取区域列表
	regionsReq := identity.ListRegionSubscriptionsRequest{TenancyId: &user.OciTenantID}
	regionsResp, err := identityClient.ListRegionSubscriptions(ctx, regionsReq)
	if err == nil {
		for _, region := range regionsResp.Items {
			if region.RegionName != nil {
				tenantInfo.Regions = append(tenantInfo.Regions, *region.RegionName)
			}
		}
	}

	// 获取用户列表
	usersReq := identity.ListUsersRequest{CompartmentId: &user.OciTenantID}
	usersResp, err := identityClient.ListUsers(ctx, usersReq)
	if err == nil {
		for _, u := range usersResp.Items {
			userInfo := models.TenantUserInfo{
				ID:    *u.Id,
				State: string(u.LifecycleState),
			}
			if u.Name != nil {
				userInfo.Name = *u.Name
			}
			if u.Email != nil {
				userInfo.Email = *u.Email
			}
			if u.EmailVerified != nil {
				userInfo.EmailVerified = *u.EmailVerified
			}
			if u.IsMfaActivated != nil {
				userInfo.IsMfaActivated = *u.IsMfaActivated
			}
			if u.TimeCreated != nil {
				userInfo.CreateTime = u.TimeCreated.Format("2006-01-02 15:04:05")
			}
			if u.LastSuccessfulLoginTime != nil {
				userInfo.LastSuccessfulLoginTime = u.LastSuccessfulLoginTime.Format("2006-01-02 15:04:05")
			}
			tenantInfo.UserList = append(tenantInfo.UserList, userInfo)
		}
	}

	// 获取密码过期策略
	passwordExpiresAfter, err := s.GetPasswordExpiresAfter(ctx, user)
	if err != nil {
		// 如果获取失败，设置为0（表示永不过期）
		tenantInfo.PasswordExpiresAfter = 0
	} else {
		tenantInfo.PasswordExpiresAfter = passwordExpiresAfter
	}

	return tenantInfo, nil
}

// GetDomainURL 获取 Identity Domain URL
func (s *OCIService) GetDomainURL(ctx context.Context, user *models.OciUser) (string, error) {
	identityClient, err := s.GetIdentityClient(user)
	if err != nil {
		return "", err
	}

	// 列出所有 domains
	listDomainsReq := identity.ListDomainsRequest{
		CompartmentId: &user.OciTenantID,
	}
	listDomainsResp, err := identityClient.ListDomains(ctx, listDomainsReq)
	if err != nil {
		return "", fmt.Errorf("failed to list domains: %w", err)
	}

	// 找到第一个 ACTIVE 状态的 domain
	for _, domain := range listDomainsResp.Items {
		if domain.LifecycleState == identity.DomainLifecycleStateActive && domain.Url != nil {
			return *domain.Url, nil
		}
	}

	return "", fmt.Errorf("no active domain found")
}

// GetPasswordExpiresAfter 获取密码过期天数
func (s *OCIService) GetPasswordExpiresAfter(ctx context.Context, user *models.OciUser) (int, error) {
	// 获取 Domain URL
	domainURL, err := s.GetDomainURL(ctx, user)
	if err != nil {
		return 0, err
	}

	// 创建 Identity Domains Client
	domainsClient, err := s.GetIdentityDomainsClient(user, domainURL)
	if err != nil {
		return 0, err
	}

	// 列出密码策略
	listPoliciesReq := identitydomains.ListPasswordPoliciesRequest{}
	listPoliciesResp, err := domainsClient.ListPasswordPolicies(ctx, listPoliciesReq)
	if err != nil {
		return 0, fmt.Errorf("failed to list password policies: %w", err)
	}

	// 查找 Custom 类型的策略
	if listPoliciesResp.PasswordPolicies.Resources != nil {
		for _, policy := range listPoliciesResp.PasswordPolicies.Resources {
			// 检查是否为 Custom 策略
			if policy.Name != nil && strings.Contains(strings.ToLower(*policy.Name), "custom") {
				if policy.PasswordExpiresAfter != nil {
					return *policy.PasswordExpiresAfter, nil
				}
			}
		}
	}

	// 如果没有找到 Custom 策略，返回0
	return 0, nil
}

// UpdatePasswordExpiresAfter 更新密码过期天数
func (s *OCIService) UpdatePasswordExpiresAfter(ctx context.Context, user *models.OciUser, expiresAfter int) error {
	// 获取 Domain URL
	domainURL, err := s.GetDomainURL(ctx, user)
	if err != nil {
		return err
	}

	// 创建 Identity Domains Client
	domainsClient, err := s.GetIdentityDomainsClient(user, domainURL)
	if err != nil {
		return err
	}

	// 列出密码策略
	listPoliciesReq := identitydomains.ListPasswordPoliciesRequest{}
	listPoliciesResp, err := domainsClient.ListPasswordPolicies(ctx, listPoliciesReq)
	if err != nil {
		return fmt.Errorf("failed to list password policies: %w", err)
	}

	// 查找并更新 Custom 策略
	if listPoliciesResp.PasswordPolicies.Resources == nil {
		return fmt.Errorf("no password policies found")
	}

	for _, policy := range listPoliciesResp.PasswordPolicies.Resources {
		// 检查是否为 Custom 策略
		if policy.Name != nil && strings.Contains(strings.ToLower(*policy.Name), "custom") {
			if policy.Id == nil {
				continue
			}

			// 获取当前策略的完整信息
			getPolicyReq := identitydomains.GetPasswordPolicyRequest{
				PasswordPolicyId: policy.Id,
			}
			getPolicyResp, err := domainsClient.GetPasswordPolicy(ctx, getPolicyReq)
			if err != nil {
				return fmt.Errorf("failed to get password policy: %w", err)
			}

			currentPolicy := getPolicyResp.PasswordPolicy

			// 只更新 PasswordExpiresAfter 字段，保留其他字段
			currentPolicy.PasswordExpiresAfter = &expiresAfter
			currentPolicy.ForcePasswordReset = boolPtr(false)

			// 构建更新请求
			putPolicyReq := identitydomains.PutPasswordPolicyRequest{
				PasswordPolicyId: policy.Id,
				PasswordPolicy:   currentPolicy,
			}

			_, err = domainsClient.PutPasswordPolicy(ctx, putPolicyReq)
			if err != nil {
				return fmt.Errorf("failed to update password policy: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("no custom password policy found")
}

// DeleteUser 删除用户
func (s *OCIService) DeleteUser(ctx context.Context, user *models.OciUser, userId string) error {
	identityClient, err := s.GetIdentityClient(user)
	if err != nil {
		return err
	}

	deleteUserReq := identity.DeleteUserRequest{
		UserId: &userId,
	}

	_, err = identityClient.DeleteUser(ctx, deleteUserReq)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetTrafficData 获取流量统计数据
func (s *OCIService) GetTrafficData(ctx context.Context, user *models.OciUser, vnicId string, startTime string, endTime string) (*models.TrafficData, error) {
	configProvider, err := s.GetConfigProvider(user)
	if err != nil {
		return nil, err
	}

	monitoringClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, err
	}

	trafficData := &models.TrafficData{
		Time:     []string{},
		Inbound:  []string{},
		Outbound: []string{},
	}

	// 构建查询 - 入站流量
	inboundQuery := fmt.Sprintf("NetworksBytesIn[1m]{resourceId = \"%s\"}.mean()", vnicId)
	outboundQuery := fmt.Sprintf("NetworksBytesOut[1m]{resourceId = \"%s\"}.mean()", vnicId)

	compartmentId := user.OciTenantID

	// 获取入站数据
	inReq := monitoring.SummarizeMetricsDataRequest{
		CompartmentId: &compartmentId,
		SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
			Namespace: stringPtr("oci_computeagent"),
			Query:     &inboundQuery,
			StartTime: &common.SDKTime{Time: parseTime(startTime)},
			EndTime:   &common.SDKTime{Time: parseTime(endTime)},
		},
	}

	inResp, err := monitoringClient.SummarizeMetricsData(ctx, inReq)
	if err == nil {
		for _, item := range inResp.Items {
			for _, dp := range item.AggregatedDatapoints {
				if dp.Timestamp != nil {
					trafficData.Time = append(trafficData.Time, dp.Timestamp.Format("15:04"))
				}
				if dp.Value != nil {
					trafficData.Inbound = append(trafficData.Inbound, fmt.Sprintf("%.2f", *dp.Value/1024/1024))
				}
			}
		}
	}

	// 获取出站数据
	outReq := monitoring.SummarizeMetricsDataRequest{
		CompartmentId: &compartmentId,
		SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
			Namespace: stringPtr("oci_computeagent"),
			Query:     &outboundQuery,
			StartTime: &common.SDKTime{Time: parseTime(startTime)},
			EndTime:   &common.SDKTime{Time: parseTime(endTime)},
		},
	}

	outResp, err := monitoringClient.SummarizeMetricsData(ctx, outReq)
	if err == nil {
		for _, item := range outResp.Items {
			for _, dp := range item.AggregatedDatapoints {
				if dp.Value != nil {
					trafficData.Outbound = append(trafficData.Outbound, fmt.Sprintf("%.2f", *dp.Value/1024/1024))
				}
			}
		}
	}

	return trafficData, nil
}

func parseTime(timeStr string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return time.Now().Add(-1 * time.Hour)
	}
	return t
}

// UpdateUserInfo 更新用户信息
func (s *OCIService) UpdateUserInfo(ctx context.Context, user *models.OciUser, userId string, email string, dbUserName string, description string) error {
	domainURL, err := s.GetDomainURL(ctx, user)
	if err != nil {
		return err
	}

	domainsClient, err := s.GetIdentityDomainsClient(user, domainURL)
	if err != nil {
		return err
	}

	operations := []identitydomains.Operations{}

	if email != "" {
		emailVal := interface{}(email)
		operations = append(operations, identitydomains.Operations{
			Op:    identitydomains.OperationsOpReplace,
			Path:  stringPtr("emails[primary eq true].value"),
			Value: &emailVal,
		})
	}

	if dbUserName != "" {
		userNameVal := interface{}(dbUserName)
		operations = append(operations, identitydomains.Operations{
			Op:    identitydomains.OperationsOpReplace,
			Path:  stringPtr("userName"),
			Value: &userNameVal,
		})
	}

	if description != "" {
		descVal := interface{}(description)
		operations = append(operations, identitydomains.Operations{
			Op:    identitydomains.OperationsOpReplace,
			Path:  stringPtr("urn:ietf:params:scim:schemas:oracle:idcs:extension:user:User:description"),
			Value: &descVal,
		})
	}

	if len(operations) == 0 {
		return fmt.Errorf("no fields to update")
	}

	patchReq := identitydomains.PatchUserRequest{
		UserId: &userId,
		PatchOp: identitydomains.PatchOp{
			Schemas:    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
			Operations: operations,
		},
	}

	_, err = domainsClient.PatchUser(ctx, patchReq)
	if err != nil {
		return fmt.Errorf("failed to update user info: %w", err)
	}

	return nil
}

// ResetUserPassword 重置用户密码
func (s *OCIService) ResetUserPassword(ctx context.Context, user *models.OciUser, userId string) error {
	identityClient, err := s.GetIdentityClient(user)
	if err != nil {
		return err
	}

	createPasswordReq := identity.CreateOrResetUIPasswordRequest{
		UserId: &userId,
	}

	_, err = identityClient.CreateOrResetUIPassword(ctx, createPasswordReq)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

// DeleteUserMfaDevices 删除用户的所有MFA设备
func (s *OCIService) DeleteUserMfaDevices(ctx context.Context, user *models.OciUser, userId string) error {
	domainURL, err := s.GetDomainURL(ctx, user)
	if err != nil {
		return err
	}

	domainsClient, err := s.GetIdentityDomainsClient(user, domainURL)
	if err != nil {
		return err
	}

	listMfaReq := identitydomains.ListMyDevicesRequest{
		Filter: stringPtr(fmt.Sprintf("user.value eq \"%s\"", userId)),
	}

	listMfaResp, err := domainsClient.ListMyDevices(ctx, listMfaReq)
	if err != nil {
		return fmt.Errorf("failed to list MFA devices: %w", err)
	}

	if listMfaResp.MyDevices.Resources == nil || len(listMfaResp.MyDevices.Resources) == 0 {
		return nil
	}

	for _, device := range listMfaResp.MyDevices.Resources {
		if device.Id != nil {
			deleteReq := identitydomains.DeleteMyDeviceRequest{
				MyDeviceId: device.Id,
			}
			_, err := domainsClient.DeleteMyDevice(ctx, deleteReq)
			if err != nil {
				return fmt.Errorf("failed to delete MFA device %s: %w", *device.Id, err)
			}
		}
	}

	return nil
}

// DeleteUserApiKeys 删除用户的所有API密钥
func (s *OCIService) DeleteUserApiKeys(ctx context.Context, user *models.OciUser, userId string) error {
	identityClient, err := s.GetIdentityClient(user)
	if err != nil {
		return err
	}

	listKeysReq := identity.ListApiKeysRequest{
		UserId: &userId,
	}

	listKeysResp, err := identityClient.ListApiKeys(ctx, listKeysReq)
	if err != nil {
		return fmt.Errorf("failed to list API keys: %w", err)
	}

	if len(listKeysResp.Items) == 0 {
		return nil
	}

	for _, apiKey := range listKeysResp.Items {
		if apiKey.KeyId != nil {
			deleteKeyReq := identity.DeleteApiKeyRequest{
				UserId:      &userId,
				Fingerprint: apiKey.Fingerprint,
			}
			_, err := identityClient.DeleteApiKey(ctx, deleteKeyReq)
			if err != nil {
				return fmt.Errorf("failed to delete API key %s: %w", *apiKey.KeyId, err)
			}
		}
	}

	return nil
}
