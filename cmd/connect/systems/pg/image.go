package pg

import (
	"context"

	b64 "encoding/base64"
)

type imageRaw struct {
	image  []byte
	base64 string
}

func newRawImage(image []byte) imageRaw {
	img := imageRaw{
		image: image,
	}

	return img
}

// EncodeBase64 reads the specified image and converts the image to a base64 string.
func (img imageRaw) EncodeBase64(ctx context.Context) (string, error) {
	if img.base64 != "" {
		return img.base64, nil
	}

	img.base64 = b64.StdEncoding.EncodeToString(img.image)

	return img.base64, nil
}
