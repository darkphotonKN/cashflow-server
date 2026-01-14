package upload

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

type Service interface {
	RequestUpload(ctx context.Context, req UploadRequest) (*UploadResponse, error)
	GetUploadStatus(ctx context.Context, uploadID string) (*UploadStatusResponse, error)
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RequestUpload(c *gin.Context) {
	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind upload request",
			slog.String("error", err.Error()))
		c.JSON(400, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	response, err := h.service.RequestUpload(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to create upload request",
			slog.String("error", err.Error()),
			slog.String("content_type", req.ContentType),
			slog.Int64("file_size", req.FileSize))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, response)
}

func (h *Handler) GetUploadStatus(c *gin.Context) {
	uploadID := c.Param("id")
	if uploadID == "" {
		c.JSON(400, gin.H{"error": "upload ID is required"})
		return
	}

	status, err := h.service.GetUploadStatus(c.Request.Context(), uploadID)
	if err != nil {
		h.logger.Error("failed to get upload status",
			slog.String("error", err.Error()),
			slog.String("upload_id", uploadID))
		c.JSON(404, gin.H{"error": "Upload not found"})
		return
	}

	c.JSON(200, status)
}