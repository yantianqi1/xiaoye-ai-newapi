package api

import (
	"net/http"

	"google-ai-proxy/internal/db"

	"github.com/gin-gonic/gin"
)

// GetModels returns all enabled platform models from the database.
// GET /api/models
func GetModels(c *gin.Context) {
	var models []db.PlatformModel
	if err := db.DB.Where("enabled = ?", true).Order("sort_order ASC, id ASC").Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询模型失败"})
		return
	}

	type ModelResponse struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		IconURL string `json:"icon_url"`
	}

	result := make([]ModelResponse, 0, len(models))
	for _, m := range models {
		result = append(result, ModelResponse{
			ID:      m.ModelID,
			Name:    m.Name,
			Type:    m.Type,
			IconURL: m.IconURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{"models": result})
}
