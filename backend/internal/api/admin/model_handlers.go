package admin

import (
	"net/http"
	"time"

	"google-ai-proxy/internal/db"

	"github.com/gin-gonic/gin"
)

// ListModels returns all platform models.
// GET /api/admin/models
func ListModels(c *gin.Context) {
	var models []db.PlatformModel
	if err := db.DB.Order("sort_order ASC, id ASC").Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}

type createModelRequest struct {
	ModelID   string `json:"model_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required"`
	ApiType   string `json:"api_type"` // task (default) | chat
	IconURL   string `json:"icon_url"`
	SortOrder int    `json:"sort_order"`
	Enabled   *bool  `json:"enabled"`
}

// CreateModel adds a new platform model.
// POST /api/admin/models
func CreateModel(c *gin.Context) {
	var req createModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: " + err.Error()})
		return
	}

	if req.Type != "image" && req.Type != "video" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type 必须是 image 或 video"})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	apiType := req.ApiType
	if apiType == "" {
		apiType = "task"
	}
	if apiType != "task" && apiType != "chat" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "api_type 必须是 task 或 chat"})
		return
	}

	model := db.PlatformModel{
		ModelID:   req.ModelID,
		Name:      req.Name,
		Type:      req.Type,
		ApiType:   apiType,
		IconURL:   req.IconURL,
		SortOrder: req.SortOrder,
		Enabled:   enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.DB.Create(&model).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "创建失败，model_id 可能已存在"})
		return
	}

	c.JSON(http.StatusOK, model)
}

type updateModelRequest struct {
	ModelID   *string `json:"model_id"`
	Name      *string `json:"name"`
	Type      *string `json:"type"`
	ApiType   *string `json:"api_type"`
	IconURL   *string `json:"icon_url"`
	SortOrder *int    `json:"sort_order"`
	Enabled   *bool   `json:"enabled"`
}

// UpdateModel updates an existing platform model.
// PUT /api/admin/models/:id
func UpdateModel(c *gin.Context) {
	id := c.Param("id")

	var model db.PlatformModel
	if err := db.DB.First(&model, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}

	var req updateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.ModelID != nil {
		updates["model_id"] = *req.ModelID
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		if *req.Type != "image" && *req.Type != "video" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type 必须是 image 或 video"})
			return
		}
		updates["type"] = *req.Type
	}
	if req.ApiType != nil {
		if *req.ApiType != "task" && *req.ApiType != "chat" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "api_type 必须是 task 或 chat"})
			return
		}
		updates["api_type"] = *req.ApiType
	}
	if req.IconURL != nil {
		updates["icon_url"] = *req.IconURL
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := db.DB.Model(&model).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	db.DB.First(&model, model.ID)
	c.JSON(http.StatusOK, model)
}

// DeleteModel removes a platform model.
// DELETE /api/admin/models/:id
func DeleteModel(c *gin.Context) {
	id := c.Param("id")
	result := db.DB.Delete(&db.PlatformModel{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// GetConfig returns a platform config value.
// GET /api/admin/config/:key
func GetConfig(c *gin.Context) {
	key := c.Param("key")
	value := db.GetConfig(key)
	c.JSON(http.StatusOK, gin.H{"key": key, "value": value})
}

type setConfigRequest struct {
	Value string `json:"value"`
}

// SetConfig sets a platform config value.
// PUT /api/admin/config/:key
func SetConfig(c *gin.Context) {
	key := c.Param("key")
	var req setConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	if err := db.SetConfig(key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": key, "value": req.Value})
}
