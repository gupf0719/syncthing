package qart

import (
	"fmt"
	"io/ioutil"
	"image"
	"image/color"
	"image/png"
	"os"
	"bytes"
	"github.com/vitrun/qart/qr"
)

// grayScale turn the image into white and black
func grayScale(src image.Image) *image.Gray{
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	gray := image.NewGray(bounds)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			oldColor := src.At(x, y)
			grayColor := color.GrayModel.Convert(oldColor)
			gray.Set(x, y, grayColor)
		}
	}
	return gray
}

// convert2PNG convert any format to PNG
func convert2PNG(i image.Image) bytes.Buffer{
	// Convert image to 128x128 gray+alpha.
//	i := grayScale(i)
	b := i.Bounds()

	const max = 128
	// If it's gigantic, it's more efficient to downsample first
	// and then resize; resizing will smooth out the roughness.
	var i1 *image.RGBA
	if b.Dx() > 4*max || b.Dy() > 4*max {
		w, h := 2*max, 2*max
		if b.Dx() > b.Dy() {
			h = b.Dy() * h / b.Dx()
		} else {
			w = b.Dx() * w / b.Dy()
		}
		i1 = qr.Resample(i, b, w, h)
	} else {
		// "Resample" to same size, just to convert to RGBA.
		i1 = qr.Resample(i, b, b.Dx(), b.Dy())
	}
	b = i1.Bounds()

	// Encode to PNG.
	dx, dy := 128, 128
	if b.Dx() > b.Dy() {
		dy = b.Dy() * dx / b.Dx()
	} else {
		dx = b.Dx() * dy / b.Dy()
	}
	i128 := qr.ResizeRGBA(i1, i1.Bounds(), dx, dy)

	var buf bytes.Buffer
	if err := png.Encode(&buf, i128); err != nil {
		panic(err)
	}
	return buf
}

//InitImage prepares the image
func InitImage(src []byte, seed int64, version, scale, mask, x, y int,
	randCtrl, dither, onlyData, saveCtrl bool) *Image{

	size, rotate := 0, 0
	if version > 8 {
		version = 8
	}
	if scale == 0 {
		scale = 8
	}
	if version >= 12 && scale >= 4 {
		scale /= 2
	}

	decodedImg, _, err := image.Decode(bytes.NewBuffer(src))
	if err != nil {
		return nil
	}

	buf := convert2PNG(decodedImg)
	target := makeTarg(buf.Bytes(), 17+4*version+size)

	img := &Image{
		Dx:           x,
		Dy:           y,
		URL:          "",
		Version:      version,
		Mask:         mask,
		RandControl:  randCtrl,
		Dither:       dither,
		OnlyDataBits: onlyData,
		SaveControl:  saveCtrl,
		Scale:        scale,
		Target:       target,
		Seed:         seed,
		Rotation:     rotate,
		Size:         size,
	}
	return img
}

// EncodeUrl encodes the url to the prepared image
func EncodeUrl(url string, img *Image) []byte {
	img.URL = url
	if err := img.Encode(); err != nil {
		fmt.Printf("error: %s\n", err)
		return nil
	}
	var dat []byte
	switch {
	case img.SaveControl:
		dat = img.Control
	default:
		dat = img.Code.PNG()
	}
	return dat
}

// Encode encodes a string with an image as the background
func Encode(url string, src []byte, seed int64, version, scale, mask, x, y int,
	randCtrl, dither, onlyData, saveCtrl bool) []byte {

	img := InitImage(src, seed, version, scale, mask, x, y, randCtrl, dither, onlyData, saveCtrl)
	return EncodeUrl(url, img)
}

// EncodeByFile encodes the given url with a specific image
func EncodeByFile(url, srcImg, dstImg string, version int) {
	data, err := ioutil.ReadFile(srcImg)
	if err != nil {
		fmt.Printf("err: %s\n", err)
		return
	}
	dst := Encode(url, data, 879633355, version, 4, 2, 4, 4, false, false, false, false)
	ioutil.WriteFile(dstImg, dst, (os.FileMode)(0644))
}
