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

// Upload godoc
// @Summary      Upload media file
// @Description  Uploads an image file (jpeg, png, webp) up to 5MB to MinIO and returns the media details. Requires authentication.
// @Tags         media
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Image file to upload"
// @Success      201 {object} response.WebResponse{data=media.MediaResponse} "File successfully uploaded"
// @Failure      400 {object} response.WebResponse{error=response.ErrorDetail} "Invalid file, missing field, or file too large"
// @Failure      401 {object} response.WebResponse{error=response.ErrorDetail} "Access denied, invalid or expired token"
// @Failure      500 {object} response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /media [post]
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

// Delete godoc
// @Summary      Delete media file
// @Description  Deletes a specific media file by its ID. Only the original uploader is authorized to delete it.
// @Tags         media
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Media ID (UUID)"
// @Success      204  "No Content"
// @Failure      401  {object}  response.WebResponse{error=response.ErrorDetail} "Access denied, invalid or expired token"
// @Failure      403  {object}  response.WebResponse{error=response.ErrorDetail} "Forbidden, you are not the uploader"
// @Failure      404  {object}  response.WebResponse{error=response.ErrorDetail} "Media not found"
// @Failure      500  {object}  response.WebResponse{error=response.ErrorDetail} "Internal server error"
// @Router       /media/{id} [delete]
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
