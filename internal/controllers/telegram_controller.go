package controllers

import (
	"net/http"

	"github.com/adiecho/oci-panel/internal/models"
	"github.com/adiecho/oci-panel/internal/services"
	"github.com/gin-gonic/gin"
)

type TelegramController struct {
	telegramService *services.TelegramService
}

func NewTelegramController(telegramService *services.TelegramService) *TelegramController {
	return &TelegramController{
		telegramService: telegramService,
	}
}

type TelegramConfigResponse struct {
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatId"`
	Enabled  bool   `json:"enabled"`
	Running  bool   `json:"running"`
}

func (tc *TelegramController) GetConfig(c *gin.Context) {
	botToken, chatID, enabled := tc.telegramService.GetConfig()

	maskedToken := ""
	if len(botToken) > 10 {
		maskedToken = botToken[:6] + "****" + botToken[len(botToken)-4:]
	} else if botToken != "" {
		maskedToken = "****"
	}

	c.JSON(http.StatusOK, models.SuccessResponse(TelegramConfigResponse{
		BotToken: maskedToken,
		ChatID:   chatID,
		Enabled:  enabled,
		Running:  tc.telegramService.IsRunning(),
	}, "success"))
}

type UpdateTelegramConfigRequest struct {
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatId"`
	Enabled  bool   `json:"enabled"`
}

func (tc *TelegramController) UpdateConfig(c *gin.Context) {
	var req UpdateTelegramConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	currentToken, _, _ := tc.telegramService.GetConfig()
	botToken := req.BotToken
	if botToken == "" || (len(botToken) > 4 && botToken[len(botToken)-4:] == "****") {
		botToken = currentToken
	}

	if err := tc.telegramService.UpdateConfig(botToken, req.ChatID, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(500, "更新配置失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "配置更新成功"))
}

func (tc *TelegramController) TestConnection(c *gin.Context) {
	if err := tc.telegramService.TestConnection(); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "连接测试失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "连接测试成功"))
}

type SendTestMessageRequest struct {
	Message string `json:"message"`
}

func (tc *TelegramController) SendTestMessage(c *gin.Context) {
	var req SendTestMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, err.Error()))
		return
	}

	message := req.Message
	if message == "" {
		message = "🔔 OCI Panel 测试消息\n\nTelegram 通知功能配置成功！"
	}

	if err := tc.telegramService.SendMessage(message); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "发送失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(nil, "消息发送成功"))
}

func (tc *TelegramController) StartBot(c *gin.Context) {
	botToken, chatID, enabled := tc.telegramService.GetConfig()
	if !enabled || botToken == "" || chatID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(400, "请先配置并启用 Telegram"))
		return
	}

	tc.telegramService.StartBot()
	c.JSON(http.StatusOK, models.SuccessResponse(nil, "Bot 已启动"))
}

func (tc *TelegramController) StopBot(c *gin.Context) {
	tc.telegramService.StopBot()
	c.JSON(http.StatusOK, models.SuccessResponse(nil, "Bot 已停止"))
}

type BotStatusResponse struct {
	Running bool `json:"running"`
}

func (tc *TelegramController) GetBotStatus(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(BotStatusResponse{
		Running: tc.telegramService.IsRunning(),
	}, "success"))
}
