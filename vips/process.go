package vips

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/kumparan/imagor"
	"github.com/kumparan/imagor/imagorpath"
	"go.uber.org/zap"
)

var imageTypeMap = map[string]ImageType{
	"gif":    ImageTypeGIF,
	"jpeg":   ImageTypeJPEG,
	"jpg":    ImageTypeJPEG,
	"magick": ImageTypeMagick,
	"pdf":    ImageTypePDF,
	"png":    ImageTypePNG,
	"svg":    ImageTypeSVG,
	"tiff":   ImageTypeTIFF,
	"webp":   ImageTypeWEBP,
	"heif":   ImageTypeHEIF,
	"bmp":    ImageTypeBMP,
	"avif":   ImageTypeAVIF,
	"jp2":    ImageTypeJP2K,
}

// Process implements imagor.Processor interface
func (v *Processor) Process(
	ctx context.Context, blob *imagor.Blob, p imagorpath.Params, load imagor.LoadFunc,
) (*imagor.Blob, error) {
	ctx = withContext(ctx)
	defer contextDone(ctx)
	var (
		thumbnailNotSupported bool
		upscale               = true
		stretch               = p.Stretch
		thumbnail             = false
		stripExif             bool
		stripMetadata         = v.StripMetadata
		orient                int
		img                   *Image
		format                = ImageTypeUnknown
		maxN                  = v.MaxAnimationFrames
		maxBytes              int
		page                  = 1
		dpi                   = 0
		focalRects            []focal
		err                   error
	)
	if p.Trim {
		thumbnailNotSupported = true
	}
	if p.FitIn {
		upscale = false
	}
	if maxN == 0 || maxN < -1 {
		maxN = 1
	}
	if blob != nil && !blob.SupportsAnimation() {
		maxN = 1
	}
	for _, p := range p.Filters {
		if v.disableFilters[p.Name] {
			continue
		}
		switch p.Name {
		case "format":
			if imageType, ok := imageTypeMap[p.Args]; ok {
				format = supportedSaveFormat(imageType)
				if !IsAnimationSupported(format) {
					// no frames if export format not support animation
					maxN = 1
				}
			}
			break
		case "max_frames":
			if n, _ := strconv.Atoi(p.Args); n > 0 && (maxN == -1 || n < maxN) {
				maxN = n
			}
			break
		case "stretch":
			stretch = true
			break
		case "upscale":
			upscale = true
			break
		case "no_upscale":
			upscale = false
			break
		case "fill", "background_color":
			if args := strings.Split(p.Args, ","); args[0] == "auto" {
				thumbnailNotSupported = true
			}
			break
		case "page":
			if n, _ := strconv.Atoi(p.Args); n > 0 {
				page = n
			}
			break
		case "dpi":
			if n, _ := strconv.Atoi(p.Args); n > 0 {
				dpi = n
			}
			break
		case "orient":
			if n, _ := strconv.Atoi(p.Args); n > 0 {
				orient = n
				thumbnailNotSupported = true
			}
			break
		case "max_bytes":
			if n, _ := strconv.Atoi(p.Args); n > 0 {
				maxBytes = n
				thumbnailNotSupported = true
			}
			break
		case "trim", "focal", "rotate":
			thumbnailNotSupported = true
			break
		case "strip_exif":
			stripExif = true
		case "strip_metadata":
			stripMetadata = true
			break
		}
	}

	if !thumbnailNotSupported &&
		p.CropBottom == 0.0 && p.CropTop == 0.0 && p.CropLeft == 0.0 && p.CropRight == 0.0 {
		// apply shrink-on-load where possible
		if p.FitIn {
			if p.Width > 0 || p.Height > 0 {
				w := p.Width
				h := p.Height
				if w == 0 {
					w = v.MaxWidth
				}
				if h == 0 {
					h = v.MaxHeight
				}
				size := SizeDown
				if upscale {
					size = SizeBoth
				}
				if img, err = v.NewThumbnail(
					ctx, blob, w, h, InterestingNone, size, maxN, page, dpi,
				); err != nil {
					return nil, err
				}
				thumbnail = true
			}
		} else if stretch {
			if p.Width > 0 && p.Height > 0 {
				if img, err = v.NewThumbnail(
					ctx, blob, p.Width, p.Height,
					InterestingNone, SizeForce, maxN, page, dpi,
				); err != nil {
					return nil, err
				}
				thumbnail = true
			}
		} else {
			if p.Width > 0 && p.Height > 0 {
				interest := InterestingNone
				if p.Smart {
					interest = InterestingAttention
					thumbnail = true
				} else if (p.VAlign == imagorpath.VAlignTop && p.HAlign == "") ||
					(p.HAlign == imagorpath.HAlignLeft && p.VAlign == "") {
					interest = InterestingLow
					thumbnail = true
				} else if (p.VAlign == imagorpath.VAlignBottom && p.HAlign == "") ||
					(p.HAlign == imagorpath.HAlignRight && p.VAlign == "") {
					interest = InterestingHigh
					thumbnail = true
				} else if (p.VAlign == "" || p.VAlign == "middle") &&
					(p.HAlign == "" || p.HAlign == "center") {
					interest = InterestingCentre
					thumbnail = true
				}
				if thumbnail {
					if img, err = v.NewThumbnail(
						ctx, blob, p.Width, p.Height,
						interest, SizeBoth, maxN, page, dpi,
					); err != nil {
						return nil, err
					}
				}
			} else if p.Width > 0 && p.Height == 0 {
				if img, err = v.NewThumbnail(
					ctx, blob, p.Width, v.MaxHeight,
					InterestingNone, SizeBoth, maxN, page, dpi,
				); err != nil {
					return nil, err
				}
				thumbnail = true
			} else if p.Height > 0 && p.Width == 0 {
				if img, err = v.NewThumbnail(
					ctx, blob, v.MaxWidth, p.Height,
					InterestingNone, SizeBoth, maxN, page, dpi,
				); err != nil {
					return nil, err
				}
				thumbnail = true
			}
		}
	}
	if !thumbnail {
		if thumbnailNotSupported {
			if img, err = v.NewImage(ctx, blob, maxN, page, dpi); err != nil {
				return nil, err
			}
		} else {
			if img, err = v.NewThumbnail(
				ctx, blob, v.MaxWidth, v.MaxHeight,
				InterestingNone, SizeDown, maxN, page, dpi,
			); err != nil {
				return nil, err
			}
		}
	}
	// this should be called BEFORE vipscontext.contextDone
	defer img.Close()

	if orient > 0 {
		// orient rotate before resize
		if err = img.Rotate(getAngle(orient)); err != nil {
			return nil, err
		}
	}
	var (
		quality     int
		bitdepth    int
		compression int
		palette     bool
		origWidth   = float64(img.Width())
		origHeight  = float64(img.PageHeight())
	)
	if format == ImageTypeUnknown {
		if blob.BlobType() == imagor.BlobTypeAVIF {
			// meta loader determined as heif
			format = ImageTypeAVIF
		} else {
			format = img.Format()
		}
	}
	if v.Debug {
		v.Logger.Debug("image",
			zap.Int("width", img.Width()),
			zap.Int("height", img.Height()),
			zap.Int("page_height", img.PageHeight()))
	}
	for _, p := range p.Filters {
		if v.disableFilters[p.Name] {
			continue
		}
		switch p.Name {
		case "quality":
			quality, _ = strconv.Atoi(p.Args)
			break
		case "autojpg":
			format = ImageTypeJPEG
			break
		case "focal":
			args := strings.FieldsFunc(p.Args, argSplit)
			switch len(args) {
			case 4:
				f := focal{}
				f.Left, _ = strconv.ParseFloat(args[0], 64)
				f.Top, _ = strconv.ParseFloat(args[1], 64)
				f.Right, _ = strconv.ParseFloat(args[2], 64)
				f.Bottom, _ = strconv.ParseFloat(args[3], 64)
				if f.Left < 1 && f.Top < 1 && f.Right <= 1 && f.Bottom <= 1 {
					f.Left *= origWidth
					f.Right *= origWidth
					f.Top *= origHeight
					f.Bottom *= origHeight
				}
				if f.Right > f.Left && f.Bottom > f.Top {
					focalRects = append(focalRects, f)
				}
			case 2:
				f := focal{}
				f.Left, _ = strconv.ParseFloat(args[0], 64)
				f.Top, _ = strconv.ParseFloat(args[1], 64)
				if f.Left < 1 && f.Top < 1 {
					f.Left *= origWidth
					f.Top *= origHeight
				}
				f.Right = f.Left + 1
				f.Bottom = f.Top + 1
				focalRects = append(focalRects, f)
			}
			break
		case "palette":
			palette = true
			break
		case "bitdepth":
			bitdepth, _ = strconv.Atoi(p.Args)
			break
		case "compression":
			compression, _ = strconv.Atoi(p.Args)
			break
		}
	}
	if err := v.process(ctx, img, p, load, thumbnail, stretch, upscale, focalRects); err != nil {
		return nil, WrapErr(err)
	}
	if p.Meta {
		// metadata without export
		return imagor.NewBlobFromJsonMarshal(metadata(img, format, stripExif)), nil
	}
	format = supportedSaveFormat(format) // convert to supported export format
	for {
		buf, err := v.export(img, format, compression, quality, palette, bitdepth, stripMetadata)
		if err != nil {
			return nil, WrapErr(err)
		}
		if maxBytes > 0 && (quality > 10 || quality == 0) && format != ImageTypePNG {
			ln := len(buf)
			if v.Debug {
				v.Logger.Debug("max_bytes",
					zap.Int("bytes", ln),
					zap.Int("quality", quality),
				)
			}
			if ln > maxBytes {
				if quality == 0 {
					quality = 80
				}
				delta := float64(ln) / float64(maxBytes)
				switch {
				case delta > 3:
					quality = quality * 25 / 100
				case delta > 1.5:
					quality = quality * 50 / 100
				default:
					quality = quality * 75 / 100
				}
				if err := ctx.Err(); err != nil {
					return nil, WrapErr(err)
				}
				continue
			}
		}
		blob := imagor.NewBlobFromBytes(buf)
		if typ, ok := ImageMimeTypes[format]; ok {
			blob.SetContentType(typ)
		}
		return blob, nil
	}
}

func (v *Processor) process(
	ctx context.Context, img *Image, p imagorpath.Params, load imagor.LoadFunc, thumbnail, stretch, upscale bool, focalRects []focal,
) error {
	var (
		origWidth  = float64(img.Width())
		origHeight = float64(img.PageHeight())
		cropLeft,
		cropTop,
		cropRight,
		cropBottom float64
	)
	if p.CropRight > 0 || p.CropLeft > 0 || p.CropBottom > 0 || p.CropTop > 0 {
		// percentage
		cropLeft = math.Max(p.CropLeft, 0)
		cropTop = math.Max(p.CropTop, 0)
		cropRight = p.CropRight
		cropBottom = p.CropBottom
		if p.CropLeft < 1 && p.CropTop < 1 && p.CropRight <= 1 && p.CropBottom <= 1 {
			cropLeft = math.Round(cropLeft * origWidth)
			cropTop = math.Round(cropTop * origHeight)
			cropRight = math.Round(cropRight * origWidth)
			cropBottom = math.Round(cropBottom * origHeight)
		}
		if cropRight == 0 {
			cropRight = origWidth - 1
		}
		if cropBottom == 0 {
			cropBottom = origHeight - 1
		}
		cropRight = math.Min(cropRight, origWidth-1)
		cropBottom = math.Min(cropBottom, origHeight-1)
	}
	if p.Trim {
		if l, t, w, h, err := findTrim(ctx, img, p.TrimBy, p.TrimTolerance); err == nil {
			cropLeft = math.Max(cropLeft, float64(l))
			cropTop = math.Max(cropTop, float64(t))
			if cropRight > 0 {
				cropRight = math.Min(cropRight, float64(l+w))
			} else {
				cropRight = float64(l + w)
			}
			if cropBottom > 0 {
				cropBottom = math.Min(cropBottom, float64(t+h))
			} else {
				cropBottom = float64(t + h)
			}
		}
	}
	if cropRight > cropLeft && cropBottom > cropTop {
		if err := img.ExtractArea(
			int(cropLeft), int(cropTop), int(cropRight-cropLeft), int(cropBottom-cropTop),
		); err != nil {
			return err
		}
	}
	var (
		w = p.Width
		h = p.Height
	)
	if w == 0 && h == 0 {
		w = img.Width()
		h = img.PageHeight()
	} else if w == 0 {
		w = img.Width() * h / img.PageHeight()
		if !upscale && w > img.Width() {
			w = img.Width()
		}
	} else if h == 0 {
		h = img.PageHeight() * w / img.Width()
		if !upscale && h > img.PageHeight() {
			h = img.PageHeight()
		}
	}
	if !thumbnail {
		if p.FitIn {
			if upscale || w < img.Width() || h < img.PageHeight() {
				if err := img.Thumbnail(w, h, InterestingNone); err != nil {
					return err
				}
			}
		} else if stretch {
			if upscale || (w < img.Width() && h < img.PageHeight()) {
				if err := img.ThumbnailWithSize(
					w, h, InterestingNone, SizeForce,
				); err != nil {
					return err
				}
			}
		} else if upscale || w < img.Width() || h < img.PageHeight() {
			interest := InterestingCentre
			if p.Smart {
				interest = InterestingAttention
			} else if float64(w)/float64(h) > float64(img.Width())/float64(img.PageHeight()) {
				if p.VAlign == imagorpath.VAlignTop {
					interest = InterestingLow
				} else if p.VAlign == imagorpath.VAlignBottom {
					interest = InterestingHigh
				}
			} else {
				if p.HAlign == imagorpath.HAlignLeft {
					interest = InterestingLow
				} else if p.HAlign == imagorpath.HAlignRight {
					interest = InterestingHigh
				}
			}
			if len(focalRects) > 0 {
				focalX, focalY := parseFocalPoint(focalRects...)
				if err := v.FocalThumbnail(
					img, w, h,
					(focalX-cropLeft)/float64(img.Width()),
					(focalY-cropTop)/float64(img.PageHeight()),
				); err != nil {
					return err
				}
			} else {
				if err := v.Thumbnail(img, w, h, interest, SizeBoth); err != nil {
					return err
				}
			}
			if _, err := v.CheckResolution(img, nil); err != nil {
				return err
			}
		}
	}
	if p.HFlip {
		if err := img.Flip(DirectionHorizontal); err != nil {
			return err
		}
	}
	if p.VFlip {
		if err := img.Flip(DirectionVertical); err != nil {
			return err
		}
	}
	for i, filter := range p.Filters {
		if err := ctx.Err(); err != nil {
			return err
		}
		if v.disableFilters[filter.Name] {
			continue
		}
		if v.MaxFilterOps > 0 && i >= v.MaxFilterOps {
			if v.Debug {
				v.Logger.Debug("max-filter-ops-exceeded",
					zap.String("name", filter.Name), zap.String("args", filter.Args))
			}
			break
		}
		start := time.Now()
		var args []string
		if filter.Args != "" {
			args = strings.Split(filter.Args, ",")
		}
		if fn := v.Filters[filter.Name]; fn != nil {
			if err := fn(ctx, img, load, args...); err != nil {
				return err
			}
		} else if filter.Name == "fill" {
			if err := v.fill(ctx, img, w, h,
				p.PaddingLeft, p.PaddingTop, p.PaddingRight, p.PaddingBottom,
				filter.Args); err != nil {
				return err
			}
		}
		if v.Debug {
			v.Logger.Debug("filter",
				zap.String("name", filter.Name), zap.String("args", filter.Args),
				zap.Duration("took", time.Since(start)))
		}
	}
	return nil
}

// Metadata image attributes
type Metadata struct {
	Format      string         `json:"format"`
	ContentType string         `json:"content_type"`
	Width       int            `json:"width"`
	Height      int            `json:"height"`
	Orientation int            `json:"orientation"`
	Pages       int            `json:"pages"`
	Bands       int            `json:"bands"`
	Exif        map[string]any `json:"exif"`
}

func metadata(img *Image, format ImageType, stripExif bool) *Metadata {
	pages := 1
	if IsAnimationSupported(format) {
		pages = img.Height() / img.PageHeight()
	}
	if format == ImageTypePDF {
		pages = img.Pages()
	}
	exif := map[string]any{}
	if !stripExif {
		exif = img.Exif()
	}
	return &Metadata{
		Format:      ImageTypes[format],
		ContentType: ImageMimeTypes[format],
		Width:       img.Width(),
		Height:      img.PageHeight(),
		Pages:       pages,
		Bands:       img.Bands(),
		Orientation: img.Orientation(),
		Exif:        exif,
	}
}

func supportedSaveFormat(format ImageType) ImageType {
	switch format {
	case ImageTypePNG, ImageTypeWEBP, ImageTypeTIFF, ImageTypeGIF, ImageTypeAVIF, ImageTypeHEIF, ImageTypeJP2K:
		if IsSaveSupported(format) {
			return format
		}
		if format == ImageTypeAVIF && IsSaveSupported(ImageTypeHEIF) {
			return ImageTypeAVIF
		}
	}
	return ImageTypeJPEG
}

func (v *Processor) export(
	image *Image, format ImageType, compression int, quality int, palette bool, bitdepth int, stripMetadata bool,
) ([]byte, error) {
	switch format {
	case ImageTypePNG:
		opts := NewPngExportParams()
		if quality > 0 {
			opts.Quality = quality
		}
		if palette {
			opts.Palette = palette
		}
		if bitdepth > 0 {
			opts.Bitdepth = bitdepth
		}
		if compression > 0 {
			opts.Compression = compression
		}
		if stripMetadata {
			opts.StripMetadata = true
		}
		return image.ExportPng(opts)
	case ImageTypeWEBP:
		opts := NewWebpExportParams()
		if quality > 0 {
			opts.Quality = quality
		}
		if stripMetadata {
			opts.StripMetadata = true
		}
		return image.ExportWebp(opts)
	case ImageTypeTIFF:
		opts := NewTiffExportParams()
		if quality > 0 {
			opts.Quality = quality
		}
		if stripMetadata {
			opts.StripMetadata = true
		}
		return image.ExportTiff(opts)
	case ImageTypeGIF:
		opts := NewGifExportParams()
		if quality > 0 {
			opts.Quality = quality
		}
		if stripMetadata {
			opts.StripMetadata = true
		}
		return image.ExportGIF(opts)
	case ImageTypeAVIF:
		opts := NewAvifExportParams()
		if quality > 0 {
			opts.Quality = quality
		}
		if stripMetadata {
			opts.StripMetadata = true
		}
		opts.Speed = v.AvifSpeed
		return image.ExportAvif(opts)
	case ImageTypeHEIF:
		opts := NewHeifExportParams()
		if quality > 0 {
			opts.Quality = quality
		}
		return image.ExportHeif(opts)
	case ImageTypeJP2K:
		opts := NewJp2kExportParams()
		if quality > 0 {
			opts.Quality = quality
		}
		return image.ExportJp2k(opts)
	default:
		opts := NewJpegExportParams()
		if v.MozJPEG {
			opts.Quality = 75
			opts.StripMetadata = true
			opts.OptimizeCoding = true
			opts.Interlace = true
			opts.OptimizeScans = true
			opts.TrellisQuant = true
			opts.QuantTable = 3
		}
		if quality > 0 {
			opts.Quality = quality
		}
		if stripMetadata {
			opts.StripMetadata = true
		}
		return image.ExportJpeg(opts)
	}
}

func argSplit(r rune) bool {
	return r == 'x' || r == ',' || r == ':'
}

type focal struct {
	Left   float64
	Right  float64
	Top    float64
	Bottom float64
}

func parseFocalPoint(focalRects ...focal) (focalX, focalY float64) {
	var sumWeight float64
	for _, f := range focalRects {
		sumWeight += (f.Right - f.Left) * (f.Bottom - f.Top)
	}
	for _, f := range focalRects {
		r := (f.Right - f.Left) * (f.Bottom - f.Top) / sumWeight
		focalX += (f.Left + f.Right) / 2 * r
		focalY += (f.Top + f.Bottom) / 2 * r
	}
	return
}

func findTrim(
	_ context.Context, img *Image, pos string, tolerance int,
) (l, t, w, h int, err error) {
	if isAnimated(img) {
		// skip animation support
		return
	}
	var x, y int
	if pos == imagorpath.TrimByBottomRight {
		x = img.Width() - 1
		y = img.PageHeight() - 1
	}
	if tolerance == 0 {
		tolerance = 1
	}
	l, t, w, h, err = img.FindTrim(float64(tolerance), x, y)
	return
}
