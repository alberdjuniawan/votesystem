package media

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"io"

	dbsqlc "github.com/alberdjuniawan/votesystem/internal/db/sqlc"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	miniopkg "github.com/alberdjuniawan/votesystem/internal/shared/minio"
	"github.com/alberdjuniawan/votesystem/internal/shared/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	maxFileSize = 5 * 1024 * 1024
)

var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type Service struct {
	repo        *Repository
	minioClient *miniopkg.Client
}

func NewService(repo *Repository, minioClient *miniopkg.Client) *Service {
	return &Service{
		repo:        repo,
		minioClient: minioClient,
	}
}

type UploadInput struct {
	File         io.Reader
	OriginalName string
	MimeType     string
	Size         int64
	UploaderID   string
}

func (s *Service) Upload(ctx context.Context, input UploadInput) (*MediaResponse, error) {
	if input.Size > maxFileSize {
		return nil, ErrFileTooLarge
	}

	ext, ok := allowedMimeTypes[input.MimeType]
	if !ok {
		return nil, ErrInvalidMimeType
	}

	objectName := fmt.Sprintf("options/%s%s", uuid.New().String(), ext)

	if err := s.minioClient.Upload(ctx, miniopkg.UploadInput{
		ObjectName:  objectName,
		Reader:      input.File,
		Size:        input.Size,
		ContentType: input.MimeType,
	}); err != nil {
		logger.Error(ctx, "Failed to upload file to MinIO", "error", err)
		return nil, ErrInternal
	}

	uploaderUID, err := utils.StrToPgUUID(input.UploaderID)
	if err != nil {
		return nil, ErrInternal
	}

	safeOriginalName := filepath.Base(input.OriginalName)

	media, err := s.repo.CreateMedia(ctx, dbsqlc.CreateMediaParams{
		UploaderID:   uploaderUID,
		Filename:     objectName,
		OriginalName: safeOriginalName,
		MimeType:     input.MimeType,
		SizeBytes:    input.Size,
		StoragePath:  objectName,
	})
	if err != nil {
		if delErr := s.minioClient.Delete(ctx, objectName); delErr != nil {
			logger.Error(ctx, "Failed to delete orphan file after DB error", "object", objectName, "error", delErr)
		}
		logger.Error(ctx, "Failed to save media metadata", "error", err)
		return nil, ErrInternal
	}

	return toMediaResponse(media, s.minioClient.GetPublicURL(media.StoragePath)), nil
}

func (s *Service) Delete(ctx context.Context, mediaID, userID string) error {
	media, err := s.repo.GetMediaByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMediaNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting media", "error", err)
		return ErrInternal
	}

	if utils.PgUUIDToStr(media.UploaderID) != userID {
		return ErrNotUploader
	}

	if err := s.minioClient.Delete(ctx, media.StoragePath); err != nil {
		logger.Error(ctx, "Failed to delete file from MinIO", "path", media.StoragePath, "error", err)
	}

	if err := s.repo.DeleteMedia(ctx, mediaID); err != nil {
		logger.Error(ctx, "Failed to delete media from DB", "media_id", mediaID, "error", err)
		return ErrInternal
	}

	return nil
}

func (s *Service) GetByID(ctx context.Context, mediaID string) (*MediaResponse, error) {
	media, err := s.repo.GetMediaByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMediaNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting media", "error", err)
		return nil, ErrInternal
	}

	url := s.minioClient.GetPublicURL(media.StoragePath)
	return toMediaResponse(media, url), nil
}

func toMediaResponse(m dbsqlc.Medium, url string) *MediaResponse {
	return &MediaResponse{
		ID:           utils.PgUUIDToStr(m.ID),
		Filename:     m.Filename,
		OriginalName: m.OriginalName,
		MimeType:     m.MimeType,
		SizeBytes:    m.SizeBytes,
		URL:          url,
		CreatedAt:    m.CreatedAt.Time,
	}
}
