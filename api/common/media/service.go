package media

import (
	"context"
	"fmt"
	"os"

	"sync-backend/arch/config"
	"sync-backend/utils"

	"sync-backend/api/common/media/model"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type MediaService interface {
	UploadMedia(filePath string, filename string, folderPath string) (model.MediaInfo, error)
	DeleteMedia(publicID string) error
}

type mediaService struct {
	logger        utils.AppLogger
	mediaUploader *cloudinary.Cloudinary
}

func NewMediaService(env config.Env) MediaService {
	cld, err := cloudinary.NewFromParams(env.CloudinaryCloudName, env.CloudinaryAPIKey, env.CloudinaryAPISecret)
	if err != nil {
		panic("Failed to initialize Cloudinary - Properly initialize Cloudinary: " + err.Error())
	}
	return &mediaService{
		logger:        utils.NewServiceLogger("MediaService"),
		mediaUploader: cld,
	}
}

func (s *mediaService) UploadMedia(filePath string, filename string, folderPath string) (model.MediaInfo, error) {
	s.logger.Debug("Uploading media file: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		s.logger.Error("Failed to open file %s: %v", filePath, err)
		return model.MediaInfo{}, err
	}
	defer file.Close()

	uploadFolder := fmt.Sprintf("sync-backend/%s", folderPath)
	uploadParams := uploader.UploadParams{
		PublicID:       filename,
		Folder:         uploadFolder,
		UniqueFilename: api.Bool(true),
		Overwrite:      api.Bool(true),
	}
	resp, err := s.mediaUploader.Upload.Upload(context.Background(), file, uploadParams)

	if err != nil {
		s.logger.Error("Failed to upload file to Cloudinary: %v", err)
		return model.MediaInfo{}, err
	}

	s.logger.Info("Media uploaded successfully: publicID=%s, url=%s", resp.PublicID, resp.SecureURL)

	return model.MediaInfo{
		Id:     resp.PublicID,
		Url:    resp.SecureURL,
		Width:  resp.Width,
		Height: resp.Height,
	}, nil
}

func (s *mediaService) DeleteMedia(publicID string) error {
	s.logger.Debug("Deleting media file with publicID: %s", publicID)
	deleteParams := uploader.DestroyParams{
		PublicID: publicID,
	}
	resp, err := s.mediaUploader.Upload.Destroy(context.Background(), deleteParams)
	if err != nil {
		s.logger.Error("Failed to delete file from Cloudinary: %v", err)
		return err
	}

	if resp.Result != "ok" {
		s.logger.Error("Failed to delete file from Cloudinary: %s", resp.Result)
		return fmt.Errorf("failed to delete file from Cloudinary: %s", resp.Result)
	}

	s.logger.Info("Media deleted successfully: publicID=%s", publicID)
	return nil
}
