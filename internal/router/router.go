package router

import (
	"github.com/adiecho/oci-panel/internal/config"
	"github.com/adiecho/oci-panel/internal/controllers"
	"github.com/adiecho/oci-panel/internal/middleware"
	"github.com/adiecho/oci-panel/internal/services"
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, cfg *config.Config) *services.SchedulerService {
	r.Use(middleware.CORS())
	r.Use(middleware.AuthMiddleware())

	// 静态资源 - 前端构建文件
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// 直接访问根路径返回前端页面
	r.GET("/", func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	ociService := services.NewOCIService(cfg)
	instanceService := services.NewInstanceService(ociService)
	ipService := services.NewIpService(ociService)
	_ = services.NewVolumeService(ociService)
	wsService := services.NewWebSocketService()
	schedulerService := services.NewSchedulerService(ociService)

	wsCtrl := controllers.NewWebSocketController(wsService)
	r.GET("/ws/logs", wsCtrl.HandleWebSocket)

	api := r.Group("/api")
	{
		sysCtrl := controllers.NewSysController(cfg, schedulerService)
		sys := api.Group("/sys")
		{
			sys.POST("/login", sysCtrl.Login)
			sys.POST("/getGlance", sysCtrl.GetGlance)
			sys.POST("/getSysCfg", sysCtrl.GetSysCfg)
			sys.POST("/updateCacheCfg", sysCtrl.UpdateCacheCfg)
			sys.POST("/refreshCache", sysCtrl.RefreshCache)
		}

		ociCtrl := controllers.NewOciController(ociService, schedulerService)
		oci := api.Group("/oci")
		{
			oci.POST("/userPage", ociCtrl.UserPage)
			oci.POST("/addCfg", ociCtrl.AddCfg)
			oci.POST("/updateCfgName", ociCtrl.UpdateCfgName)
			oci.POST("/removeCfg", ociCtrl.RemoveCfg)
			oci.POST("/createInstance", ociCtrl.CreateInstance)
			oci.POST("/createTaskPage", ociCtrl.CreateTaskPage)
			oci.POST("/uploadKey", ociCtrl.UploadKey)
			oci.POST("/details", ociCtrl.GetConfigDetails)
			oci.POST("/details/instances", ociCtrl.GetConfigInstances)
			oci.POST("/details/volumes", ociCtrl.GetConfigVolumes)
			oci.POST("/details/vcns", ociCtrl.GetConfigVCNs)
			oci.POST("/details/clearCache", ociCtrl.ClearConfigCache)
			oci.POST("/tenant/info", ociCtrl.GetTenantInfo)
			oci.POST("/tenant/updatePwdEx", ociCtrl.UpdatePasswordExpiry)
			oci.POST("/tenant/updateUserInfo", ociCtrl.UpdateUserInfo)
			oci.POST("/tenant/deleteUser", ociCtrl.DeleteUser)
			oci.POST("/tenant/resetPassword", ociCtrl.ResetPassword)
			oci.POST("/tenant/deleteMfaDevice", ociCtrl.DeleteMfaDevice)
			oci.POST("/tenant/deleteApiKey", ociCtrl.DeleteApiKey)
			oci.POST("/traffic/data", ociCtrl.GetTrafficData)
			oci.GET("/traffic/condition", ociCtrl.GetTrafficCondition)
			oci.GET("/traffic/vnics", ociCtrl.GetInstanceVnics)
		}

		instanceCtrl := controllers.NewInstanceController(instanceService)
		instance := api.Group("/instance")
		{
			instance.POST("/list", instanceCtrl.ListInstances)
			instance.POST("/start", instanceCtrl.StartInstance)
			instance.POST("/stop", instanceCtrl.StopInstance)
			instance.POST("/reboot", instanceCtrl.RebootInstance)
			instance.POST("/terminate", instanceCtrl.TerminateInstance)
			instance.POST("/updateName", instanceCtrl.UpdateInstanceName)
			instance.POST("/changeIP", instanceCtrl.ChangePublicIP)
			instance.POST("/updateConfig", instanceCtrl.UpdateInstanceConfig)
			instance.POST("/updateBootVolume", instanceCtrl.UpdateBootVolume)
			instance.POST("/createCloudShell", instanceCtrl.CreateCloudShell)
		}

		ipCtrl := controllers.NewIpController(ipService)
		ip := api.Group("/ip")
		{
			ip.POST("/change", ipCtrl.ChangePublicIp)
			ip.POST("/attachIpv6", ipCtrl.AttachIpv6)
		}
	}

	// SPA fallback - 所有未匹配的路由都返回 index.html，让前端路由接管
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	return schedulerService
}
