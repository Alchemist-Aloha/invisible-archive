package main

import (
	"archive/zip"
	"os"
)

func main() {
	f, err := os.Create("test_library/large.zip")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	zf, err := zw.Create("large_video.mp4")
	if err != nil {
		panic(err)
	}

	// Write 10MB of data
	data := make([]byte, 10*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	zf.Write(data)

	zw.Close()
}
