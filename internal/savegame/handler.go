// Package savegame - 存档模块HTTP处理层
// 功能：处理存档相关的HTTP请求
// 架构：MVC中的Controller层
package savegame

import (
	errcode "faulty_in_culture/go_back/internal/shared/errors"
	"faulty_in_culture/go_back/internal/shared/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ============================================================
// Handler层 - MVC的Controller
// 职责：HTTP请求处理和响应
// ============================================================

// Handler 存档处理器
type Handler struct {
	service *Service
}

// NewHandler 创建存档处理器
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// QueryBySlot 查询指定槽位的存档
// @Summary 查询存档
// @Tags 存档
// @Produce json
// @Param slot_number query int true "槽位号(1-6)" minimum(1) maximum(6)
// @Success 200 {object} response.Response{data=SaveGameVO}
// @Router /api/savegame [get]
func (h *Handler) QueryBySlot(c *gin.Context) {
	userID := c.GetUint("user_id")
	slotNumber, _ := strconv.Atoi(c.Query("slot_number"))

	if slotNumber < 1 || slotNumber > 6 {
		response.Error(c, http.StatusBadRequest, errcode.InvalidSlotNumber)
		return
	}

	save, err := h.service.QueryBySlot(userID, slotNumber)
	if err != nil {
		if strings.Contains(err.Error(), "存档不存在") {
			response.Error(c, http.StatusNotFound, errcode.SaveGameNotFound)
		} else {
			response.Error(c, http.StatusInternalServerError, errcode.ServerError)
		}
		return
	}

	vo := SaveGameVO{
		UserID:     save.UserID,
		SlotNumber: save.SlotNumber,
		GameData:   save.GameData,
		SavedAt:    save.SavedAt,
	}

	response.Success(c, vo)
}

// QueryAll 查询所有存档
// @Summary 查询所有存档
// @Tags 存档
// @Produce json
// @Success 200 {object} response.Response{data=SaveGameListResponse}
// @Router /api/savegame/all [get]
func (h *Handler) QueryAll(c *gin.Context) {
	userID := c.GetUint("user_id")

	saves, err := h.service.QueryAll(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errcode.ServerError)
		return
	}

	vos := make([]SaveGameVO, len(saves))
	for i, s := range saves {
		vos[i] = SaveGameVO{
			UserID:     s.UserID,
			SlotNumber: s.SlotNumber,
			GameData:   s.GameData,
			SavedAt:    s.SavedAt,
		}
	}

	response.Success(c, SaveGameListResponse{Total: len(vos), List: vos})
}

// CreateOrUpdate 创建或更新存档
// @Summary 创建或更新存档
// @Tags 存档
// @Accept json
// @Produce json
// @Param request body CreateRequest true "存档数据"
// @Success 200 {object} response.Response{data=SaveGameVO}
// @Router /api/savegame [post]
func (h *Handler) CreateOrUpdate(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.InvalidParams)
		return
	}

	save, err := h.service.CreateOrUpdate(userID, req.SlotNumber, req.GameData)
	if err != nil {
		if strings.Contains(err.Error(), "槽位号无效") {
			response.Error(c, http.StatusBadRequest, errcode.InvalidSlotNumber)
		} else {
			response.Error(c, http.StatusInternalServerError, errcode.ServerError)
		}
		return
	}

	vo := SaveGameVO{
		UserID:     save.UserID,
		SlotNumber: save.SlotNumber,
		GameData:   save.GameData,
		SavedAt:    save.SavedAt,
	}

	response.SuccessWithMessage(c, "保存成功", vo)
}

// Delete 删除存档
// @Summary 删除存档
// @Tags 存档
// @Produce json
// @Param slot_number query int true "槽位号(1-6)" minimum(1) maximum(6)
// @Success 200 {object} response.Response
// @Router /api/savegame [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	slotNumber, _ := strconv.Atoi(c.Query("slot_number"))

	if slotNumber < 1 || slotNumber > 6 {
		response.Error(c, http.StatusBadRequest, errcode.InvalidSlotNumber)
		return
	}

	err := h.service.Delete(userID, slotNumber)
	if err != nil {
		if strings.Contains(err.Error(), "存档不存在") {
			response.Error(c, http.StatusNotFound, errcode.SaveGameNotFound)
		} else {
			response.Error(c, http.StatusInternalServerError, errcode.ServerError)
		}
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
