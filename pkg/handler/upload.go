package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/Sovpalo/sovpalo-backend/pkg/service"
	"github.com/gin-gonic/gin"
)

func readMultipartImage(c *gin.Context, field string) (string, []byte, error) {
	fileHeader, err := c.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", nil, nil
		}
		return "", nil, errors.New("failed to read image file")
	}
	if fileHeader.Size > maxAvatarUploadSize {
		return "", nil, service.ErrAvatarTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", nil, errors.New("failed to open image file")
	}
	defer file.Close()

	fileData, err := io.ReadAll(io.LimitReader(file, maxAvatarUploadSize+1))
	if err != nil {
		return "", nil, errors.New("failed to read image file")
	}
	if len(fileData) > maxAvatarUploadSize {
		return "", nil, service.ErrAvatarTooLarge
	}

	return fileHeader.Filename, fileData, nil
}
