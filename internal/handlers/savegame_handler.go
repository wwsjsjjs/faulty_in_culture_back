package handlers

import (
	"faulty_in_culture/go_back/internal/logger"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"faulty_in_culture/go_back/internal/cache"
	"faulty_in_culture/go_back/internal/database"
	"faulty_in_culture/go_back/internal/dto"
	"faulty_in_culture/go_back/internal/models"
	"faulty_in_culture/go_back/internal/vo"

	"github.com/gin-gonic/gin"
)

// SaveGameHandler 存档处理器
type SaveGameHandler struct{}

// NewSaveGameHandler 创建存档处理器
func NewSaveGameHandler() *SaveGameHandler {
	return &SaveGameHandler{}
}

// GetSaveGames 获取用户所有存档
// @Summary 获取用户所有存档
// @Description 获取当前用户的所有存档（最多6个）
// @Tags savegame
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} vo.SaveGameResponse
// @Failure 401 {object} vo.ErrorResponse
// @Router /api/savegames [get]
func (h *SaveGameHandler) GetSaveGames(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		logger.Warn("未授权访问存档列表", zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{Error: "未授权"})
		return
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("savegames:user:%d", userID)
	var responses []vo.SaveGameResponse

	cacheClient := cache.GetCache()
	if cacheClient != nil {
		err := cacheClient.Get(cacheKey, &responses)
		if err == nil && len(responses) >= 0 {
			// 缓存命中
			c.JSON(http.StatusOK, responses)
			return
		}
	}

	var saveGames []models.SaveGame
	if err := database.DB.Where("user_id = ?", userID).Order("slot_number ASC").Find(&saveGames).Error; err != nil {
		logger.Error("查询存档失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: "查询存档失败"})
		return
	}

	responses = make([]vo.SaveGameResponse, len(saveGames))
	for i, sg := range saveGames {
		responses[i] = vo.SaveGameResponse{
			ID:         sg.ID,
			UserID:     sg.UserID,
			SlotNumber: sg.SlotNumber,
			Data:       sg.Data,
			CreatedAt:  sg.CreatedAt,
			UpdatedAt:  sg.UpdatedAt,
		}
	}

	// 缓存结果（5分钟）
	if cacheClient != nil {
		_ = cacheClient.Set(cacheKey, responses, 5*time.Minute)
	}

	c.JSON(http.StatusOK, responses)
}

// GetSaveGame 获取指定槽位的存档
// @Summary 获取指定槽位的存档
// @Description 获取当前用户指定槽位的存档
// @Tags savegame
// @Security ApiKeyAuth
// @Produce json
// @Param slot path int true "存档槽位(1-6)"
// @Success 200 {object} vo.SaveGameResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 404 {object} vo.ErrorResponse
// @Router /api/savegames/{slot} [get]
func (h *SaveGameHandler) GetSaveGame(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		logger.Warn("未授权访问单个存档", zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{Error: "未授权"})
		return
	}

	slotNumber, err := strconv.Atoi(c.Param("slot"))
	if err != nil || slotNumber < 1 || slotNumber > 6 {
		logger.Warn("槽位号参数错误", zap.String("slot", c.Param("slot")))
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "槽位号必须在1-6之间"})
		return
	}

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("savegame:user:%d:slot:%d", userID, slotNumber)
	var response vo.SaveGameResponse

	cacheClient := cache.GetCache()
	if cacheClient != nil {
		err := cacheClient.Get(cacheKey, &response)
		if err == nil && response.ID > 0 {
			// 缓存命中
			c.JSON(http.StatusOK, response)
			return
		}
	}

	var saveGame models.SaveGame
	if err := database.DB.Where("user_id = ? AND slot_number = ?", userID, slotNumber).First(&saveGame).Error; err != nil {
		logger.Warn("存档不存在", zap.Int("slot", slotNumber))
		c.JSON(http.StatusNotFound, vo.ErrorResponse{Error: "存档不存在"})
		return
	}

	response = vo.SaveGameResponse{
		ID:         saveGame.ID,
		UserID:     saveGame.UserID,
		SlotNumber: saveGame.SlotNumber,
		Data:       saveGame.Data,
		CreatedAt:  saveGame.CreatedAt,
		UpdatedAt:  saveGame.UpdatedAt,
	}

	// 缓存结果（5分钟）
	if cacheClient != nil {
		_ = cacheClient.Set(cacheKey, response, 5*time.Minute)
	}

	c.JSON(http.StatusOK, response)
}

// CreateOrUpdateSaveGame 创建或更新存档
// @Summary 创建或更新存档
// @Description 在指定槽位创建或更新存档
// @Tags savegame
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param slot path int true "存档槽位(1-6)"
// @Param data body dto.SaveGameRequest true "存档数据"
// @Success 200 {object} vo.SaveGameResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 400 {object} vo.ErrorResponse
// @Router /api/savegames/{slot} [put]
func (h *SaveGameHandler) CreateOrUpdateSaveGame(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		logger.Warn("未授权创建/更新存档", zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{Error: "未授权"})
		return
	}

	slotNumber, err := strconv.Atoi(c.Param("slot"))
	if err != nil || slotNumber < 1 || slotNumber > 6 {
		logger.Warn("槽位号参数错误", zap.String("slot", c.Param("slot")))
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "槽位号必须在1-6之间"})
		return
	}

	var req dto.SaveGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("存档参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "参数错误"})
		return
	}

	// 查找是否已存在
	var saveGame models.SaveGame
	result := database.DB.Where("user_id = ? AND slot_number = ?", userID, slotNumber).First(&saveGame)

	if result.Error != nil {
		// 不存在，创建新存档
		saveGame = models.SaveGame{
			UserID:     userID,
			SlotNumber: slotNumber,
			Data:       req.Data,
		}
		if err := database.DB.Create(&saveGame).Error; err != nil {
			logger.Error("创建存档失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: "创建存档失败"})
			return
		}
	} else {
		// 已存在，更新
		saveGame.Data = req.Data
		if err := database.DB.Save(&saveGame).Error; err != nil {
			logger.Error("更新存档失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: "更新存档失败"})
			return
		}
	}

	// 清除相关缓存
	cacheClient := cache.GetCache()
	if cacheClient != nil {
		_ = cacheClient.Delete(fmt.Sprintf("savegames:user:%d", userID))
		_ = cacheClient.Delete(fmt.Sprintf("savegame:user:%d:slot:%d", userID, slotNumber))
	}

	c.JSON(http.StatusOK, vo.SaveGameResponse{
		ID:         saveGame.ID,
		UserID:     saveGame.UserID,
		SlotNumber: saveGame.SlotNumber,
		Data:       saveGame.Data,
		CreatedAt:  saveGame.CreatedAt,
		UpdatedAt:  saveGame.UpdatedAt,
	})
}

// DeleteSaveGame 删除存档
// @Summary 删除存档
// @Description 删除指定槽位的存档
// @Tags savegame
// @Security ApiKeyAuth
// @Produce json
// @Param slot path int true "存档槽位(1-6)"
// @Success 200 {object} vo.SuccessMessageResponse
// @Failure 401 {object} vo.ErrorResponse
// @Failure 404 {object} vo.ErrorResponse
// @Router /api/savegames/{slot} [delete]
func (h *SaveGameHandler) DeleteSaveGame(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		logger.Warn("未授权删除存档", zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusUnauthorized, vo.ErrorResponse{Error: "未授权"})
		return
	}

	slotNumber, err := strconv.Atoi(c.Param("slot"))
	if err != nil || slotNumber < 1 || slotNumber > 6 {
		logger.Warn("槽位号参数错误", zap.String("slot", c.Param("slot")))
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "槽位号必须在1-6之间"})
		return
	}

	var saveGame models.SaveGame
	if err := database.DB.Where("user_id = ? AND slot_number = ?", userID, slotNumber).First(&saveGame).Error; err != nil {
		logger.Warn("删除存档时未找到", zap.Int("slot", slotNumber))
		c.JSON(http.StatusNotFound, vo.ErrorResponse{Error: "存档不存在"})
		return
	}

	if err := database.DB.Delete(&saveGame).Error; err != nil {
		logger.Error("删除存档失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: "删除存档失败"})
		return
	}

	// 清除相关缓存
	cacheClient := cache.GetCache()
	if cacheClient != nil {
		_ = cacheClient.Delete(fmt.Sprintf("savegames:user:%d", userID))
		_ = cacheClient.Delete(fmt.Sprintf("savegame:user:%d:slot:%d", userID, slotNumber))
	}

	c.JSON(http.StatusOK, vo.SuccessMessageResponse{Message: "存档删除成功"})
}
