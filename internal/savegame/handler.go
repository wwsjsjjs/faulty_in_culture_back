package savegame

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ============================================================
// Handler层 - MVC的Controller
// 职责：HTTP请求处理和响应
// ============================================================

// Response 通用响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

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
// @Success 200 {object} Response{data=SaveGameVO}
// @Router /api/savegame [get]
func (h *Handler) QueryBySlot(c *gin.Context) {
	userID := c.GetUint("user_id")
	slotNumber, _ := strconv.Atoi(c.Query("slot_number"))

	if slotNumber < 1 || slotNumber > 6 {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: "槽位号必须在1-6之间"})
		return
	}

	save, err := h.service.QueryBySlot(userID, slotNumber)
	if err != nil {
		if err == ErrSaveGameNotFound {
			c.JSON(http.StatusNotFound, Response{Code: 404, Msg: "存档不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, Response{Code: 500, Msg: err.Error()})
		}
		return
	}

	vo := SaveGameVO{
		UserID:     save.UserID,
		SlotNumber: save.SlotNumber,
		GameData:   save.GameData,
		SavedAt:    save.SavedAt,
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: vo})
}

// QueryAll 查询所有存档
// @Summary 查询所有存档
// @Tags 存档
// @Produce json
// @Success 200 {object} Response{data=SaveGameListResponse}
// @Router /api/savegame/all [get]
func (h *Handler) QueryAll(c *gin.Context) {
	userID := c.GetUint("user_id")

	saves, err := h.service.QueryAll(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Msg: err.Error()})
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

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: SaveGameListResponse{Total: len(vos), List: vos},
	})
}

// CreateOrUpdate 创建或更新存档
// @Summary 创建或更新存档
// @Tags 存档
// @Accept json
// @Produce json
// @Param request body CreateRequest true "存档数据"
// @Success 200 {object} Response{data=SaveGameVO}
// @Router /api/savegame [post]
func (h *Handler) CreateOrUpdate(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: "参数错误"})
		return
	}

	save, err := h.service.CreateOrUpdate(userID, req.SlotNumber, req.GameData)
	if err != nil {
		if err == ErrInvalidSlotNumber {
			c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, Response{Code: 500, Msg: err.Error()})
		}
		return
	}

	vo := SaveGameVO{
		UserID:     save.UserID,
		SlotNumber: save.SlotNumber,
		GameData:   save.GameData,
		SavedAt:    save.SavedAt,
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "保存成功", Data: vo})
}

// Delete 删除存档
// @Summary 删除存档
// @Tags 存档
// @Produce json
// @Param slot_number query int true "槽位号(1-6)" minimum(1) maximum(6)
// @Success 200 {object} Response
// @Router /api/savegame [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	slotNumber, _ := strconv.Atoi(c.Query("slot_number"))

	if slotNumber < 1 || slotNumber > 6 {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: "槽位号必须在1-6之间"})
		return
	}

	err := h.service.Delete(userID, slotNumber)
	if err != nil {
		if err == ErrSaveGameNotFound {
			c.JSON(http.StatusNotFound, Response{Code: 404, Msg: "存档不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, Response{Code: 500, Msg: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "删除成功"})
}
