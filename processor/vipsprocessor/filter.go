package vipsprocessor

import (
	"fmt"
	"github.com/cshum/govips/v2/vips"
	"github.com/cshum/imagor"
	"golang.org/x/image/colornames"
	"image/color"
	"net/url"
	"strconv"
	"strings"
)

func trim(img *vips.ImageRef, pos string, tolerance int) error {
	var x, y int
	if pos == "bottom-right" {
		x = img.Width() - 1
		y = img.Height() - 1
	}
	if tolerance == 0 {
		tolerance = 1
	}
	p, err := img.GetPoint(x, y)
	if err != nil {
		return err
	}
	l, t, w, h, err := img.FindTrim(float64(tolerance), &vips.Color{
		R: uint8(p[0]), G: uint8(p[1]), B: uint8(p[2]),
	})
	if err != nil {
		return err
	}
	if err = img.ExtractArea(l, t, w, h); err != nil {
		return err
	}
	return nil
}

func fill(img *vips.ImageRef, w, h int, fill string, upscale bool) (err error) {
	fill = strings.ToLower(fill)
	if img.HasAlpha() && fill != "blur" {
		if err = img.Flatten(getColor(fill)); err != nil {
			return
		}
	}
	if fill == "black" {
		if err = img.Embed(
			(w-img.Width())/2, (h-img.Height())/2,
			w, h, vips.ExtendBlack,
		); err != nil {
			return
		}
	} else if fill == "white" {
		if err = img.Embed(
			(w-img.Width())/2, (h-img.Height())/2,
			w, h, vips.ExtendWhite,
		); err != nil {
			return
		}
	} else {
		var cp *vips.ImageRef
		if cp, err = img.Copy(); err != nil {
			return
		}
		defer cp.Close()
		if upscale || w < cp.Width() || h < cp.Height() {
			if err = cp.Thumbnail(w, h, vips.InterestingNone); err != nil {
				return
			}
		}
		if err = img.ResizeWithVScale(
			float64(w)/float64(img.Width()), float64(h)/float64(img.Height()),
			vips.KernelLinear,
		); err != nil {
			return
		}
		if fill == "blur" {
			if err = img.GaussianBlur(50); err != nil {
				return
			}
		} else {
			c := getColor(fill)
			if err = img.DrawRect(vips.ColorRGBA{
				R: c.R, G: c.G, B: c.B, A: 255,
			}, 0, 0, w, h, true); err != nil {
				return
			}
		}
		if err = img.Composite(
			cp, vips.BlendModeOver, (w-cp.Width())/2, (h-cp.Height())/2); err != nil {
			return
		}
	}
	return
}

func watermark(img *vips.ImageRef, load imagor.LoadFunc, args ...string) (err error) {
	ln := len(args)
	if ln < 1 {
		return
	}
	image := args[0]
	if unescape, e := url.QueryUnescape(args[0]); e == nil {
		image = unescape
	}
	var buf []byte
	if buf, err = load(image); err != nil {
		return
	}
	var overlay *vips.ImageRef
	if overlay, err = vips.NewImageFromBuffer(buf); err != nil {
		return
	}
	defer overlay.Close()
	var x, y, w, h int

	// w_ratio h_ratio
	if ln >= 6 {
		w = img.Width()
		h = img.Height()
		if args[4] != "none" {
			w, _ = strconv.Atoi(args[4])
			w = img.Width() * w / 100
		}
		if args[5] != "none" {
			h, _ = strconv.Atoi(args[5])
			h = img.Height() * h / 100
		}
		if w < overlay.Width() || h < overlay.Height() {
			if err = overlay.Thumbnail(w, h, vips.InterestingNone); err != nil {
				return
			}
		}
	}
	// alpha
	if ln >= 4 {
		alpha, _ := strconv.ParseFloat(args[3], 64)
		alpha = 1 - alpha/100
		if err = overlay.AddAlpha(); err != nil {
			return
		}
		if err = overlay.Linear([]float64{1, 1, 1, alpha}, []float64{0, 0, 0, 0}); err != nil {
			return
		}
	}
	// x y
	if ln >= 3 {
		if args[1] == "center" {
			x = (img.Width() - overlay.Width()) / 2
		} else if strings.HasSuffix(args[1], "p") {
			x, _ = strconv.Atoi(strings.TrimSuffix(args[1], "p"))
			x = x * img.Width() / 100
		} else {
			x, _ = strconv.Atoi(args[1])
		}
		if args[2] == "center" {
			y = (img.Height() - overlay.Height()) / 2
		} else if strings.HasSuffix(args[2], "p") {
			y, _ = strconv.Atoi(strings.TrimSuffix(args[2], "p"))
			y = y * img.Height() / 100
		} else {
			y, _ = strconv.Atoi(args[2])
		}
		if x < 0 {
			x += img.Width() - overlay.Width()
		}
		if y < 0 {
			y += img.Height() - overlay.Height()
		}
	}
	if err = img.Composite(overlay, vips.BlendModeOver, x, y); err != nil {
		return
	}
	return
}

func roundCorner(img *vips.ImageRef, _ imagor.LoadFunc, args ...string) (err error) {
	var rx, ry int
	if len(args) == 0 {
		return
	}
	rx, _ = strconv.Atoi(args[0])
	ry = rx
	if len(args) > 1 {
		rx, _ = strconv.Atoi(args[1])
	}

	var rect *vips.ImageRef
	var w = img.Width()
	var h = img.Height()
	if rect, err = vips.NewImageFromBuffer([]byte(fmt.Sprintf(`
		<svg viewBox="0 0 %d %d">
			<rect rx="%d" ry="%d" 
			 x="0" y="0" width="%d" height="%d" 
			 fill="#fff"/>
		</svg>
	`, w, h, rx, ry, w, h))); err != nil {
		return
	}
	defer rect.Close()
	if err = img.Composite(rect, vips.BlendModeDestIn, 0, 0); err != nil {
		return
	}
	return nil
}

func rotate(img *vips.ImageRef, _ imagor.LoadFunc, args ...string) (err error) {
	if len(args) == 0 {
		return
	}
	if angle, _ := strconv.Atoi(args[0]); angle > 0 {
		vAngle := vips.Angle0
		switch angle {
		case 90:
			vAngle = vips.Angle270
		case 180:
			vAngle = vips.Angle180
		case 270:
			vAngle = vips.Angle90
		}
		if err = img.Rotate(vAngle); err != nil {
			return err
		}
	}
	return
}

func grayscale(img *vips.ImageRef, _ imagor.LoadFunc, _ ...string) (err error) {
	return img.Modulate(1, 0, 0)
}

func brightness(img *vips.ImageRef, _ imagor.LoadFunc, args ...string) (err error) {
	if len(args) == 0 {
		return
	}
	b, _ := strconv.ParseFloat(args[0], 64)
	b = b * 256 / 100
	return img.Linear([]float64{1, 1, 1}, []float64{b, b, b})
}

func contrast(img *vips.ImageRef, _ imagor.LoadFunc, args ...string) (err error) {
	if len(args) == 0 {
		return
	}
	a, _ := strconv.ParseFloat(args[0], 64)
	a = a * 256 / 100
	b := 128 - a*128
	return img.Linear([]float64{a, a, a}, []float64{b, b, b})
}

func hue(img *vips.ImageRef, _ imagor.LoadFunc, args ...string) (err error) {
	if len(args) == 0 {
		return
	}
	h, _ := strconv.ParseFloat(args[0], 64)
	return img.Modulate(1, 1, h)
}

func saturation(img *vips.ImageRef, _ imagor.LoadFunc, args ...string) (err error) {
	if len(args) == 0 {
		return
	}
	s, _ := strconv.ParseFloat(args[0], 64)
	s = 1 + s/100
	return img.Modulate(1, s, 0)
}

func rgb(img *vips.ImageRef, _ imagor.LoadFunc, args ...string) (err error) {
	if len(args) != 3 {
		return
	}
	r, _ := strconv.ParseFloat(args[0], 64)
	g, _ := strconv.ParseFloat(args[1], 64)
	b, _ := strconv.ParseFloat(args[2], 64)
	r = r * 256 / 100
	g = g * 256 / 100
	b = b * 256 / 100
	return img.Linear([]float64{1, 1, 1}, []float64{r, g, b})
}

func blur(img *vips.ImageRef, _ imagor.LoadFunc, args ...string) (err error) {
	var sigma float64
	switch len(args) {
	case 2:
		sigma, _ = strconv.ParseFloat(args[1], 64)
		break
	case 1:
		sigma, _ = strconv.ParseFloat(args[0], 64)
		break
	}
	sigma /= 2
	if sigma > 0 {
		return img.GaussianBlur(sigma)
	}
	return
}

func sharpen(img *vips.ImageRef, _ imagor.LoadFunc, args ...string) (err error) {
	var sigma float64
	switch len(args) {
	case 1:
		sigma, _ = strconv.ParseFloat(args[0], 64)
		break
	case 2, 3:
		sigma, _ = strconv.ParseFloat(args[1], 64)
		break
	}
	sigma = 1 + sigma*2
	return img.Sharpen(sigma, 1, 2)
}

func stripIcc(img *vips.ImageRef, _ imagor.LoadFunc, _ ...string) (err error) {
	return img.RemoveICCProfile()
}

func stripExif(img *vips.ImageRef, _ imagor.LoadFunc, _ ...string) (err error) {
	return img.RemoveICCProfile()
}

func getColor(name string) *vips.Color {
	vc := &vips.Color{}
	strings.TrimPrefix(strings.ToLower(name), "#")
	if c, ok := colornames.Map[strings.ToLower(name)]; ok {
		vc.R = c.R
		vc.G = c.G
		vc.B = c.B
	} else if c, ok := parseHexColor(name); ok {
		vc.R = c.R
		vc.G = c.G
		vc.B = c.B
	}
	return vc
}

func parseHexColor(s string) (c color.RGBA, ok bool) {
	c.A = 0xff
	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		return 0
	}
	switch len(s) {
	case 6:
		c.R = hexToByte(s[0])<<4 + hexToByte(s[1])
		c.G = hexToByte(s[2])<<4 + hexToByte(s[3])
		c.B = hexToByte(s[4])<<4 + hexToByte(s[5])
		ok = true
	case 3:
		c.R = hexToByte(s[0]) * 17
		c.G = hexToByte(s[1]) * 17
		c.B = hexToByte(s[2]) * 17
		ok = true
	}
	return
}