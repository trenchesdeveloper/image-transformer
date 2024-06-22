package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/trenchesdeveloper/image-transformer/primitive"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body>
			<form action="/upload" method="post" enctype="multipart/form-data">
				<input type="file" name="image">
				<button type="submit">Upload Image</button>
			</form>
		</body></html>`

		w.Write([]byte(html))

	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form in the request
		err := r.ParseMultipartForm(10 << 20) // 10 MB

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get a reference to the file
		file, header, err := r.FormFile("image")

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()

		ext := filepath.Ext(header.Filename)[1:]
		_ = ext

		a, err := genImage(file, ext, 100, primitive.ModeTriangle)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file.Seek(0, 0)

		b, err := genImage(file, ext, 100, primitive.ModeRect)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file.Seek(0, 0)

		c, err := genImage(file, ext, 100, primitive.ModeEllipse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file.Seek(0, 0)

		d, err := genImage(file, ext, 100, primitive.ModeCircle)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file.Seek(0, 0)
		e, err := genImage(file, ext, 100, primitive.ModeRotatedRect)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file.Seek(0, 0)
		f, err := genImage(file, ext, 100, primitive.ModeBeziers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file.Seek(0, 0)
		g, err := genImage(file, ext, 100, primitive.ModeRotatedEllipse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		html := `<html><body>
			{{range .}}
				<img src="/{{.}}" />
			{{end}}
			</body></html>`
		templ := template.Must(template.New("").Parse(html))

		images := []string{a, b, c, d, e, f, g}

		templ.Execute(w, images)

	})

	mux.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))

	log.Fatal(http.ListenAndServe(":3000", mux))

}

func genImage(file io.Reader, ext string, numShapes int, mode primitive.Mode) (string, error) {
	output, err := primitive.Transform(file, ext, numShapes, primitive.WithMode(mode))

		if err != nil {
			return "",err

		}

		outFile, err := tempfile("", ext)

		if err != nil {

			return "", err
		}

		defer outFile.Close()

		io.Copy(outFile, output)

		return outFile.Name(), nil
}

func tempfile(prefix, ext string) (*os.File, error) {
	in, err := os.CreateTemp("./img/", prefix)

	if err != nil {
		return nil, errors.New("main: failed to create temporary file")
	}

	defer os.Remove(in.Name())

	return os.Create(fmt.Sprintf("%s.%s", in.Name(), ext))
}
