package primitive

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Mode int

// Mode are the different shapes that can be used by the primitive package
const (
	ModeCombo Mode = iota
	ModeTriangle
	ModeRect
	ModeEllipse
	ModeCircle
	ModeRotatedRect
	ModeBeziers
	ModeRotatedEllipse
	ModePolygon
)

// WithMode is an option for the Transform function that will define
// the mode you want to use. By default, ModeTraiangle wi;ll be used.
func WithMode(mode Mode) func() []string {
	return func() []string {
		return []string{"-m", fmt.Sprint(mode)}
	}
}

// Transform will take the provided image and apply the primitive algorithm
// to it, then return the resulting image.
func Transform(image io.Reader, ext string, numShapes int, opts ...func() []string) (io.Reader, error) {
	var args []string

	for _, opt := range opts {
		args = append(args, opt()...)
	}
	inputFile, err := tempfile("input_", ext)
	if err != nil {
		return nil, err
	}

	outputFile, err := tempfile("input_", ext)
	if err != nil {
		return nil, err
	}

	defer os.Remove(outputFile.Name())

	// Write the image data to the input file
	_, err = io.Copy(inputFile, image)

	if err != nil {
		return nil, err
	}

	// Run the primitive command
	stdCombo, err := primitive(inputFile.Name(), outputFile.Name(), numShapes, args...)

	if err != nil {
		return nil, err
	}

	fmt.Println(stdCombo)

	// read out into a buffer
	outputBuffer := new(bytes.Buffer)

	_, err = io.Copy(outputBuffer, outputFile)

	if err != nil {
		return nil, err
	}

	return outputBuffer, nil
}

func primitive(inputFile, outputFile string, numberOfShapes int,args ...string) (string, error) {
	argStr := fmt.Sprintf("-i %s -o %s -n %d", inputFile, outputFile, numberOfShapes)

	args = append(args, strings.Fields(argStr)...)
	cmd := exec.Command("primitive", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", err
	}

	return string(output), nil
}


func tempfile(prefix, ext string) (*os.File, error) {
	in, err := os.CreateTemp("", prefix)

	if err != nil {
		return nil, err

	}

	defer os.Remove(in.Name())

	return os.Create(fmt.Sprintf("%s.%s", in.Name(), ext))
}