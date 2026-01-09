package assets

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/lenaxia/tinyrsvp/internal/storage"
)

type ImageService interface {
	UploadImage(ctx context.Context, eventID int64, filename string, data io.Reader) (*ImageMetadata, error)
	DeleteImage(ctx context.Context, path string) error
	GetImageURL(ctx context.Context, path string) (string, error)
}

type ImageMetadata struct {
	Path        string `json:"path"`
	PublicURL   string `json:"public_url"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

type imageService struct {
	provider  storage.Provider
	validator *ImageValidator
}

func NewImageService(provider storage.Provider) ImageService {
	return &imageService{
		provider:  provider,
		validator: NewImageValidator(),
	}
}

func (s *imageService) UploadImage(ctx context.Context, eventID int64, filename string, data io.Reader) (*ImageMetadata, error) {
	imageData, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	result, err := s.validator.Validate(imageData)
	if err != nil {
		return nil, err
	}

	cleanData, err := stripEXIF(imageData, result.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	uniqueFilename := generateUniqueFilename(filename)
	path := fmt.Sprintf("images/%d/%s", eventID, uniqueFilename)

	if err := s.provider.PutObject(ctx, path, bytes.NewReader(cleanData), result.ContentType); err != nil {
		return nil, fmt.Errorf("failed to store image: %w", err)
	}

	publicURL, err := s.provider.GetPublicURL(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get public URL: %w", err)
	}

	return &ImageMetadata{
		Path:        path,
		PublicURL:   publicURL,
		Filename:    uniqueFilename,
		ContentType: result.ContentType,
		Size:        int64(len(cleanData)),
		Width:       result.Width,
		Height:      result.Height,
	}, nil
}

func (s *imageService) DeleteImage(ctx context.Context, path string) error {
	return s.provider.DeleteObject(ctx, path)
}

func (s *imageService) GetImageURL(ctx context.Context, path string) (string, error) {
	return s.provider.GetPublicURL(ctx, path)
}
