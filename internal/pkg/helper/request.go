package helper

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parxyws/cozybox/internal/domain"
)

func ReadImageRequest(c *gin.Context, fieldName string) (*domain.UploadInput, error) {
	image, err := c.FormFile(fieldName)
	if err != nil {
		return nil, err
	}

	if image.Size > 1<<20 {
		return nil, errors.New("image size too big")
	}

	allowedContentTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/jpg":  true,
	}

	file, err := image.Open()
	if err != nil {
		return nil, errors.New("unable to open image")
	}
	defer func(file multipart.File) {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close multipart file: %v\n", err)
		}
	}(file)

	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(file); err != nil {
		return nil, errors.New("unable to read image")
	}

	contentType := http.DetectContentType(buf.Bytes())
	if !allowedContentTypes[contentType] {
		return nil, errors.New("invalid image")
	}

	fileExtension := ".jpg"
	if contentType == "image/png" {
		fileExtension = ".png"
	}
	if contentType == "image/jpeg" {
		fileExtension = ".jpeg"
	}

	newFileName := fmt.Sprintf("%s%s", generateShortUUID(), fileExtension)

	return &domain.UploadInput{
		Object:      bytes.NewReader(buf.Bytes()),
		ObjectName:  newFileName,
		ObjectSize:  image.Size,
		ContentType: contentType,
	}, nil
}

func generateShortUUID() string {
	u := uuid.New()
	return u.String()[:12]
}
