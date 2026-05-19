package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devflow/internal/response"

	"github.com/gin-gonic/gin"
)

const maxImageUploadSize = 4 * 1024 * 1024

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

func (a *App) uploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "image file is required")
		return
	}
	if file.Size <= 0 || file.Size > maxImageUploadSize {
		response.Error(c, http.StatusBadRequest, "image file is too large")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		response.Error(c, http.StatusBadRequest, "unsupported image type")
		return
	}

	dir := filepath.Join(".", "uploads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to prepare upload directory")
		return
	}

	name := randomHex(12) + "-" + time.Now().Format("20060102150405") + ext
	dst := filepath.Join(dir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to save image")
		return
	}

	response.OK(c, gin.H{"url": "/uploads/" + name})
}

func randomHex(size int) string {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "image"
	}
	return hex.EncodeToString(bytes)
}
