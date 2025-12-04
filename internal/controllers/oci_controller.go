package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/adiecho/oci-panel/internal/database"
	"github.com/adiecho/oci-panel/internal/models"
	"github.com/adiecho/oci-panel/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/oracle/oci-go-sdk/v65/core"
)

type OciController struct {
	ociService       *services.OCIService
	schedulerService *services.SchedulerService
}

func NewOciController(ociService *services.OCIService, schedulerService *services.SchedulerService) *OciController {
	return &OciController{
		ociService:       ociService,
		schedulerService: schedulerService,
	}
}

type UserPageRequest struct {
	Page     int    `json:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" binding:"required,min=1,max=100"`
	Username string `json:"username"`
}

type UserPageResponse struct {
	List     []models.OciUserListResponse `json:"list"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"pageSize"`
}

func (oc *OciController) UserPage(c *gin.Context) {
	var req UserPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var users []models.OciUser
	var total int64

	query := db.Model(&models.OciUser{})
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}

	query.Count(&total)
	offset := (req.Page - 1) * req.PageSize
	query.Order("create_time DESC").Limit(req.PageSize).Offset(offset).Find(&users)

	responseList := make([]models.OciUserListResponse, len(users))
	cacheEnabled := oc.schedulerService.IsCacheEnabled()

	if cacheEnabled {
		// 从数据库缓存读取
		for i, user := range users {
			tenantCreateTime := ""
			if user.TenantCreateTime != nil {
				tenantCreateTime = user.TenantCreateTime.Format("2006-01-02 15:04:05")
			}
			responseList[i] = models.OciUserListResponse{
				ID:               user.ID,
				Username:         user.Username,
				TenantName:       user.TenantName,
				TenantCreateTime: tenantCreateTime,
				OciTenantID:      user.OciTenantID,
				OciRegion:        user.OciRegion,
				CreateTime:       user.CreateTime.Format("2006-01-02 15:04:05"),
				InstanceCount:    0,
				RunningInstances: 0,
			}

			cache, err := oc.schedulerService.GetConfigCache(user.ID)
			if err == nil {
				responseList[i].InstanceCount = cache.InstanceCount
				responseList[i].RunningInstances = cache.RunningInstances
			}
		}
	} else {
		// 实时获取（并发）
		type instanceCountResult struct {
			index            int
			instanceCount    int
			runningInstances int
		}
		resultChan := make(chan instanceCountResult, len(users))

		ctx := context.Background()
		for i, user := range users {
			go func(idx int, u models.OciUser) {
				result := instanceCountResult{index: idx}
				instances, err := oc.ociService.ListInstances(ctx, &u, u.OciTenantID)
				if err == nil {
					result.instanceCount = len(instances)
					for _, inst := range instances {
						if inst.LifecycleState == core.InstanceLifecycleStateRunning {
							result.runningInstances++
						}
					}
				}
				resultChan <- result
			}(i, user)
		}

		for i, user := range users {
			tenantCreateTime := ""
			if user.TenantCreateTime != nil {
				tenantCreateTime = user.TenantCreateTime.Format("2006-01-02 15:04:05")
			}
			responseList[i] = models.OciUserListResponse{
				ID:               user.ID,
				Username:         user.Username,
				TenantName:       user.TenantName,
				TenantCreateTime: tenantCreateTime,
				OciTenantID:      user.OciTenantID,
				OciRegion:        user.OciRegion,
				CreateTime:       user.CreateTime.Format("2006-01-02 15:04:05"),
				InstanceCount:    0,
				RunningInstances: 0,
			}
		}

		for i := 0; i < len(users); i++ {
			result := <-resultChan
			responseList[result.index].InstanceCount = result.instanceCount
			responseList[result.index].RunningInstances = result.runningInstances
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(UserPageResponse{
		List:     responseList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "success"))
}

type AddCfgRequest struct {
	Username       string `json:"username" binding:"required"`
	TenantName     string `json:"tenantName" binding:"required"`
	OciTenantID    string `json:"ociTenantId" binding:"required"`
	OciUserID      string `json:"ociUserId" binding:"required"`
	OciFingerprint string `json:"ociFingerprint" binding:"required"`
	OciRegion      string `json:"ociRegion" binding:"required"`
	OciKeyPath     string `json:"ociKeyPath" binding:"required"`
}

func (oc *OciController) AddCfg(c *gin.Context) {
	var req AddCfgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	user := models.OciUser{
		ID:             uuid.New().String(),
		Username:       req.Username,
		TenantName:     req.TenantName,
		OciTenantID:    req.OciTenantID,
		OciUserID:      req.OciUserID,
		OciFingerprint: req.OciFingerprint,
		OciRegion:      req.OciRegion,
		OciKeyPath:     req.OciKeyPath,
		CreateTime:     time.Now(),
	}

	// 获取真正的租户名称和创建时间
	ctx := context.Background()
	tenantInfo, err := oc.ociService.GetTenantInfo(ctx, &user)
	if err == nil && tenantInfo != nil {
		if tenantInfo.Name != "" {
			user.TenantName = tenantInfo.Name
		}
		if tenantInfo.CreateTime != "" {
			if parsedTime, err := time.Parse("2006-01-02 15:04:05", tenantInfo.CreateTime); err == nil {
				user.TenantCreateTime = &parsedTime
			}
		}
	}

	if err := database.GetDB().Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to create user"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "Configuration added successfully"))
}

type UpdateCfgNameRequest struct {
	ID         string `json:"id" binding:"required"`
	Username   string `json:"username" binding:"required"`
	OciKeyPath string `json:"ociKeyPath"`
}

func (oc *OciController) UpdateCfgName(c *gin.Context) {
	var req UpdateCfgNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	updates := map[string]interface{}{
		"username": req.Username,
	}

	if req.OciKeyPath != "" {
		updates["oci_key_path"] = req.OciKeyPath
	}

	if err := database.GetDB().Model(&models.OciUser{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to update"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "Updated successfully"))
}

type RemoveCfgRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

func (oc *OciController) RemoveCfg(c *gin.Context) {
	var req RemoveCfgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	var users []models.OciUser
	if err := database.GetDB().Where("id IN ?", req.IDs).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to query users"))
		return
	}

	for _, user := range users {
		if user.OciKeyPath != "" {
			keyPath := filepath.Join("./keys", user.OciKeyPath)
			os.Remove(keyPath)
		}
	}

	if err := database.GetDB().Where("id IN ?", req.IDs).Delete(&models.OciUser{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to delete"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "Deleted successfully"))
}

type CreateInstanceRequest struct {
	UserID          string  `json:"userId" binding:"required"`
	OciRegion       string  `json:"ociRegion" binding:"required"`
	Ocpus           float64 `json:"ocpus"`
	Memory          float64 `json:"memory"`
	Disk            int     `json:"disk"`
	Architecture    string  `json:"architecture"`
	OperationSystem string  `json:"operationSystem"`
	SSHKeyID        string  `json:"sshKeyId" binding:"required"`
}

func (oc *OciController) CreateInstance(c *gin.Context) {
	var req CreateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	// 验证SSH密钥是否存在
	var sshKey models.SSHKey
	if err := database.GetDB().First(&sshKey, "id = ?", req.SSHKeyID).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "SSH密钥不存在"))
		return
	}

	task := models.OciCreateTask{
		ID:              uuid.New().String(),
		UserID:          req.UserID,
		OciRegion:       req.OciRegion,
		Ocpus:           req.Ocpus,
		Memory:          req.Memory,
		Disk:            req.Disk,
		Architecture:    req.Architecture,
		OperationSystem: req.OperationSystem,
		SSHKeyID:        req.SSHKeyID,
		CreateTime:      time.Now(),
	}

	if err := database.GetDB().Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to create task"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "Task created successfully"))
}

type CreateTaskPageRequest struct {
	Page     int    `json:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" binding:"required,min=1,max=100"`
	UserID   string `json:"userId"`
}

type CreateTaskPageResponse struct {
	List     []models.OciCreateTask `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

func (oc *OciController) CreateTaskPage(c *gin.Context) {
	var req CreateTaskPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var tasks []models.OciCreateTask
	var total int64

	query := db.Model(&models.OciCreateTask{})
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	query.Count(&total)
	offset := (req.Page - 1) * req.PageSize
	query.Order("create_time DESC").Limit(req.PageSize).Offset(offset).Find(&tasks)

	c.JSON(http.StatusOK, models.SuccessResponse(CreateTaskPageResponse{
		List:     tasks,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "success"))
}

func (oc *OciController) UploadKey(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "No file uploaded"))
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".pem" && ext != ".key" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "Only .pem or .key files are allowed"))
		return
	}

	keysDir := "./keys"
	if _, err := os.Stat(keysDir); os.IsNotExist(err) {
		if err := os.MkdirAll(keysDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to create keys directory"))
			return
		}
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(keysDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to save file"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(filename, "File uploaded successfully"))
}

type GetConfigDetailsRequest struct {
	ConfigID string `json:"configId" binding:"required"`
}

func (oc *OciController) GetConfigDetails(c *gin.Context) {
	var req GetConfigDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	details := models.OciConfigDetails{
		UserID:      user.ID,
		Username:    user.Username,
		TenantID:    user.OciTenantID,
		TenantName:  user.TenantName,
		Fingerprint: user.OciFingerprint,
		KeyPath:     filepath.Base(user.OciKeyPath),
		Region:      user.OciRegion,
		CreateTime:  user.CreateTime.Format("2006-01-02 15:04:05"),
		Instances:   []models.InstanceInfo{},
		Volumes:     []models.VolumeInfo{},
		VCNs:        []models.VCNInfo{},
	}

	c.JSON(http.StatusOK, models.SuccessResponse(details, "Success"))
}

type GetResourceRequest struct {
	ConfigID   string `json:"configId" binding:"required"`
	ClearCache bool   `json:"clearCache"`
}

// GetConfigInstances 获取配置的实例列表
func (oc *OciController) GetConfigInstances(c *gin.Context) {
	var req GetResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	// 如果需要刷新缓存，先更新
	if req.ClearCache {
		oc.schedulerService.UpdateConfigCache(req.ConfigID)
	}

	// 尝试从数据库缓存获取
	if oc.schedulerService.IsCacheEnabled() {
		cache, err := oc.schedulerService.GetConfigCache(req.ConfigID)
		if err == nil && cache.InstancesData != "" {
			var instances []models.InstanceInfo
			if json.Unmarshal([]byte(cache.InstancesData), &instances) == nil {
				c.JSON(http.StatusOK, models.SuccessResponse(instances, "Success (cached)"))
				return
			}
		}
	}

	// 实时获取
	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	compartmentId := user.OciTenantID

	instances := []models.InstanceInfo{}
	instanceList, err := oc.ociService.ListInstances(ctx, &user, compartmentId)
	if err == nil {
		for _, inst := range instanceList {
			if inst.Id != nil {
				instanceDetail, err := oc.ociService.GetInstanceDetails(ctx, &user, *inst.Id)
				if err == nil {
					instances = append(instances, *instanceDetail)
				}
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(instances, "Success"))
}

// GetConfigVolumes 获取配置的存储卷列表
func (oc *OciController) GetConfigVolumes(c *gin.Context) {
	var req GetResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	if req.ClearCache {
		oc.schedulerService.UpdateConfigCache(req.ConfigID)
	}

	// 尝试从数据库缓存获取
	if oc.schedulerService.IsCacheEnabled() {
		cache, err := oc.schedulerService.GetConfigCache(req.ConfigID)
		if err == nil && cache.VolumesData != "" {
			var volumes []models.VolumeInfo
			if json.Unmarshal([]byte(cache.VolumesData), &volumes) == nil {
				c.JSON(http.StatusOK, models.SuccessResponse(volumes, "Success (cached)"))
				return
			}
		}
	}

	// 实时获取
	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	compartmentId := user.OciTenantID

	volumes, err := oc.ociService.ListBootVolumes(ctx, &user, compartmentId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(volumes, "Success"))
}

// GetConfigVCNs 获取配置的VCN列表
func (oc *OciController) GetConfigVCNs(c *gin.Context) {
	var req GetResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	if req.ClearCache {
		oc.schedulerService.UpdateConfigCache(req.ConfigID)
	}

	// 尝试从数据库缓存获取
	if oc.schedulerService.IsCacheEnabled() {
		cache, err := oc.schedulerService.GetConfigCache(req.ConfigID)
		if err == nil && cache.VcnsData != "" {
			var vcns []models.VCNInfo
			if json.Unmarshal([]byte(cache.VcnsData), &vcns) == nil {
				c.JSON(http.StatusOK, models.SuccessResponse(vcns, "Success (cached)"))
				return
			}
		}
	}

	// 实时获取
	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	compartmentId := user.OciTenantID

	vcns, err := oc.ociService.ListVCNs(ctx, &user, compartmentId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(vcns, "Success"))
}

// ClearConfigCache 刷新配置的缓存
func (oc *OciController) ClearConfigCache(c *gin.Context) {
	var req GetConfigDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	// 重新获取并更新缓存
	go oc.schedulerService.UpdateConfigCache(req.ConfigID)

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "Cache refresh started"))
}

// GetTenantInfo 获取租户详情
func (oc *OciController) GetTenantInfo(c *gin.Context) {
	var req GetResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	if req.ClearCache {
		oc.schedulerService.UpdateConfigCache(req.ConfigID)
	}

	// 尝试从数据库缓存获取
	if oc.schedulerService.IsCacheEnabled() {
		cache, err := oc.schedulerService.GetConfigCache(req.ConfigID)
		if err == nil && cache.TenantData != "" {
			var tenantInfo models.TenantInfo
			if json.Unmarshal([]byte(cache.TenantData), &tenantInfo) == nil {
				c.JSON(http.StatusOK, models.SuccessResponse(tenantInfo, "Success (cached)"))
				return
			}
		}
	}

	// 实时获取
	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	tenantInfo, err := oc.ociService.GetTenantInfo(ctx, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(tenantInfo, "Success"))
}

type GetTrafficDataRequest struct {
	ConfigID   string `json:"configId" binding:"required"`
	InstanceID string `json:"instanceId" binding:"required"`
	VnicID     string `json:"vnicId" binding:"required"`
	StartTime  string `json:"startTime" binding:"required"`
	EndTime    string `json:"endTime" binding:"required"`
}

// GetTrafficData 获取流量统计数据
func (oc *OciController) GetTrafficData(c *gin.Context) {
	var req GetTrafficDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	trafficData, err := oc.ociService.GetTrafficData(ctx, &user, req.VnicID, req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(trafficData, "Success"))
}

// GetTrafficCondition 获取流量查询条件（区域和实例列表）
func (oc *OciController) GetTrafficCondition(c *gin.Context) {
	configId := c.Query("configId")
	if configId == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "configId is required"))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", configId).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	condition := models.TrafficCondition{
		Regions:   []models.ValueLabel{},
		Instances: []models.ValueLabel{},
	}

	// 获取租户区域
	tenantInfo, err := oc.ociService.GetTenantInfo(ctx, &user)
	if err == nil {
		for _, region := range tenantInfo.Regions {
			condition.Regions = append(condition.Regions, models.ValueLabel{
				Value: region,
				Label: region,
			})
		}
	}

	// 获取当前区域的实例
	compartmentId := user.OciTenantID
	instances, err := oc.ociService.ListInstances(ctx, &user, compartmentId)
	if err == nil {
		for _, inst := range instances {
			if inst.Id != nil && inst.DisplayName != nil {
				condition.Instances = append(condition.Instances, models.ValueLabel{
					Value: *inst.Id,
					Label: *inst.DisplayName,
				})
			}
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(condition, "Success"))
}

type UpdatePasswordExpiryRequest struct {
	CfgID                string `json:"cfgId" binding:"required"`
	PasswordExpiresAfter int    `json:"passwordExpiresAfter"`
}

type UserManagementRequest struct {
	OciCfgID string `json:"ociCfgId" binding:"required"`
	UserID   string `json:"userId" binding:"required"`
}

type UpdateUserInfoRequest struct {
	OciCfgID    string `json:"ociCfgId" binding:"required"`
	UserID      string `json:"userId" binding:"required"`
	Email       string `json:"email" binding:"required"`
	DbUserName  string `json:"dbUserName" binding:"required"`
	Description string `json:"description"`
}

// UpdatePasswordExpiry 更新密码过期策略
func (oc *OciController) UpdatePasswordExpiry(c *gin.Context) {
	var req UpdatePasswordExpiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.CfgID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	err := oc.ociService.UpdatePasswordExpiresAfter(ctx, &user, req.PasswordExpiresAfter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	// 更新缓存
	go oc.schedulerService.UpdateConfigCache(req.CfgID)

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "密码过期策略已更新"))
}

// UpdateUserInfo 更新用户信息
func (oc *OciController) UpdateUserInfo(c *gin.Context) {
	var req UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.OciCfgID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	err := oc.ociService.UpdateUserInfo(ctx, &user, req.UserID, req.Email, req.DbUserName, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	go oc.schedulerService.UpdateConfigCache(req.OciCfgID)

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "用户信息更新成功"))
}

// DeleteUser 删除用户
func (oc *OciController) DeleteUser(c *gin.Context) {
	var req UserManagementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.OciCfgID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	err := oc.ociService.DeleteUser(ctx, &user, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	go oc.schedulerService.UpdateConfigCache(req.OciCfgID)

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "用户删除成功"))
}

// ResetPassword 重置用户密码
func (oc *OciController) ResetPassword(c *gin.Context) {
	var req UserManagementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.OciCfgID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	err := oc.ociService.ResetUserPassword(ctx, &user, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "密码重置成功"))
}

// DeleteMfaDevice 清除用户MFA设备
func (oc *OciController) DeleteMfaDevice(c *gin.Context) {
	var req UserManagementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.OciCfgID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	err := oc.ociService.DeleteUserMfaDevices(ctx, &user, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	go oc.schedulerService.UpdateConfigCache(req.OciCfgID)

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "MFA设备清除成功"))
}

// DeleteApiKey 清除用户API密钥
func (oc *OciController) DeleteApiKey(c *gin.Context) {
	var req UserManagementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.OciCfgID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	err := oc.ociService.DeleteUserApiKeys(ctx, &user, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "API密钥清除成功"))
}

// GetInstanceVnics 获取实例的VNIC列表
func (oc *OciController) GetInstanceVnics(c *gin.Context) {
	configId := c.Query("configId")
	instanceId := c.Query("instanceId")
	if configId == "" || instanceId == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "configId and instanceId are required"))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", configId).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	instanceDetail, err := oc.ociService.GetInstanceDetails(ctx, &user, instanceId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	vnics := []models.ValueLabel{}
	for _, vnic := range instanceDetail.VnicList {
		label := vnic.Name
		if label == "" {
			label = vnic.VnicID
		}
		vnics = append(vnics, models.ValueLabel{
			Value: vnic.VnicID,
			Label: label,
		})
	}

	c.JSON(http.StatusOK, models.SuccessResponse(vnics, "Success"))
}

// GetSecurityListRequest 获取安全列表请求
type GetSecurityListRequest struct {
	ConfigID string `json:"configId" binding:"required"`
	VcnID    string `json:"vcnId" binding:"required"`
}

// GetSecurityList 获取VCN安全列表
func (oc *OciController) GetSecurityList(c *gin.Context) {
	var req GetSecurityListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	securityList, err := oc.ociService.GetSecurityListByVcnId(ctx, &user, req.VcnID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(securityList, "Success"))
}

// AddSecurityRuleRequest 添加安全规则请求
type AddSecurityRuleRequest struct {
	ConfigID    string `json:"configId" binding:"required"`
	VcnID       string `json:"vcnId" binding:"required"`
	IsIngress   bool   `json:"isIngress"`
	Protocol    string `json:"protocol" binding:"required"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	PortMin     int    `json:"portMin"`
	PortMax     int    `json:"portMax"`
	Description string `json:"description"`
	IsStateless bool   `json:"isStateless"`
}

// AddSecurityRule 添加安全规则
func (oc *OciController) AddSecurityRule(c *gin.Context) {
	var req AddSecurityRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	rule := &models.SecurityRule{
		Protocol:     req.Protocol,
		Source:       req.Source,
		Destination:  req.Destination,
		PortRangeMin: req.PortMin,
		PortRangeMax: req.PortMax,
		Description:  req.Description,
		IsStateless:  req.IsStateless,
	}

	ctx := context.Background()
	if err := oc.ociService.AddSecurityRule(ctx, &user, req.VcnID, rule, req.IsIngress); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "安全规则添加成功"))
}

// ReleaseSecurityRulesRequest 放行安全规则请求
type ReleaseSecurityRulesRequest struct {
	ConfigID string `json:"configId" binding:"required"`
	VcnID    string `json:"vcnId" binding:"required"`
}

// ReleaseSecurityRules 一键放行安全规则
func (oc *OciController) ReleaseSecurityRules(c *gin.Context) {
	var req ReleaseSecurityRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	if err := oc.ociService.ReleaseSecurityRules(&user, req.VcnID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "安全规则放行成功"))
}

// DeleteVcnRequest 删除VCN请求
type DeleteVcnRequest struct {
	ConfigID string `json:"configId" binding:"required"`
	VcnID    string `json:"vcnId" binding:"required"`
}

// DeleteVcn 删除VCN
func (oc *OciController) DeleteVcn(c *gin.Context) {
	var req DeleteVcnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	db := database.GetDB()
	var user models.OciUser
	if err := db.Where("id = ?", req.ConfigID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(404, "Configuration not found"))
		return
	}

	ctx := context.Background()
	if err := oc.ociService.DeleteVcn(ctx, &user, req.VcnID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "VCN删除成功"))
}
