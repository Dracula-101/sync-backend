package middleware

import (
	"fmt"
	"os"
	"path/filepath"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"
	"time"

	"slices"

	"github.com/gin-gonic/gin"
)

type UploadProvider interface {
	network.BaseMiddlewareProvider
	Middleware(string) gin.HandlerFunc
	GetUploadedFiles(ctx *gin.Context) *UploadedFiles
	DeleteUploadedFiles(ctx *gin.Context) error
}

type uploadMiddleware struct {
	network.ResponseSender
	common.ContextPayload
	logger utils.AppLogger
	config *FileUploadConfig
}

type UploadedFiles struct {
	Files []UploadedFile `json:"files"`
}

type UploadedFile struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
	URL      string `json:"url"`
}

type FileUploadConfig struct {
	StoragePath       string
	MaxSize           int64
	AllowedExtensions []string
	UseUserID         bool
}

func NewUploadProvider() *uploadMiddleware {
	return &uploadMiddleware{
		ResponseSender: network.NewResponseSender(),
		ContextPayload: common.NewContextPayload(),
		logger:         utils.NewServiceLogger("UploadMiddleware"),
		config: &FileUploadConfig{
			StoragePath:       "./uploads",
			MaxSize:           50 * 1024 * 1024, // 50 MB
			AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".mp4", ".mov"},
			UseUserID:         true,
		},
	}
}

func (p *uploadMiddleware) Middleware(fieldName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" && c.Request.Method != "PUT" {
			c.Next()
			return
		}

		var userID string
		if p.config.UseUserID {
			userIDValue := p.ContextPayload.MustGetUserId(c)
			if userIDValue == nil {
				p.logger.Warn("User ID not found in context")
			} else {
				userID = *userIDValue
			}
		}

		storageDir := p.config.StoragePath
		if p.config.UseUserID && userID != "" {
			storageDir = filepath.Join(storageDir, userID)
		}
		if err := os.MkdirAll(storageDir, 0755); err != nil {
			p.logger.Error("Failed to create upload directory: %v", err)
			c.Next()
			return
		}
		form, err := c.MultipartForm()
		if err != nil {
			c.Next()
			return
		}
		files := form.File[fieldName]
		if len(files) == 0 {
			p.logger.Warn("No files found in the request")
			c.Next()
			return
		}
		uploadedFiles := &UploadedFiles{
			Files: make([]UploadedFile, 0, len(files)),
		}

		for _, file := range files {
			p.logger.Debug("Processing file: %s", file.Filename)

			if p.config.MaxSize > 0 && file.Size > p.config.MaxSize {
				p.logger.Warn("File %s exceeds maximum size limit of %d bytes", file.Filename, p.config.MaxSize)
				continue
			}
			if len(p.config.AllowedExtensions) > 0 {
				ext := filepath.Ext(file.Filename)
				allowed := slices.Contains(p.config.AllowedExtensions, ext)
				if !allowed {
					p.logger.Warn("File %s has disallowed extension: %s", file.Filename, ext)
					continue
				}
			}

			uniqueFilename := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
			filePath := filepath.Join(storageDir, uniqueFilename)

			if err := c.SaveUploadedFile(file, filePath); err != nil {
				p.logger.Error("Failed to save file %s: %v", file.Filename, err)
				continue
			}

			urlPath := fmt.Sprintf("/uploads/%s/%s", userID, uniqueFilename)
			if !p.config.UseUserID || userID == "" {
				urlPath = fmt.Sprintf("/uploads/common/%s", uniqueFilename)
			}

			uploadedFiles.Files = append(uploadedFiles.Files, UploadedFile{
				Filename: file.Filename,
				Size:     file.Size,
				Path:     filePath,
				URL:      urlPath,
			})

			p.logger.Info("File uploaded successfully: %s -> %s", file.Filename, filePath)
		}
		c.Set("uploadedFiles", uploadedFiles)
		c.Next()
	}
}

func (p *uploadMiddleware) GetUploadedFiles(c *gin.Context) *UploadedFiles {
	files, exists := c.Get("uploadedFiles")
	if !exists {
		p.logger.Warn("No uploaded files found in context")
		return &UploadedFiles{Files: []UploadedFile{}}
	}

	if uploadedFiles, ok := files.(*UploadedFiles); ok {
		return uploadedFiles
	}

	p.logger.Warn("Uploaded files in context are not of type *UploadedFiles")
	return &UploadedFiles{Files: []UploadedFile{}}
}

func (p *uploadMiddleware) DeleteUploadedFiles(c *gin.Context) error {
	files, exists := c.Get("uploadedFiles")
	if !exists {
		p.logger.Warn("No uploaded files found in context")
		return nil
	}

	if uploadedFiles, ok := files.(*UploadedFiles); ok {
		for _, file := range uploadedFiles.Files {
			if err := os.Remove(file.Path); err != nil {
				p.logger.Error("Failed to delete file %s: %v", file.Path, err)
				return err
			}
			p.logger.Info("File deleted successfully: %s", file.Path)
		}

	} else {
		p.logger.Warn("Uploaded files in context are not of type *UploadedFiles")
	}
	return nil
}
