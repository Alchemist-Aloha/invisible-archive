package main

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func createTestImage(format string) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var c color.Color
	switch format {
	case "png":
		c = color.RGBA{255, 0, 0, 255} // Red
	case "jpeg":
		c = color.RGBA{0, 255, 0, 255} // Green
	default:
		c = color.RGBA{0, 0, 255, 255} // Blue
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)

	buf := new(bytes.Buffer)
	if format == "png" {
		_ = png.Encode(buf, img)
	} else {
		_ = jpeg.Encode(buf, img, nil)
	}
	return buf.Bytes()
}

func createZip(path string, files map[string][]byte, explicitDirs []string) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for _, dir := range explicitDirs {
		_, _ = zw.Create(dir + "/")
	}

	for name, content := range files {
		f, err := zw.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		f.Write(content)
	}

	zw.Close()
	os.WriteFile(path, buf.Bytes(), 0644)
}

func main() {
	base := "test_library"
	_ = os.MkdirAll(base, 0755)

	pngData := createTestImage("png")
	jpgData := createTestImage("jpeg")

	// 1. Media ZIP with real images
	createZip(filepath.Join(base, "photos.zip"), map[string][]byte{
		"red_square.png":   pngData,
		"green_square.jpg": jpgData,
		"vacation/blue.jpg": jpgData,
	}, []string{"vacation"})

	// 2. Mixed content ZIP
	createZip(filepath.Join(base, "mixed.zip"), map[string][]byte{
		"note.txt":      []byte("Just a text file"),
		"thumbnail.png": pngData,
	}, nil)

	log.Printf("Generated test library with real images in %s/", base)
}
