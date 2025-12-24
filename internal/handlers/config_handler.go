package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"faulty_in_culture/go_back/internal/config"

	"github.com/gin-gonic/gin"
)

// GetPublicConfig 返回前端需要的公共配置，例如 API 基础地址
func GetPublicConfig(c *gin.Context) {
	base := strings.TrimRight(config.AppConfig.Server.PublicBaseURL, "/")
	if base == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		base = fmt.Sprintf("http://localhost:%s/api", port)
	}

	c.JSON(http.StatusOK, gin.H{
		"api_base_url": base,
	})
}
