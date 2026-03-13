package main
import (
	"fmt"
	"mime"
)
func main() {
	fmt.Printf("m4v: %s\n", mime.TypeByExtension(".m4v"))
	fmt.Printf("mp4: %s\n", mime.TypeByExtension(".mp4"))
}
