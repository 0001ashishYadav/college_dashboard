package utils

import (
	"context"
	"mime/multipart"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryUploader struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryUploader() (*CloudinaryUploader, error) {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return nil, err
	}

	return &CloudinaryUploader{cld: cld}, nil
}

func (u *CloudinaryUploader) UploadImage(
	ctx context.Context,
	file multipart.File,
	folder string,
) (string, error) {

	resp, err := u.cld.Upload.Upload(
		ctx,
		file,
		uploader.UploadParams{
			Folder:       folder,
			ResourceType: "image",
		},
	)
	if err != nil {
		return "", err
	}

	return resp.SecureURL, nil
}
