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
	Middleware(fieldNames ...string) gin.HandlerFunc
	GetUploadedFiles(ctx *gin.Context, fieldname string) *UploadedFiles
	DeleteUploadedFiles(ctx *gin.Context, fieldname string) error
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

func (uf *UploadedFiles) IsEmpty() bool {
	return len(uf.Files) == 0
}

func (uf *UploadedFiles) First() (*UploadedFile, error) {
	if len(uf.Files) == 0 {
		return nil, fmt.Errorf("no files uploaded")
	}
	return &uf.Files[0], nil
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

func (p *uploadMiddleware) Middleware(fieldNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" && c.Request.Method != "PUT" {
			c.Next()
			return
		}

		var userID string
		if p.config.UseUserID {
			userIDValue := p.ContextPayload.GetUserId(c)
			if userIDValue != nil {
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
		for _, fieldName := range fieldNames {
			files := form.File[fieldName]
			// Print all files for debugging
			p.logger.Debug("Processing field: %s, files count: %d", fieldName, len(files))

			if len(files) == 0 {
				p.logger.Debug("No files found for field: %s", fieldName)
				continue
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

			if len(uploadedFiles.Files) > 0 {
				c.Set("uploadedFiles-"+fieldName, uploadedFiles)
			}
		}
		c.Next()
	}
}

func (p *uploadMiddleware) GetUploadedFiles(c *gin.Context, fieldName string) *UploadedFiles {
	files, exists := c.Get("uploadedFiles-" + fieldName)
	if !exists {
		p.logger.Debug("No uploaded files found in context for field: %s", fieldName)
		return &UploadedFiles{Files: []UploadedFile{}}
	}

	uploadedFiles, ok := files.(*UploadedFiles)
	if !ok {
		p.logger.Error("Invalid type in context for field %s: expected *UploadedFiles", fieldName)
		return &UploadedFiles{Files: []UploadedFile{}}
	}

	// Filter out any potentially invalid files
	if len(uploadedFiles.Files) > 0 {
		validFiles := make([]UploadedFile, 0, len(uploadedFiles.Files))
		for _, file := range uploadedFiles.Files {
			if file.Path != "" && file.Filename != "" {
				// Check if file actually exists
				if _, err := os.Stat(file.Path); err == nil {
					validFiles = append(validFiles, file)
				} else {
					p.logger.Warn("File doesn't exist on disk: %s", file.Path)
				}
			}
		}
		uploadedFiles.Files = validFiles
	}

	return uploadedFiles
}

func (p *uploadMiddleware) DeleteUploadedFiles(c *gin.Context, fieldName string) error {
	files, exists := c.Get("uploadedFiles-" + fieldName)
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
