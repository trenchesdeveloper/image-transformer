package main

import (
	"errors"
	"fmt"
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
		output, err := primitive.Transform(file, ext, 50, primitive.WithMode(primitive.ModeCombo))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		outFile, err := tempfile("", ext)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer outFile.Close()

		io.Copy(outFile, output)

		redirectUrl := fmt.Sprintf("/img/%s", filepath.Base(outFile.Name()))

		http.Redirect(w, r, redirectUrl, http.StatusFound)

	})

	mux.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))

	log.Fatal(http.ListenAndServe(":3000", mux))

}

func tempfile(prefix, ext string) (*os.File, error) {
	in, err := os.CreateTemp("./img/", prefix)

	if err != nil {
		return nil, errors.New("main: failed to create temporary file")
	}

	defer os.Remove(in.Name())

	return os.Create(fmt.Sprintf("%s.%s", in.Name(), ext))
}
