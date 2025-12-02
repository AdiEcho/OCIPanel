package services

import (
	"context"
	"fmt"
	"github.com/adiecho/oci-panel/internal/database"
	"github.com/adiecho/oci-panel/internal/models"
	"github.com/oracle/oci-go-sdk/v65/core"
)

type InstanceService struct {
	ociService *OCIService
}

func NewInstanceService(ociService *OCIService) *InstanceService {
	return &InstanceService{ociService: ociService}
}

type InstanceInfo struct {
	ID                 string `json:"id"`
	DisplayName        string `json:"displayName"`
	State              string `json:"state"`
	AvailabilityDomain string `json:"availabilityDomain"`
	Shape              string `json:"shape"`
	TimeCreated        string `json:"timeCreated"`
	PublicIp           string `json:"publicIp"`
	PrivateIp          string `json:"privateIp"`
}

func (s *InstanceService) ListInstances(userId string, compartmentId string) ([]InstanceInfo, error) {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	instances, err := s.ociService.ListInstances(context.Background(), &user, compartmentId)
	if err != nil {
		return nil, err
	}

	var result []InstanceInfo
	for _, inst := range instances {
		info := InstanceInfo{
			ID:                 *inst.Id,
			DisplayName:        *inst.DisplayName,
			State:              string(inst.LifecycleState),
			AvailabilityDomain: *inst.AvailabilityDomain,
			Shape:              *inst.Shape,
			TimeCreated:        inst.TimeCreated.String(),
		}
		result = append(result, info)
	}

	return result, nil
}

func (s *InstanceService) StartInstance(userId string, instanceId string) error {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return s.ociService.InstanceAction(context.Background(), &user, instanceId, "START")
}

func (s *InstanceService) StopInstance(userId string, instanceId string) error {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return s.ociService.InstanceAction(context.Background(), &user, instanceId, "STOP")
}

func (s *InstanceService) RebootInstance(userId string, instanceId string) error {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return s.ociService.InstanceAction(context.Background(), &user, instanceId, "RESET")
}

func (s *InstanceService) TerminateInstance(userId string, instanceId string) error {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return s.ociService.TerminateInstance(context.Background(), &user, instanceId)
}

func (s *InstanceService) UpdateInstanceName(userId string, instanceId string, displayName string) error {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return s.ociService.UpdateInstance(context.Background(), &user, instanceId, displayName)
}

// ChangePublicIP 更改实例公网IP
func (s *InstanceService) ChangePublicIP(userId string, instanceId string) (string, error) {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	ctx := context.Background()

	// 获取实例详情以找到VNIC
	details, err := s.ociService.GetInstanceDetails(ctx, &user, instanceId)
	if err != nil {
		return "", fmt.Errorf("failed to get instance details: %w", err)
	}

	if len(details.VnicList) == 0 {
		return "", fmt.Errorf("no VNIC found for instance")
	}

	// 使用第一个VNIC更改IP
	vnicId := details.VnicList[0].VnicID
	newIP, err := s.ociService.ChangePublicIP(ctx, &user, vnicId)
	if err != nil {
		return "", fmt.Errorf("failed to change public IP: %w", err)
	}

	return newIP, nil
}

// UpdateInstanceConfig 更新实例配置（CPU和内存）
func (s *InstanceService) UpdateInstanceConfig(userId string, instanceId string, ocpus float32, memoryInGBs float32) error {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return s.ociService.UpdateInstanceShape(context.Background(), &user, instanceId, ocpus, memoryInGBs)
}

// UpdateBootVolumeConfig 更新引导卷配置
func (s *InstanceService) UpdateBootVolumeConfig(userId string, instanceId string, sizeInGBs int64, vpusPerGB int64) error {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	ctx := context.Background()

	// 获取实例信息
	instance, err := s.ociService.GetInstance(ctx, &user, instanceId)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// 获取引导卷ID
	computeClient, err := s.ociService.GetComputeClient(&user)
	if err != nil {
		return fmt.Errorf("failed to get compute client: %w", err)
	}

	listAttachReq := core.ListBootVolumeAttachmentsRequest{
		CompartmentId:      instance.CompartmentId,
		InstanceId:         instance.Id,
		AvailabilityDomain: instance.AvailabilityDomain,
	}
	attachResp, err := computeClient.ListBootVolumeAttachments(ctx, listAttachReq)
	if err != nil {
		return fmt.Errorf("failed to list boot volume attachments: %w", err)
	}

	if len(attachResp.Items) == 0 {
		return fmt.Errorf("no boot volume found for instance")
	}

	bootVolumeId := *attachResp.Items[0].BootVolumeId
	return s.ociService.UpdateBootVolume(ctx, &user, bootVolumeId, sizeInGBs, vpusPerGB)
}

// CreateCloudShellConnection 创建Cloud Shell连接
func (s *InstanceService) CreateCloudShellConnection(userId string, instanceId string, publicKey string) (map[string]string, error) {
	var user models.OciUser
	if err := database.GetDB().Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	ctx := context.Background()

	// 创建控制台连接
	connectionId, err := s.ociService.CreateConsoleConnection(ctx, &user, instanceId, publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create console connection: %w", err)
	}

	// 获取连接字符串
	connectionString, err := s.ociService.GetConsoleConnectionString(ctx, &user, connectionId)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	result := map[string]string{
		"connectionId":     connectionId,
		"connectionString": connectionString,
	}

	return result, nil
}
