package controllers

import (
	"github.com/adiecho/oci-panel/internal/config"
	"github.com/adiecho/oci-panel/internal/database"
	"github.com/adiecho/oci-panel/internal/middleware"
	"github.com/adiecho/oci-panel/internal/models"
	"github.com/adiecho/oci-panel/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SysController struct {
	cfg              *config.Config
	schedulerService *services.SchedulerService
}

func NewSysController(cfg *config.Config, schedulerService *services.SchedulerService) *SysController {
	return &SysController{
		cfg:              cfg,
		schedulerService: schedulerService,
	}
}

type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	NeedMFA  bool   `json:"needMfa"`
}

func (sc *SysController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	if req.Account != sc.cfg.Web.Account || req.Password != sc.cfg.Web.Password {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(401, "Invalid credentials"))
		return
	}

	token, err := middleware.GenerateToken(req.Account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(LoginResponse{
		Token:    token,
		Username: req.Account,
		NeedMFA:  false,
	}, "Login successful"))
}

type GlanceResponse struct {
	TotalConfigs int64 `json:"totalConfigs"`
	TotalTasks   int64 `json:"totalTasks"`
}

func (sc *SysController) GetGlance(c *gin.Context) {
	db := database.GetDB()

	var totalConfigs int64
	db.Model(&models.OciUser{}).Count(&totalConfigs)

	var totalTasks int64
	db.Model(&models.OciCreateTask{}).Count(&totalTasks)

	c.JSON(http.StatusOK, models.SuccessResponse(GlanceResponse{
		TotalConfigs: totalConfigs,
		TotalTasks:   totalTasks,
	}, "success"))
}

type SysCfgResponse struct {
	LogLevel      string `json:"logLevel"`
	AIEnabled     bool   `json:"aiEnabled"`
	CacheEnabled  bool   `json:"cacheEnabled"`
	CacheInterval int    `json:"cacheInterval"`
}

func (sc *SysController) GetSysCfg(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(SysCfgResponse{
		LogLevel:      sc.cfg.Logging.Level,
		AIEnabled:     sc.cfg.AI.APIKey != "",
		CacheEnabled:  sc.schedulerService.IsCacheEnabled(),
		CacheInterval: sc.schedulerService.GetCacheInterval(),
	}, "success"))
}

type UpdateCacheCfgRequest struct {
	CacheEnabled  bool `json:"cacheEnabled"`
	CacheInterval int  `json:"cacheInterval"`
}

func (sc *SysController) UpdateCacheCfg(c *gin.Context) {
	var req UpdateCacheCfgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	if err := sc.schedulerService.SetCacheEnabled(req.CacheEnabled); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to update cache enabled"))
		return
	}

	if req.CacheInterval > 0 {
		if err := sc.schedulerService.SetCacheInterval(req.CacheInterval); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "Failed to update cache interval"))
			return
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "Cache configuration updated"))
}

func (sc *SysController) RefreshCache(c *gin.Context) {
	if !sc.schedulerService.IsCacheEnabled() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "Cache is not enabled"))
		return
	}

	sc.schedulerService.RefreshAllCaches()
	c.JSON(http.StatusOK, models.SuccessResponse(nil, "Cache refresh started"))
}
