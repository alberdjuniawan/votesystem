package media

import (
	"errors"
	"net/http"

	"github.com/alberdjuniawan/votesystem/internal/middleware"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/alberdjuniawan/votesystem/internal/shared/response"
	"github.com/gin-gonic/gin"
)

const maxMultipartMemory = 5 * 1024 * 1024

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Upload(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxMultipartMemory); err != nil {
		response.NewError(c, response.ErrBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.NewError(c, response.ErrBadRequest, "file field is required")
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		response.NewError(c, response.ErrBadRequest, "failed to read file")
		return
	}
	mimeType := http.DetectContentType(buffer)

	if _, err := file.Seek(0, 0); err != nil {
		response.NewError(c, response.ErrInternal, nil)
		return
	}

	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)

	result, err := h.service.Upload(ctx, UploadInput{
		File:         file,
		OriginalName: header.Filename,
		MimeType:     mimeType,
		Size:         header.Size,
		UploaderID:   userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrFileTooLarge):
			response.NewError(c, response.ErrBadRequest, err.Error())
		case errors.Is(err, ErrInvalidMimeType):
			response.NewError(c, response.ErrBadRequest, err.Error())
		default:
			logger.Error(ctx, "Upload failed", "user_id", userID, "filename", header.Filename)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Media uploaded", "media_id", result.ID, "user_id", userID)
	response.Created(c, result)
}

func (h *Handler) Delete(c *gin.Context) {
	mediaID := c.Param("id")
	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)

	err := h.service.Delete(ctx, mediaID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrMediaNotFound):
			response.NewError(c, response.ErrNotFound, err.Error())
		case errors.Is(err, ErrNotUploader):
			response.NewError(c, response.ErrForbidden, err.Error())
		default:
			logger.Error(ctx, "Delete media failed", "media_id", mediaID, "user_id", userID)
			response.NewError(c, response.ErrInternal, nil)
		}
		return
	}

	logger.Info(ctx, "Media deleted", "media_id", mediaID)
	response.NoContent(c)
}
