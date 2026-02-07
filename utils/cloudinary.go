package utils

import (
	"context"
	"io"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadImageStream(
	ctx context.Context,
	file io.Reader,
	folder string,
) (string, string, error) {

	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return "", "", err
	}

	resp, err := cld.Upload.Upload(
		ctx,
		file,
		uploader.UploadParams{
			Folder: folder,
		},
	)
	if err != nil {
		return "", "", err
	}

	return resp.SecureURL, resp.PublicID, nil
}

func DeleteImage(ctx context.Context, publicID string) error {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return err
	}

	_, err = cld.Upload.Destroy(
		ctx,
		uploader.DestroyParams{
			PublicID: publicID,
		},
	)
	return err
}
