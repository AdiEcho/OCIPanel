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
	"github.com/oracle/oci-go-sdk/v65/networkloadbalancer"
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
	BootVolumeSizeGBs  int64
	BootVolumeVpuPerGB int64
}

func (s *OCIService) LaunchInstance(ctx context.Context, user *models.OciUser, params LaunchInstanceParams) (*core.Instance, error) {
	client, err := s.GetComputeClient(user)
	if err != nil {
		return nil, err
	}

	// 构建引导卷配置
	sourceDetails := core.InstanceSourceViaImageDetails{
		ImageId: &params.ImageId,
	}
	if params.BootVolumeSizeGBs > 0 {
		sourceDetails.BootVolumeSizeInGBs = &params.BootVolumeSizeGBs
	}
	if params.BootVolumeVpuPerGB > 0 {
		sourceDetails.BootVolumeVpusPerGB = &params.BootVolumeVpuPerGB
	}

	req := core.LaunchInstanceRequest{
		LaunchInstanceDetails: core.LaunchInstanceDetails{
			CompartmentId:      &params.CompartmentId,
			AvailabilityDomain: &params.AvailabilityDomain,
			DisplayName:        &params.DisplayName,
			SourceDetails:      &sourceDetails,
			Shape:              &params.Shape,
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

// CreateInstance 自动创建实例（自动获取AD、VCN、子网，可指定镜像ID）
func (s *OCIService) CreateInstance(ctx context.Context, user *models.OciUser, region, architecture, operationSystem string, ocpus, memory float64, disk int, vpusPerGB int64, sshPublicKey string, imageIdParam string) error {
	// 临时切换用户区域
	originalRegion := user.OciRegion
	user.OciRegion = region
	defer func() { user.OciRegion = originalRegion }()

	compartmentId := user.OciTenantID

	// 1. 获取身份客户端
	identityClient, err := s.GetIdentityClient(user)
	if err != nil {
		return fmt.Errorf("获取身份客户端失败: %w", err)
	}

	// 2. 获取可用域列表
	adResp, err := identityClient.ListAvailabilityDomains(ctx, identity.ListAvailabilityDomainsRequest{
		CompartmentId: &compartmentId,
	})
	if err != nil {
		return fmt.Errorf("获取可用域失败: %w", err)
	}
	if len(adResp.Items) == 0 {
		return fmt.Errorf("没有可用的可用域")
	}
	availabilityDomain := *adResp.Items[0].Name

	// 3. 获取或创建VCN和子网
	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return fmt.Errorf("获取网络客户端失败: %w", err)
	}

	// 列出现有VCN
	vcnResp, err := vnClient.ListVcns(ctx, core.ListVcnsRequest{
		CompartmentId: &compartmentId,
	})
	if err != nil {
		return fmt.Errorf("获取VCN列表失败: %w", err)
	}

	var subnetId string
	if len(vcnResp.Items) > 0 {
		// 使用现有VCN的子网
		vcn := vcnResp.Items[0]
		subnetResp, err := vnClient.ListSubnets(ctx, core.ListSubnetsRequest{
			CompartmentId: &compartmentId,
			VcnId:         vcn.Id,
		})
		if err != nil {
			return fmt.Errorf("获取子网列表失败: %w", err)
		}
		if len(subnetResp.Items) > 0 {
			subnetId = *subnetResp.Items[0].Id
		}
	}

	if subnetId == "" {
		return fmt.Errorf("没有可用的子网，请先创建VCN和子网")
	}

	// 4. 确定Shape
	shape := "VM.Standard.A1.Flex"
	if architecture == "AMD" {
		shape = "VM.Standard.E2.1.Micro"
	}

	// 5. 获取镜像
	var imageId string
	if imageIdParam != "" {
		// 使用指定的镜像ID
		imageId = imageIdParam
	} else {
		// 自动获取最新镜像
		computeClient, err := s.GetComputeClient(user)
		if err != nil {
			return fmt.Errorf("获取计算客户端失败: %w", err)
		}

		osName := "Canonical Ubuntu"
		if operationSystem == "CentOS" {
			osName = "CentOS"
		} else if operationSystem == "Oracle Linux" {
			osName = "Oracle Linux"
		}

		imageResp, err := computeClient.ListImages(ctx, core.ListImagesRequest{
			CompartmentId:   &compartmentId,
			OperatingSystem: &osName,
			Shape:           &shape,
			SortBy:          core.ListImagesSortByTimecreated,
			SortOrder:       core.ListImagesSortOrderDesc,
		})
		if err != nil {
			return fmt.Errorf("获取镜像列表失败: %w", err)
		}
		if len(imageResp.Items) == 0 {
			return fmt.Errorf("没有找到合适的镜像")
		}
		imageId = *imageResp.Items[0].Id
	}

	// 6. 生成实例名称
	displayName := fmt.Sprintf("instance-%s-%d", architecture, time.Now().Unix())

	// 7. 创建实例
	params := LaunchInstanceParams{
		CompartmentId:      compartmentId,
		AvailabilityDomain: availabilityDomain,
		DisplayName:        displayName,
		ImageId:            imageId,
		Shape:              shape,
		SubnetId:           subnetId,
		Ocpus:              float32(ocpus),
		MemoryInGBs:        float32(memory),
		SshPublicKey:       sshPublicKey,
		BootVolumeSizeGBs:  int64(disk),
		BootVolumeVpuPerGB: vpusPerGB,
	}

	_, err = s.LaunchInstance(ctx, user, params)
	if err != nil {
		return fmt.Errorf("创建实例失败: %w", err)
	}

	return nil
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

// ImageInfo 镜像信息
type ImageInfo struct {
	ID                     string `json:"id"`
	DisplayName            string `json:"displayName"`
	OperatingSystem        string `json:"operatingSystem"`
	OperatingSystemVersion string `json:"operatingSystemVersion"`
	SizeInMBs              int64  `json:"sizeInMBs"`
	TimeCreated            string `json:"timeCreated"`
}

// ListImages 获取可用镜像列表
func (s *OCIService) ListImages(ctx context.Context, user *models.OciUser, region, architecture string) ([]ImageInfo, error) {
	// 临时切换用户区域
	originalRegion := user.OciRegion
	user.OciRegion = region
	defer func() { user.OciRegion = originalRegion }()

	computeClient, err := s.GetComputeClient(user)
	if err != nil {
		return nil, fmt.Errorf("获取计算客户端失败: %w", err)
	}

	compartmentId := user.OciTenantID

	// 确定Shape
	shape := "VM.Standard.A1.Flex"
	if architecture == "AMD" {
		shape = "VM.Standard.E2.1.Micro"
	}

	req := core.ListImagesRequest{
		CompartmentId: &compartmentId,
		Shape:         &shape,
		SortBy:        core.ListImagesSortByTimecreated,
		SortOrder:     core.ListImagesSortOrderDesc,
	}

	resp, err := computeClient.ListImages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %w", err)
	}

	var images []ImageInfo
	for _, img := range resp.Items {
		info := ImageInfo{
			ID:                     *img.Id,
			DisplayName:            *img.DisplayName,
			OperatingSystem:        *img.OperatingSystem,
			OperatingSystemVersion: *img.OperatingSystemVersion,
		}
		if img.SizeInMBs != nil {
			info.SizeInMBs = *img.SizeInMBs
		}
		if img.TimeCreated != nil {
			info.TimeCreated = img.TimeCreated.Format("2006-01-02 15:04:05")
		}
		images = append(images, info)
	}

	return images, nil
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

	// 获取租户创建时间（通过compartment的创建时间）
	compartmentReq := identity.GetCompartmentRequest{CompartmentId: &user.OciTenantID}
	compartmentResp, err := identityClient.GetCompartment(ctx, compartmentReq)
	if err == nil && compartmentResp.TimeCreated != nil {
		tenantInfo.CreateTime = compartmentResp.TimeCreated.Format("2006-01-02 15:04:05")
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

	// 查找 Custom 类型的策略（通过 PasswordStrength 字段判断）
	if listPoliciesResp.PasswordPolicies.Resources != nil {
		for _, policy := range listPoliciesResp.PasswordPolicies.Resources {
			// 检查是否为 Custom 策略（使用 PasswordStrength 字段）
			if policy.PasswordStrength == identitydomains.PasswordPolicyPasswordStrengthCustom {
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
		// 检查是否为 Custom 策略（使用 PasswordStrength 字段）
		if policy.PasswordStrength == identitydomains.PasswordPolicyPasswordStrengthCustom {
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

// CreateIpv6ByInstanceId 通过实例ID创建并附加IPv6地址
func (s *OCIService) CreateIpv6ByInstanceId(ctx context.Context, user *models.OciUser, instanceId string) (string, error) {
	computeClient, err := s.GetComputeClient(user)
	if err != nil {
		return "", err
	}

	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return "", err
	}

	// 获取实例信息
	instance, err := s.GetInstance(ctx, user, instanceId)
	if err != nil {
		return "", fmt.Errorf("failed to get instance: %w", err)
	}

	// 获取VNIC附件
	listVnicReq := core.ListVnicAttachmentsRequest{
		CompartmentId: instance.CompartmentId,
		InstanceId:    instance.Id,
	}
	vnicResp, err := computeClient.ListVnicAttachments(ctx, listVnicReq)
	if err != nil {
		return "", fmt.Errorf("failed to list VNIC attachments: %w", err)
	}

	if len(vnicResp.Items) == 0 {
		return "", fmt.Errorf("no VNIC found for instance")
	}

	// 获取第一个VNIC
	vnicId := vnicResp.Items[0].VnicId
	if vnicId == nil {
		return "", fmt.Errorf("VNIC ID is nil")
	}

	// 获取VNIC详情
	vnicReq := core.GetVnicRequest{VnicId: vnicId}
	vnic, err := vnClient.GetVnic(ctx, vnicReq)
	if err != nil {
		return "", fmt.Errorf("failed to get VNIC: %w", err)
	}

	// 检查是否已有IPv6
	ipv6ListReq := core.ListIpv6sRequest{VnicId: vnicId}
	ipv6ListResp, err := vnClient.ListIpv6s(ctx, ipv6ListReq)
	if err == nil && len(ipv6ListResp.Items) > 0 {
		if ipv6ListResp.Items[0].IpAddress != nil {
			return "", fmt.Errorf("instance already has IPv6 address: %s", *ipv6ListResp.Items[0].IpAddress)
		}
	}

	// 获取子网信息
	subnetReq := core.GetSubnetRequest{SubnetId: vnic.SubnetId}
	subnetResp, err := vnClient.GetSubnet(ctx, subnetReq)
	if err != nil {
		return "", fmt.Errorf("failed to get subnet: %w", err)
	}

	// 获取VCN信息
	vcnReq := core.GetVcnRequest{VcnId: subnetResp.VcnId}
	vcnResp, err := vnClient.GetVcn(ctx, vcnReq)
	if err != nil {
		return "", fmt.Errorf("failed to get VCN: %w", err)
	}

	// 检查VCN是否启用了IPv6
	if len(vcnResp.Ipv6CidrBlocks) == 0 {
		return "", fmt.Errorf("VCN does not have IPv6 enabled. Please enable IPv6 on VCN first via Oracle Cloud Console")
	}

	ipv6SubnetCidr := vcnResp.Ipv6CidrBlocks[0]

	// 创建IPv6
	createIpv6Req := core.CreateIpv6Request{
		CreateIpv6Details: core.CreateIpv6Details{
			VnicId:         vnicId,
			Ipv6SubnetCidr: &ipv6SubnetCidr,
		},
	}
	createIpv6Resp, err := vnClient.CreateIpv6(ctx, createIpv6Req)
	if err != nil {
		return "", fmt.Errorf("failed to create IPv6: %w", err)
	}

	if createIpv6Resp.IpAddress == nil {
		return "", fmt.Errorf("IPv6 address is nil")
	}

	return *createIpv6Resp.IpAddress, nil
}

// GetInstanceById 根据实例ID获取实例
func (s *OCIService) GetInstanceById(user *models.OciUser, instanceID string) (*core.Instance, error) {
	ctx := context.Background()
	client, err := s.GetComputeClient(user)
	if err != nil {
		return nil, err
	}

	resp, err := client.GetInstance(ctx, core.GetInstanceRequest{InstanceId: &instanceID})
	if err != nil {
		return nil, err
	}

	return &resp.Instance, nil
}

// GetBootVolumeByInstanceId 根据实例ID获取引导卷
func (s *OCIService) GetBootVolumeByInstanceId(user *models.OciUser, instanceID string) (*core.BootVolume, error) {
	ctx := context.Background()

	computeClient, err := s.GetComputeClient(user)
	if err != nil {
		return nil, err
	}

	blockClient, err := s.GetBlockstorageClient(user)
	if err != nil {
		return nil, err
	}

	// 获取实例
	instance, err := computeClient.GetInstance(ctx, core.GetInstanceRequest{InstanceId: &instanceID})
	if err != nil {
		return nil, err
	}

	// 获取引导卷附件
	bvaResp, err := computeClient.ListBootVolumeAttachments(ctx, core.ListBootVolumeAttachmentsRequest{
		CompartmentId:      instance.CompartmentId,
		AvailabilityDomain: instance.AvailabilityDomain,
		InstanceId:         &instanceID,
	})
	if err != nil {
		return nil, err
	}

	if len(bvaResp.Items) == 0 {
		return nil, fmt.Errorf("no boot volume attachment found")
	}

	// 获取引导卷
	bvResp, err := blockClient.GetBootVolume(ctx, core.GetBootVolumeRequest{
		BootVolumeId: bvaResp.Items[0].BootVolumeId,
	})
	if err != nil {
		return nil, err
	}

	return &bvResp.BootVolume, nil
}

// GetNetworkLoadBalancerClient 获取网络负载均衡器客户端
func (s *OCIService) GetNetworkLoadBalancerClient(user *models.OciUser) (networkloadbalancer.NetworkLoadBalancerClient, error) {
	configProvider, err := s.GetConfigProvider(user)
	if err != nil {
		return networkloadbalancer.NetworkLoadBalancerClient{}, err
	}

	client, err := networkloadbalancer.NewNetworkLoadBalancerClientWithConfigurationProvider(configProvider)
	if err != nil {
		return networkloadbalancer.NetworkLoadBalancerClient{}, err
	}

	return client, nil
}

// AutoRescueParams 自动救援参数
type AutoRescueParams struct {
	InstanceID       string
	InstanceName     string
	KeepBackupVolume bool
}

// AutoRescueProgress 自动救援进度
type AutoRescueProgress struct {
	Step        int    `json:"step"`
	TotalSteps  int    `json:"totalSteps"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	PublicIP    string `json:"publicIp,omitempty"`
	SSHPassword string `json:"sshPassword,omitempty"`
}

// AutoRescue 自动救援/缩小硬盘 (9步骤)
func (s *OCIService) AutoRescue(user *models.OciUser, params AutoRescueParams, progressChan chan<- AutoRescueProgress) error {
	ctx := context.Background()

	computeClient, err := s.GetComputeClient(user)
	if err != nil {
		return fmt.Errorf("failed to get compute client: %w", err)
	}

	blockClient, err := s.GetBlockstorageClient(user)
	if err != nil {
		return fmt.Errorf("failed to get blockstorage client: %w", err)
	}

	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return fmt.Errorf("failed to get virtual network client: %w", err)
	}

	sendProgress := func(step int, status, message string) {
		if progressChan != nil {
			progressChan <- AutoRescueProgress{
				Step:       step,
				TotalSteps: 9,
				Status:     status,
				Message:    message,
			}
		}
	}

	// 获取实例信息
	instance, err := s.GetInstanceById(user, params.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// 获取引导卷
	bootVolume, err := s.GetBootVolumeByInstanceId(user, params.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to get boot volume: %w", err)
	}

	// Step 1: 关机
	sendProgress(1, "running", "正在关机...")
	_, err = computeClient.InstanceAction(ctx, core.InstanceActionRequest{
		InstanceId: instance.Id,
		Action:     core.InstanceActionActionStop,
	})
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	// 等待实例停止
	for {
		instResp, err := computeClient.GetInstance(ctx, core.GetInstanceRequest{InstanceId: instance.Id})
		if err != nil {
			return fmt.Errorf("failed to get instance status: %w", err)
		}
		if instResp.LifecycleState == core.InstanceLifecycleStateStopped {
			break
		}
		time.Sleep(2 * time.Second)
	}
	sendProgress(1, "completed", "关机成功")

	// 等待引导卷可用
	for {
		bvResp, err := blockClient.GetBootVolume(ctx, core.GetBootVolumeRequest{BootVolumeId: bootVolume.Id})
		if err != nil {
			return fmt.Errorf("failed to get boot volume status: %w", err)
		}
		if bvResp.LifecycleState == core.BootVolumeLifecycleStateAvailable {
			break
		}
		time.Sleep(2 * time.Second)
	}

	// Step 2: 备份原引导卷
	sendProgress(2, "running", "正在备份原引导卷...")
	backupName := "Old-BootVolume-Backup"
	backupResp, err := blockClient.CreateBootVolumeBackup(ctx, core.CreateBootVolumeBackupRequest{
		CreateBootVolumeBackupDetails: core.CreateBootVolumeBackupDetails{
			BootVolumeId: bootVolume.Id,
			DisplayName:  &backupName,
			Type:         core.CreateBootVolumeBackupDetailsTypeFull,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create boot volume backup: %w", err)
	}
	backupId := backupResp.Id
	sendProgress(2, "completed", "备份原引导卷成功")

	time.Sleep(3 * time.Second)

	// Step 3: 分离原引导卷
	sendProgress(3, "running", "正在分离原引导卷...")
	// 获取引导卷附件
	bvaListResp, err := computeClient.ListBootVolumeAttachments(ctx, core.ListBootVolumeAttachmentsRequest{
		CompartmentId:      instance.CompartmentId,
		AvailabilityDomain: instance.AvailabilityDomain,
		InstanceId:         instance.Id,
	})
	if err != nil {
		return fmt.Errorf("failed to list boot volume attachments: %w", err)
	}
	if len(bvaListResp.Items) == 0 {
		return fmt.Errorf("no boot volume attachment found")
	}
	bvaId := bvaListResp.Items[0].Id

	_, err = computeClient.DetachBootVolume(ctx, core.DetachBootVolumeRequest{
		BootVolumeAttachmentId: bvaId,
	})
	if err != nil {
		return fmt.Errorf("failed to detach boot volume: %w", err)
	}
	sendProgress(3, "completed", "分离原引导卷成功")

	// 等待备份完成
	for {
		backupStatusResp, err := blockClient.GetBootVolumeBackup(ctx, core.GetBootVolumeBackupRequest{BootVolumeBackupId: backupId})
		if err != nil {
			return fmt.Errorf("failed to get backup status: %w", err)
		}
		if backupStatusResp.LifecycleState == core.BootVolumeBackupLifecycleStateAvailable {
			break
		}
		time.Sleep(2 * time.Second)
	}

	// Step 4: 删除原引导卷
	sendProgress(4, "running", "正在删除原引导卷...")
	_, err = blockClient.DeleteBootVolume(ctx, core.DeleteBootVolumeRequest{
		BootVolumeId: bootVolume.Id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete boot volume: %w", err)
	}

	// 等待引导卷删除完成
	for {
		bvResp, err := blockClient.GetBootVolume(ctx, core.GetBootVolumeRequest{BootVolumeId: bootVolume.Id})
		if err != nil {
			break // 404 means deleted
		}
		if bvResp.LifecycleState == core.BootVolumeLifecycleStateTerminated {
			break
		}
		time.Sleep(2 * time.Second)
	}
	sendProgress(4, "completed", "删除原引导卷成功")

	// Step 5: 从备份创建新的47GB引导卷
	sendProgress(5, "running", "正在创建47GB引导卷...")
	newBvName := "Restored-Boot-Volume-47GB"
	sizeInGBs := int64(47)
	newBvResp, err := blockClient.CreateBootVolume(ctx, core.CreateBootVolumeRequest{
		CreateBootVolumeDetails: core.CreateBootVolumeDetails{
			CompartmentId:      instance.CompartmentId,
			AvailabilityDomain: instance.AvailabilityDomain,
			DisplayName:        &newBvName,
			SizeInGBs:          &sizeInGBs,
			SourceDetails: core.BootVolumeSourceFromBootVolumeBackupDetails{
				Id: backupId,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create new boot volume: %w", err)
	}
	newBvId := newBvResp.Id

	// 等待新引导卷可用
	for {
		bvResp, err := blockClient.GetBootVolume(ctx, core.GetBootVolumeRequest{BootVolumeId: newBvId})
		if err != nil {
			return fmt.Errorf("failed to get new boot volume status: %w", err)
		}
		if bvResp.LifecycleState == core.BootVolumeLifecycleStateAvailable {
			break
		}
		time.Sleep(2 * time.Second)
	}
	sendProgress(5, "completed", "创建47GB引导卷成功")

	// Step 6: 附加新引导卷到实例
	sendProgress(6, "running", "正在附加新引导卷到实例...")
	attachName := "New-Boot-Volume"
	_, err = computeClient.AttachBootVolume(ctx, core.AttachBootVolumeRequest{
		AttachBootVolumeDetails: core.AttachBootVolumeDetails{
			BootVolumeId: newBvId,
			InstanceId:   instance.Id,
			DisplayName:  &attachName,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to attach boot volume: %w", err)
	}
	sendProgress(6, "completed", "附加新引导卷成功")

	// Step 7: 删除备份（如果不保留）
	if !params.KeepBackupVolume {
		sendProgress(7, "running", "正在删除原引导卷备份...")
		_, err = blockClient.DeleteBootVolumeBackup(ctx, core.DeleteBootVolumeBackupRequest{
			BootVolumeBackupId: backupId,
		})
		if err != nil {
			// 不影响后续操作，只记录错误
			sendProgress(7, "warning", "删除原引导卷备份失败，但不影响继续操作")
		} else {
			sendProgress(7, "completed", "删除原引导卷备份成功")
		}
	} else {
		sendProgress(7, "skipped", "保留原引导卷备份")
	}

	// Step 8: 等待引导卷附件完成
	sendProgress(8, "running", "正在等待引导卷附件完成...")
	time.Sleep(5 * time.Second)
	sendProgress(8, "completed", "引导卷附件完成")

	// Step 9: 启动实例
	sendProgress(9, "running", "正在启动实例...")
	for i := 0; i < 30; i++ {
		instResp, err := computeClient.GetInstance(ctx, core.GetInstanceRequest{InstanceId: instance.Id})
		if err != nil {
			continue
		}
		if instResp.LifecycleState == core.InstanceLifecycleStateRunning {
			break
		}
		_, _ = computeClient.InstanceAction(ctx, core.InstanceActionRequest{
			InstanceId: instance.Id,
			Action:     core.InstanceActionActionStart,
		})
		time.Sleep(3 * time.Second)
	}

	// 获取公网IP
	publicIP := ""
	vnicAttachments, err := computeClient.ListVnicAttachments(ctx, core.ListVnicAttachmentsRequest{
		CompartmentId: instance.CompartmentId,
		InstanceId:    instance.Id,
	})
	if err == nil && len(vnicAttachments.Items) > 0 {
		vnicId := vnicAttachments.Items[0].VnicId
		if vnicId != nil {
			vnicResp, err := vnClient.GetVnic(ctx, core.GetVnicRequest{VnicId: vnicId})
			if err == nil && vnicResp.PublicIp != nil {
				publicIP = *vnicResp.PublicIp
			}
		}
	}

	if progressChan != nil {
		progressChan <- AutoRescueProgress{
			Step:       9,
			TotalSteps: 9,
			Status:     "completed",
			Message:    "实例救援成功，已启动",
			PublicIP:   publicIP,
		}
	}

	return nil
}

// Enable500Mbps 一键开启下行500Mbps
func (s *OCIService) Enable500Mbps(user *models.OciUser, instanceID string, sshPort int) (string, error) {
	ctx := context.Background()

	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return "", fmt.Errorf("failed to get virtual network client: %w", err)
	}

	nlbClient, err := s.GetNetworkLoadBalancerClient(user)
	if err != nil {
		return "", fmt.Errorf("failed to get network load balancer client: %w", err)
	}

	// 获取实例信息
	instance, err := s.GetInstanceById(user, instanceID)
	if err != nil {
		return "", fmt.Errorf("failed to get instance: %w", err)
	}

	// 检查是否为AMD实例
	if !strings.Contains(*instance.Shape, "E2.1.Micro") {
		return "", fmt.Errorf("only AMD E2.1.Micro instances support 500Mbps")
	}

	// 获取VCN
	vcn, err := s.GetVcnByInstanceId(user, instanceID)
	if err != nil {
		return "", fmt.Errorf("failed to get VCN: %w", err)
	}

	// 获取VNIC
	vnic, err := s.GetVnicByInstanceId(user, instanceID)
	if err != nil {
		return "", fmt.Errorf("failed to get VNIC: %w", err)
	}

	// 获取私有IP
	privateIpResp, err := vnClient.ListPrivateIps(ctx, core.ListPrivateIpsRequest{VnicId: vnic.Id})
	if err != nil || len(privateIpResp.Items) == 0 {
		return "", fmt.Errorf("failed to get private IP: %w", err)
	}
	privateIP := *privateIpResp.Items[0].IpAddress

	compartmentID := *instance.CompartmentId

	// 创建或获取NAT网关
	natGatewayResp, err := vnClient.ListNatGateways(ctx, core.ListNatGatewaysRequest{
		CompartmentId:  &compartmentID,
		VcnId:          vcn.Id,
		LifecycleState: core.NatGatewayLifecycleStateAvailable,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list NAT gateways: %w", err)
	}

	var natGatewayId *string
	if len(natGatewayResp.Items) > 0 {
		natGatewayId = natGatewayResp.Items[0].Id
	} else {
		// 创建NAT网关
		natName := "nat-gateway"
		createNatResp, err := vnClient.CreateNatGateway(ctx, core.CreateNatGatewayRequest{
			CreateNatGatewayDetails: core.CreateNatGatewayDetails{
				CompartmentId: &compartmentID,
				VcnId:         vcn.Id,
				DisplayName:   &natName,
			},
		})
		if err != nil {
			return "", fmt.Errorf("failed to create NAT gateway: %w", err)
		}
		natGatewayId = createNatResp.Id

		// 等待NAT网关可用
		for {
			natResp, err := vnClient.GetNatGateway(ctx, core.GetNatGatewayRequest{NatGatewayId: natGatewayId})
			if err != nil {
				return "", fmt.Errorf("failed to get NAT gateway status: %w", err)
			}
			if natResp.LifecycleState == core.NatGatewayLifecycleStateAvailable {
				break
			}
			time.Sleep(2 * time.Second)
		}
	}

	// 获取子网
	subnetResp, err := vnClient.ListSubnets(ctx, core.ListSubnetsRequest{
		CompartmentId: &compartmentID,
		VcnId:         vcn.Id,
	})
	if err != nil || len(subnetResp.Items) == 0 {
		return "", fmt.Errorf("failed to list subnets: %w", err)
	}
	subnetId := subnetResp.Items[0].Id

	// 删除现有网络负载均衡器
	existingNlbResp, err := nlbClient.ListNetworkLoadBalancers(ctx, networkloadbalancer.ListNetworkLoadBalancersRequest{
		CompartmentId:  &compartmentID,
		LifecycleState: networkloadbalancer.ListNetworkLoadBalancersLifecycleStateActive,
	})
	if err == nil && existingNlbResp.NetworkLoadBalancerCollection.Items != nil {
		for _, nlb := range existingNlbResp.NetworkLoadBalancerCollection.Items {
			_, _ = nlbClient.DeleteNetworkLoadBalancer(ctx, networkloadbalancer.DeleteNetworkLoadBalancerRequest{
				NetworkLoadBalancerId: nlb.Id,
			})
		}
		time.Sleep(5 * time.Second)
	}

	// 创建网络负载均衡器
	nlbName := fmt.Sprintf("nlb-%s", time.Now().Format("20060102150405"))
	isPrivate := false
	port := 0
	weight := 1
	isPreserveSource := true
	isFailOpen := true

	createNlbResp, err := nlbClient.CreateNetworkLoadBalancer(ctx, networkloadbalancer.CreateNetworkLoadBalancerRequest{
		CreateNetworkLoadBalancerDetails: networkloadbalancer.CreateNetworkLoadBalancerDetails{
			CompartmentId: &compartmentID,
			DisplayName:   &nlbName,
			SubnetId:      subnetId,
			IsPrivate:     &isPrivate,
			Listeners: map[string]networkloadbalancer.ListenerDetails{
				"listener1": {
					Name:                  stringPtr("listener1"),
					DefaultBackendSetName: stringPtr("backend1"),
					Protocol:              networkloadbalancer.ListenerProtocolsTcpAndUdp,
					Port:                  &port,
				},
			},
			BackendSets: map[string]networkloadbalancer.BackendSetDetails{
				"backend1": {
					Policy:           networkloadbalancer.NetworkLoadBalancingPolicyTwoTuple,
					IsPreserveSource: &isPreserveSource,
					IsFailOpen:       &isFailOpen,
					HealthChecker: &networkloadbalancer.HealthChecker{
						Protocol: networkloadbalancer.HealthCheckProtocolsTcp,
						Port:     &sshPort,
					},
					Backends: []networkloadbalancer.Backend{
						{
							IpAddress: &privateIP,
							TargetId:  instance.Id,
							Port:      &port,
							Weight:    &weight,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create network load balancer: %w", err)
	}

	nlbId := createNlbResp.Id

	// 等待NLB可用
	for {
		nlbResp, err := nlbClient.GetNetworkLoadBalancer(ctx, networkloadbalancer.GetNetworkLoadBalancerRequest{
			NetworkLoadBalancerId: nlbId,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get NLB status: %w", err)
		}
		if nlbResp.LifecycleState == networkloadbalancer.LifecycleStateActive {
			break
		}
		time.Sleep(3 * time.Second)
	}

	// 获取NLB公网IP
	nlbResp, err := nlbClient.GetNetworkLoadBalancer(ctx, networkloadbalancer.GetNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: nlbId,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get NLB: %w", err)
	}

	var publicIP string
	for _, ip := range nlbResp.IpAddresses {
		if ip.IpAddress != nil && !isPrivateIP(*ip.IpAddress) {
			publicIP = *ip.IpAddress
			break
		}
	}

	// 创建或更新NAT路由表
	routeTableResp, err := vnClient.ListRouteTables(ctx, core.ListRouteTablesRequest{
		CompartmentId:  &compartmentID,
		VcnId:          vcn.Id,
		LifecycleState: core.RouteTableLifecycleStateAvailable,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list route tables: %w", err)
	}

	var natRouteTableId *string
	for _, rt := range routeTableResp.Items {
		for _, rule := range rt.RouteRules {
			if rule.NetworkEntityId != nil && *rule.NetworkEntityId == *natGatewayId &&
				rule.Destination != nil && *rule.Destination == "0.0.0.0/0" {
				natRouteTableId = rt.Id
				break
			}
		}
		if natRouteTableId != nil {
			break
		}
	}

	if natRouteTableId == nil {
		// 创建新路由表
		rtName := "nat-route"
		destination := "0.0.0.0/0"
		createRtResp, err := vnClient.CreateRouteTable(ctx, core.CreateRouteTableRequest{
			CreateRouteTableDetails: core.CreateRouteTableDetails{
				CompartmentId: &compartmentID,
				VcnId:         vcn.Id,
				DisplayName:   &rtName,
				RouteRules: []core.RouteRule{
					{
						Destination:     &destination,
						NetworkEntityId: natGatewayId,
						DestinationType: core.RouteRuleDestinationTypeCidrBlock,
					},
				},
			},
		})
		if err != nil {
			return "", fmt.Errorf("failed to create route table: %w", err)
		}
		natRouteTableId = createRtResp.Id

		// 等待路由表可用
		for {
			rtResp, err := vnClient.GetRouteTable(ctx, core.GetRouteTableRequest{RtId: natRouteTableId})
			if err != nil {
				return "", fmt.Errorf("failed to get route table status: %w", err)
			}
			if rtResp.LifecycleState == core.RouteTableLifecycleStateAvailable {
				break
			}
			time.Sleep(2 * time.Second)
		}
	}

	// 更新VNIC绑定路由表并跳过源/目的地检查
	skipSourceDestCheck := true
	_, err = vnClient.UpdateVnic(ctx, core.UpdateVnicRequest{
		VnicId: vnic.Id,
		UpdateVnicDetails: core.UpdateVnicDetails{
			SkipSourceDestCheck: &skipSourceDestCheck,
			RouteTableId:        natRouteTableId,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to update VNIC: %w", err)
	}

	// 放行安全规则
	_ = s.ReleaseSecurityRules(user, *vcn.Id)

	return publicIP, nil
}

// Disable500Mbps 关闭下行500Mbps
func (s *OCIService) Disable500Mbps(user *models.OciUser, instanceID string, retainNatGw, retainNlb bool) error {
	ctx := context.Background()

	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return fmt.Errorf("failed to get virtual network client: %w", err)
	}

	nlbClient, err := s.GetNetworkLoadBalancerClient(user)
	if err != nil {
		return fmt.Errorf("failed to get network load balancer client: %w", err)
	}

	// 获取实例信息
	instance, err := s.GetInstanceById(user, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// 检查是否为AMD实例
	if !strings.Contains(*instance.Shape, "E2.1.Micro") {
		return fmt.Errorf("only AMD E2.1.Micro instances support this operation")
	}

	// 获取VCN
	vcn, err := s.GetVcnByInstanceId(user, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get VCN: %w", err)
	}

	// 获取VNIC
	vnic, err := s.GetVnicByInstanceId(user, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get VNIC: %w", err)
	}

	compartmentID := *instance.CompartmentId

	// 获取所有路由表
	routeTableResp, err := vnClient.ListRouteTables(ctx, core.ListRouteTablesRequest{
		CompartmentId:  &compartmentID,
		VcnId:          vcn.Id,
		LifecycleState: core.RouteTableLifecycleStateAvailable,
	})
	if err != nil {
		return fmt.Errorf("failed to list route tables: %w", err)
	}

	// 查找默认路由表（不含NAT规则的）
	var defaultRouteTableId *string
	var natRouteTableIds []*string
	for _, rt := range routeTableResp.Items {
		hasNatRule := false
		for _, rule := range rt.RouteRules {
			if rule.Destination != nil && *rule.Destination == "0.0.0.0/0" {
				// 检查是否指向NAT网关
				natGwResp, _ := vnClient.ListNatGateways(ctx, core.ListNatGatewaysRequest{
					CompartmentId:  &compartmentID,
					VcnId:          vcn.Id,
					LifecycleState: core.NatGatewayLifecycleStateAvailable,
				})
				for _, natGw := range natGwResp.Items {
					if rule.NetworkEntityId != nil && *rule.NetworkEntityId == *natGw.Id {
						hasNatRule = true
						break
					}
				}
			}
		}
		if hasNatRule {
			natRouteTableIds = append(natRouteTableIds, rt.Id)
		} else if defaultRouteTableId == nil {
			defaultRouteTableId = rt.Id
		}
	}

	if defaultRouteTableId == nil && len(routeTableResp.Items) > 0 {
		defaultRouteTableId = routeTableResp.Items[0].Id
	}

	// 更新VNIC绑定到默认路由表
	if defaultRouteTableId != nil {
		skipSourceDestCheck := true
		_, err = vnClient.UpdateVnic(ctx, core.UpdateVnicRequest{
			VnicId: vnic.Id,
			UpdateVnicDetails: core.UpdateVnicDetails{
				SkipSourceDestCheck: &skipSourceDestCheck,
				RouteTableId:        defaultRouteTableId,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to update VNIC: %w", err)
		}
	}

	// 删除NAT路由表
	if !retainNatGw {
		for _, rtId := range natRouteTableIds {
			// 先清空路由规则
			_, _ = vnClient.UpdateRouteTable(ctx, core.UpdateRouteTableRequest{
				RtId: rtId,
				UpdateRouteTableDetails: core.UpdateRouteTableDetails{
					RouteRules: []core.RouteRule{},
				},
			})
			time.Sleep(2 * time.Second)
			// 删除路由表
			_, _ = vnClient.DeleteRouteTable(ctx, core.DeleteRouteTableRequest{RtId: rtId})
		}

		// 删除NAT网关
		natGwResp, _ := vnClient.ListNatGateways(ctx, core.ListNatGatewaysRequest{
			CompartmentId:  &compartmentID,
			VcnId:          vcn.Id,
			LifecycleState: core.NatGatewayLifecycleStateAvailable,
		})
		for _, natGw := range natGwResp.Items {
			_, _ = vnClient.DeleteNatGateway(ctx, core.DeleteNatGatewayRequest{NatGatewayId: natGw.Id})
		}
	}

	// 删除网络负载均衡器
	if !retainNlb {
		nlbResp, _ := nlbClient.ListNetworkLoadBalancers(ctx, networkloadbalancer.ListNetworkLoadBalancersRequest{
			CompartmentId: &compartmentID,
		})
		if nlbResp.NetworkLoadBalancerCollection.Items != nil {
			for _, nlb := range nlbResp.NetworkLoadBalancerCollection.Items {
				_, _ = nlbClient.DeleteNetworkLoadBalancer(ctx, networkloadbalancer.DeleteNetworkLoadBalancerRequest{
					NetworkLoadBalancerId: nlb.Id,
				})
			}
		}
	}

	return nil
}

// ReleaseSecurityRules 放行安全规则
func (s *OCIService) ReleaseSecurityRules(user *models.OciUser, vcnId string) error {
	ctx := context.Background()

	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return fmt.Errorf("failed to get virtual network client: %w", err)
	}

	// 获取VCN
	vcnResp, err := vnClient.GetVcn(ctx, core.GetVcnRequest{VcnId: &vcnId})
	if err != nil {
		return fmt.Errorf("failed to get VCN: %w", err)
	}

	// 获取安全列表
	secListResp, err := vnClient.ListSecurityLists(ctx, core.ListSecurityListsRequest{
		CompartmentId: vcnResp.CompartmentId,
		VcnId:         &vcnId,
	})
	if err != nil {
		return fmt.Errorf("failed to list security lists: %w", err)
	}

	allProtocol := "all"
	ipv4Cidr := "0.0.0.0/0"
	ipv6Cidr := "::/0"
	internalCidr := "10.0.0.0/16"

	for _, secList := range secListResp.Items {
		// 更新入站规则
		ingressRules := []core.IngressSecurityRule{
			{
				Protocol: &allProtocol,
				Source:   &ipv4Cidr,
			},
			{
				Protocol: &allProtocol,
				Source:   &ipv6Cidr,
			},
			{
				Protocol: &allProtocol,
				Source:   &internalCidr,
			},
		}

		// 更新出站规则
		egressRules := []core.EgressSecurityRule{
			{
				Protocol:    &allProtocol,
				Destination: &ipv4Cidr,
			},
			{
				Protocol:    &allProtocol,
				Destination: &ipv6Cidr,
			},
		}

		_, err = vnClient.UpdateSecurityList(ctx, core.UpdateSecurityListRequest{
			SecurityListId: secList.Id,
			UpdateSecurityListDetails: core.UpdateSecurityListDetails{
				IngressSecurityRules: ingressRules,
				EgressSecurityRules:  egressRules,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to update security list: %w", err)
		}
	}

	return nil
}

// GetVcnByInstanceId 根据实例ID获取VCN
func (s *OCIService) GetVcnByInstanceId(user *models.OciUser, instanceID string) (*core.Vcn, error) {
	ctx := context.Background()

	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual network client: %w", err)
	}

	vnic, err := s.GetVnicByInstanceId(user, instanceID)
	if err != nil {
		return nil, err
	}

	// 获取子网
	subnetResp, err := vnClient.GetSubnet(ctx, core.GetSubnetRequest{SubnetId: vnic.SubnetId})
	if err != nil {
		return nil, fmt.Errorf("failed to get subnet: %w", err)
	}

	// 获取VCN
	vcnResp, err := vnClient.GetVcn(ctx, core.GetVcnRequest{VcnId: subnetResp.VcnId})
	if err != nil {
		return nil, fmt.Errorf("failed to get VCN: %w", err)
	}

	return &vcnResp.Vcn, nil
}

// GetVnicByInstanceId 根据实例ID获取VNIC
func (s *OCIService) GetVnicByInstanceId(user *models.OciUser, instanceID string) (*core.Vnic, error) {
	ctx := context.Background()

	computeClient, err := s.GetComputeClient(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get compute client: %w", err)
	}

	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual network client: %w", err)
	}

	// 获取实例
	instResp, err := computeClient.GetInstance(ctx, core.GetInstanceRequest{InstanceId: &instanceID})
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// 获取VNIC附件
	vnicAttachResp, err := computeClient.ListVnicAttachments(ctx, core.ListVnicAttachmentsRequest{
		CompartmentId: instResp.CompartmentId,
		InstanceId:    &instanceID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list VNIC attachments: %w", err)
	}

	if len(vnicAttachResp.Items) == 0 {
		return nil, fmt.Errorf("no VNIC attachment found")
	}

	// 获取VNIC
	vnicResp, err := vnClient.GetVnic(ctx, core.GetVnicRequest{VnicId: vnicAttachResp.Items[0].VnicId})
	if err != nil {
		return nil, fmt.Errorf("failed to get VNIC: %w", err)
	}

	return &vnicResp.Vnic, nil
}

// isPrivateIP 检查是否为私有IP
func isPrivateIP(ip string) bool {
	return strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "172.16.") ||
		strings.HasPrefix(ip, "172.17.") ||
		strings.HasPrefix(ip, "172.18.") ||
		strings.HasPrefix(ip, "172.19.") ||
		strings.HasPrefix(ip, "172.20.") ||
		strings.HasPrefix(ip, "172.21.") ||
		strings.HasPrefix(ip, "172.22.") ||
		strings.HasPrefix(ip, "172.23.") ||
		strings.HasPrefix(ip, "172.24.") ||
		strings.HasPrefix(ip, "172.25.") ||
		strings.HasPrefix(ip, "172.26.") ||
		strings.HasPrefix(ip, "172.27.") ||
		strings.HasPrefix(ip, "172.28.") ||
		strings.HasPrefix(ip, "172.29.") ||
		strings.HasPrefix(ip, "172.30.") ||
		strings.HasPrefix(ip, "172.31.") ||
		strings.HasPrefix(ip, "192.168.")
}

// GetSecurityListByVcnId 获取VCN的安全列表详情
func (s *OCIService) GetSecurityListByVcnId(ctx context.Context, user *models.OciUser, vcnId string) (*models.SecurityListInfo, error) {
	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual network client: %w", err)
	}

	// 获取VCN
	vcnResp, err := vnClient.GetVcn(ctx, core.GetVcnRequest{VcnId: &vcnId})
	if err != nil {
		return nil, fmt.Errorf("failed to get VCN: %w", err)
	}

	// 获取默认安全列表
	secListResp, err := vnClient.GetSecurityList(ctx, core.GetSecurityListRequest{
		SecurityListId: vcnResp.DefaultSecurityListId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get security list: %w", err)
	}

	secList := secListResp.SecurityList
	result := &models.SecurityListInfo{
		ID:           *secList.Id,
		DisplayName:  *secList.DisplayName,
		VcnId:        vcnId,
		IngressRules: []models.SecurityRule{},
		EgressRules:  []models.SecurityRule{},
	}

	// 解析入站规则
	for _, rule := range secList.IngressSecurityRules {
		sr := models.SecurityRule{
			IsStateless: false,
			Protocol:    *rule.Protocol,
			Source:      *rule.Source,
		}
		if rule.IsStateless != nil {
			sr.IsStateless = *rule.IsStateless
		}
		if rule.Description != nil {
			sr.Description = *rule.Description
		}
		sr.ProtocolName = getProtocolName(*rule.Protocol)
		// TCP端口
		if *rule.Protocol == "6" && rule.TcpOptions != nil {
			if rule.TcpOptions.DestinationPortRange != nil {
				sr.PortRangeMin = *rule.TcpOptions.DestinationPortRange.Min
				sr.PortRangeMax = *rule.TcpOptions.DestinationPortRange.Max
			}
		}
		// UDP端口
		if *rule.Protocol == "17" && rule.UdpOptions != nil {
			if rule.UdpOptions.DestinationPortRange != nil {
				sr.PortRangeMin = *rule.UdpOptions.DestinationPortRange.Min
				sr.PortRangeMax = *rule.UdpOptions.DestinationPortRange.Max
			}
		}
		// ICMP
		if *rule.Protocol == "1" && rule.IcmpOptions != nil {
			sr.IcmpType = rule.IcmpOptions.Type
			sr.IcmpCode = rule.IcmpOptions.Code
		}
		result.IngressRules = append(result.IngressRules, sr)
	}

	// 解析出站规则
	for _, rule := range secList.EgressSecurityRules {
		sr := models.SecurityRule{
			IsStateless: false,
			Protocol:    *rule.Protocol,
			Destination: *rule.Destination,
		}
		if rule.IsStateless != nil {
			sr.IsStateless = *rule.IsStateless
		}
		if rule.Description != nil {
			sr.Description = *rule.Description
		}
		sr.ProtocolName = getProtocolName(*rule.Protocol)
		// TCP端口
		if *rule.Protocol == "6" && rule.TcpOptions != nil {
			if rule.TcpOptions.DestinationPortRange != nil {
				sr.PortRangeMin = *rule.TcpOptions.DestinationPortRange.Min
				sr.PortRangeMax = *rule.TcpOptions.DestinationPortRange.Max
			}
		}
		// UDP端口
		if *rule.Protocol == "17" && rule.UdpOptions != nil {
			if rule.UdpOptions.DestinationPortRange != nil {
				sr.PortRangeMin = *rule.UdpOptions.DestinationPortRange.Min
				sr.PortRangeMax = *rule.UdpOptions.DestinationPortRange.Max
			}
		}
		// ICMP
		if *rule.Protocol == "1" && rule.IcmpOptions != nil {
			sr.IcmpType = rule.IcmpOptions.Type
			sr.IcmpCode = rule.IcmpOptions.Code
		}
		result.EgressRules = append(result.EgressRules, sr)
	}

	return result, nil
}

// getProtocolName 获取协议名称
func getProtocolName(protocol string) string {
	switch protocol {
	case "all":
		return "所有协议"
	case "1":
		return "ICMP"
	case "6":
		return "TCP"
	case "17":
		return "UDP"
	case "58":
		return "ICMPv6"
	default:
		return protocol
	}
}

// AddSecurityRule 添加安全规则
func (s *OCIService) AddSecurityRule(ctx context.Context, user *models.OciUser, vcnId string, rule *models.SecurityRule, isIngress bool) error {
	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return fmt.Errorf("failed to get virtual network client: %w", err)
	}

	// 获取VCN
	vcnResp, err := vnClient.GetVcn(ctx, core.GetVcnRequest{VcnId: &vcnId})
	if err != nil {
		return fmt.Errorf("failed to get VCN: %w", err)
	}

	// 获取当前安全列表
	secListResp, err := vnClient.GetSecurityList(ctx, core.GetSecurityListRequest{
		SecurityListId: vcnResp.DefaultSecurityListId,
	})
	if err != nil {
		return fmt.Errorf("failed to get security list: %w", err)
	}

	ingressRules := secListResp.IngressSecurityRules
	egressRules := secListResp.EgressSecurityRules

	if isIngress {
		newRule := core.IngressSecurityRule{
			Protocol:    &rule.Protocol,
			Source:      &rule.Source,
			IsStateless: &rule.IsStateless,
		}
		if rule.Description != "" {
			newRule.Description = &rule.Description
		}
		// TCP
		if rule.Protocol == "6" && (rule.PortRangeMin > 0 || rule.PortRangeMax > 0) {
			newRule.TcpOptions = &core.TcpOptions{
				DestinationPortRange: &core.PortRange{
					Min: &rule.PortRangeMin,
					Max: &rule.PortRangeMax,
				},
			}
		}
		// UDP
		if rule.Protocol == "17" && (rule.PortRangeMin > 0 || rule.PortRangeMax > 0) {
			newRule.UdpOptions = &core.UdpOptions{
				DestinationPortRange: &core.PortRange{
					Min: &rule.PortRangeMin,
					Max: &rule.PortRangeMax,
				},
			}
		}
		// ICMP
		if rule.Protocol == "1" && rule.IcmpType != nil {
			newRule.IcmpOptions = &core.IcmpOptions{
				Type: rule.IcmpType,
				Code: rule.IcmpCode,
			}
		}
		ingressRules = append(ingressRules, newRule)
	} else {
		newRule := core.EgressSecurityRule{
			Protocol:    &rule.Protocol,
			Destination: &rule.Destination,
			IsStateless: &rule.IsStateless,
		}
		if rule.Description != "" {
			newRule.Description = &rule.Description
		}
		// TCP
		if rule.Protocol == "6" && (rule.PortRangeMin > 0 || rule.PortRangeMax > 0) {
			newRule.TcpOptions = &core.TcpOptions{
				DestinationPortRange: &core.PortRange{
					Min: &rule.PortRangeMin,
					Max: &rule.PortRangeMax,
				},
			}
		}
		// UDP
		if rule.Protocol == "17" && (rule.PortRangeMin > 0 || rule.PortRangeMax > 0) {
			newRule.UdpOptions = &core.UdpOptions{
				DestinationPortRange: &core.PortRange{
					Min: &rule.PortRangeMin,
					Max: &rule.PortRangeMax,
				},
			}
		}
		// ICMP
		if rule.Protocol == "1" && rule.IcmpType != nil {
			newRule.IcmpOptions = &core.IcmpOptions{
				Type: rule.IcmpType,
				Code: rule.IcmpCode,
			}
		}
		egressRules = append(egressRules, newRule)
	}

	// 更新安全列表
	_, err = vnClient.UpdateSecurityList(ctx, core.UpdateSecurityListRequest{
		SecurityListId: vcnResp.DefaultSecurityListId,
		UpdateSecurityListDetails: core.UpdateSecurityListDetails{
			IngressSecurityRules: ingressRules,
			EgressSecurityRules:  egressRules,
		},
	})
	return err
}

// DeleteVcn 删除VCN及其相关资源
func (s *OCIService) DeleteVcn(ctx context.Context, user *models.OciUser, vcnId string) error {
	vnClient, err := s.GetVirtualNetworkClient(user)
	if err != nil {
		return fmt.Errorf("failed to get virtual network client: %w", err)
	}

	// 获取VCN
	vcnResp, err := vnClient.GetVcn(ctx, core.GetVcnRequest{VcnId: &vcnId})
	if err != nil {
		return fmt.Errorf("failed to get VCN: %w", err)
	}
	vcn := vcnResp.Vcn

	// 1. 清空路由表规则
	if vcn.DefaultRouteTableId != nil {
		_, err = vnClient.UpdateRouteTable(ctx, core.UpdateRouteTableRequest{
			RtId: vcn.DefaultRouteTableId,
			UpdateRouteTableDetails: core.UpdateRouteTableDetails{
				RouteRules: []core.RouteRule{},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to clear route table: %w", err)
		}
	}

	// 2. 删除所有子网
	subnetsResp, err := vnClient.ListSubnets(ctx, core.ListSubnetsRequest{
		CompartmentId: vcn.CompartmentId,
		VcnId:         &vcnId,
	})
	if err == nil {
		for _, subnet := range subnetsResp.Items {
			_, err = vnClient.DeleteSubnet(ctx, core.DeleteSubnetRequest{
				SubnetId: subnet.Id,
			})
			if err != nil {
				return fmt.Errorf("failed to delete subnet %s: %w", *subnet.DisplayName, err)
			}
			// 等待子网删除完成
			time.Sleep(2 * time.Second)
		}
	}

	// 3. 删除Internet网关
	igwsResp, err := vnClient.ListInternetGateways(ctx, core.ListInternetGatewaysRequest{
		CompartmentId: vcn.CompartmentId,
		VcnId:         &vcnId,
	})
	if err == nil {
		for _, igw := range igwsResp.Items {
			_, err = vnClient.DeleteInternetGateway(ctx, core.DeleteInternetGatewayRequest{
				IgId: igw.Id,
			})
			if err != nil {
				return fmt.Errorf("failed to delete internet gateway: %w", err)
			}
		}
	}

	// 4. 删除NAT网关
	natGwsResp, err := vnClient.ListNatGateways(ctx, core.ListNatGatewaysRequest{
		CompartmentId: vcn.CompartmentId,
		VcnId:         &vcnId,
	})
	if err == nil {
		for _, natGw := range natGwsResp.Items {
			_, err = vnClient.DeleteNatGateway(ctx, core.DeleteNatGatewayRequest{
				NatGatewayId: natGw.Id,
			})
			if err != nil {
				return fmt.Errorf("failed to delete NAT gateway: %w", err)
			}
		}
	}

	// 5. 删除服务网关
	sgwsResp, err := vnClient.ListServiceGateways(ctx, core.ListServiceGatewaysRequest{
		CompartmentId: vcn.CompartmentId,
		VcnId:         &vcnId,
	})
	if err == nil {
		for _, sgw := range sgwsResp.Items {
			_, err = vnClient.DeleteServiceGateway(ctx, core.DeleteServiceGatewayRequest{
				ServiceGatewayId: sgw.Id,
			})
			if err != nil {
				return fmt.Errorf("failed to delete service gateway: %w", err)
			}
		}
	}

	// 6. 删除网络安全组
	nsgsResp, err := vnClient.ListNetworkSecurityGroups(ctx, core.ListNetworkSecurityGroupsRequest{
		CompartmentId: vcn.CompartmentId,
		VcnId:         &vcnId,
	})
	if err == nil {
		for _, nsg := range nsgsResp.Items {
			// 先清空安全规则
			_, _ = vnClient.UpdateNetworkSecurityGroupSecurityRules(ctx, core.UpdateNetworkSecurityGroupSecurityRulesRequest{
				NetworkSecurityGroupId: nsg.Id,
				UpdateNetworkSecurityGroupSecurityRulesDetails: core.UpdateNetworkSecurityGroupSecurityRulesDetails{
					SecurityRules: []core.UpdateSecurityRuleDetails{},
				},
			})
			_, err = vnClient.DeleteNetworkSecurityGroup(ctx, core.DeleteNetworkSecurityGroupRequest{
				NetworkSecurityGroupId: nsg.Id,
			})
			if err != nil {
				return fmt.Errorf("failed to delete network security group: %w", err)
			}
		}
	}

	// 7. 删除VCN
	_, err = vnClient.DeleteVcn(ctx, core.DeleteVcnRequest{
		VcnId: &vcnId,
	})
	if err != nil {
		return fmt.Errorf("failed to delete VCN: %w", err)
	}

	return nil
}

// MonthlyTrafficStats 月度流量统计结果
type MonthlyTrafficStats struct {
	InstanceCount   int
	InboundTraffic  int64
	OutboundTraffic int64
}

// GetMonthlyTrafficStats 获取指定配置的月度流量统计
func (s *OCIService) GetMonthlyTrafficStats(ctx context.Context, user *models.OciUser) (*MonthlyTrafficStats, error) {
	configProvider, err := s.GetConfigProvider(user)
	if err != nil {
		return nil, err
	}

	computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, err
	}

	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, err
	}

	monitoringClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, err
	}

	compartmentId := user.OciTenantID
	stats := &MonthlyTrafficStats{}

	// 获取实例列表
	instances, err := s.ListInstances(ctx, user, compartmentId)
	if err != nil {
		return nil, err
	}
	stats.InstanceCount = len(instances)

	// 获取本月时间范围
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	// 遍历每个实例获取VNIC流量
	for _, instance := range instances {
		if instance.Id == nil {
			continue
		}

		// 获取实例的VNIC附件
		vnicAttachReq := core.ListVnicAttachmentsRequest{
			CompartmentId: &compartmentId,
			InstanceId:    instance.Id,
		}
		vnicAttachResp, err := computeClient.ListVnicAttachments(ctx, vnicAttachReq)
		if err != nil {
			continue
		}

		for _, attach := range vnicAttachResp.Items {
			if attach.VnicId == nil {
				continue
			}

			// 获取VNIC信息
			vnicReq := core.GetVnicRequest{VnicId: attach.VnicId}
			vnicResp, err := vnClient.GetVnic(ctx, vnicReq)
			if err != nil {
				continue
			}

			if vnicResp.Id == nil {
				continue
			}

			vnicId := *vnicResp.Id

			// 查询入站流量 (VnicToNetworkBytes)
			inQuery := fmt.Sprintf("VnicToNetworkBytes[1d]{resourceId = \"%s\"}.sum()", vnicId)
			inReq := monitoring.SummarizeMetricsDataRequest{
				CompartmentId: &compartmentId,
				SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
					Namespace: stringPtr("oci_vcn"),
					Query:     &inQuery,
					StartTime: &common.SDKTime{Time: startOfMonth},
					EndTime:   &common.SDKTime{Time: endOfMonth},
				},
			}
			inResp, err := monitoringClient.SummarizeMetricsData(ctx, inReq)
			if err == nil {
				for _, item := range inResp.Items {
					for _, dp := range item.AggregatedDatapoints {
						if dp.Value != nil {
							stats.InboundTraffic += int64(*dp.Value)
						}
					}
				}
			}

			// 查询出站流量 (VnicFromNetworkBytes)
			outQuery := fmt.Sprintf("VnicFromNetworkBytes[1d]{resourceId = \"%s\"}.sum()", vnicId)
			outReq := monitoring.SummarizeMetricsDataRequest{
				CompartmentId: &compartmentId,
				SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
					Namespace: stringPtr("oci_vcn"),
					Query:     &outQuery,
					StartTime: &common.SDKTime{Time: startOfMonth},
					EndTime:   &common.SDKTime{Time: endOfMonth},
				},
			}
			outResp, err := monitoringClient.SummarizeMetricsData(ctx, outReq)
			if err == nil {
				for _, item := range outResp.Items {
					for _, dp := range item.AggregatedDatapoints {
						if dp.Value != nil {
							stats.OutboundTraffic += int64(*dp.Value)
						}
					}
				}
			}
		}
	}

	return stats, nil
}

// FormatBytes 格式化字节数为人类可读格式
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
