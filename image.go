package imagor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/kumparan/bimg"
	"strings"
)

// Image stores an image binary buffer and its MIME type
type Image struct {
	Body []byte
	Mime string
}

func Process(buf []byte, opts bimg.Options) (out Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch value := r.(type) {
			case error:
				err = value
			case string:
				err = errors.New(value)
			default:
				err = errors.New("libvips internal error")
			}
			out = Image{}
		}
	}()

	if opts.Type == bimg.GIF || (bimg.DetermineImageType(buf) == bimg.GIF && opts.Type == bimg.UNKNOWN) {
		return ProcessGIF(buf, opts)
	}

	// Resize image via bimg
	ibuf, err := bimg.Resize(buf, opts)

	// Handle specific type encode errors gracefully
	if err != nil && strings.Contains(err.Error(), "encode") && (opts.Type == bimg.WEBP || opts.Type == bimg.HEIF) {
		// Always fallback to JPEG
		opts.Type = bimg.JPEG
		ibuf, err = bimg.Resize(buf, opts)
	}

	if err != nil {
		return Image{}, err
	}

	mime := GetImageMimeType(bimg.DetermineImageType(ibuf))
	return Image{Body: ibuf, Mime: mime}, nil
}

// GetImageMimeType returns the MIME type based on the given image type code.
func GetImageMimeType(code bimg.ImageType) string {
	switch code {
	case bimg.PNG:
		return "image/png"
	case bimg.WEBP:
		return "image/webp"
	case bimg.AVIF:
		return "image/avif"
	case bimg.TIFF:
		return "image/tiff"
	case bimg.GIF:
		return "image/gif"
	case bimg.SVG:
		return "image/svg+xml"
	case bimg.PDF:
		return "application/pdf"
	case bimg.JXL:
		return "image/jxl"
	default:
		return "image/jpeg"
	}
}

func readJSONBodyData(data []byte) ([]byte, error) {
	type supportedJSONField struct {
		Base64 string `json:"base64"`
	}

	jsonField := new(supportedJSONField)
	if err := json.Unmarshal(data, jsonField); err != nil {
		return nil, err
	}

	if jsonField.Base64 != "" {
		base64Str := jsonField.Base64
		base64Split := strings.Split(jsonField.Base64, "base64,")
		if len(base64Split) > 1 {
			base64Str = base64Split[1]
		}
		return base64.StdEncoding.DecodeString(base64Str)
	}

	return nil, ErrEmptyBody
}
