package main

import (
	"io"
	"log"
	"os"

	"github.com/trenchesdeveloper/image-transformer/primitive"
)

func main() {
	f, err := os.Open("cover.png")

	if err != nil {
		log.Fatal("fail to open image file")
	}

	defer f.Close()

	output, err := primitive.Transform(f, 50)

	if err != nil {
		log.Fatal(err)
	}

	os.Remove("output.png")

	outFile, err := os.Create("output.png")

	if err != nil {
		log.Fatal(err)
	}

	defer outFile.Close()

	// Write the image data to the outFile file
	_, err = io.Copy(outFile, output)

	if err != nil {
		log.Fatal(err)
	}

}
