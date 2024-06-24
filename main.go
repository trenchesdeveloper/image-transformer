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
	"strconv"

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

	mux.HandleFunc("/modify/", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./img/" + filepath.Base(r.URL.Path))

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer f.Close()

		ext := filepath.Ext(f.Name())[1:]

		modeStr := r.URL.Query().Get("mode")

		fmt.Println("getting mode", modeStr)

		if modeStr == "" {
			renderModeChoices(w, r, f, ext)
			return
		}

		mode, err := strconv.Atoi(modeStr)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		numShapes := r.URL.Query().Get("n")

		fmt.Println("getting numShapes", numShapes)

		if numShapes == "" {
			renderNumShapesChoices(w, r, f, ext, primitive.Mode(mode))
			return
		}

		n, err := strconv.Atoi(numShapes)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_ = n

		http.Redirect(w, r, "/img/"+filepath.Base(f.Name()), http.StatusFound)
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
		onDisk, err := tempfile("", ext)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer onDisk.Close()

		_, err = io.Copy(onDisk, file)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/modify/"+filepath.Base(onDisk.Name()), http.StatusFound)

	})

	mux.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))

	log.Fatal(http.ListenAndServe(":3000", mux))

}

func renderNumShapesChoices(w http.ResponseWriter, r *http.Request, rs io.ReadSeeker, ext string, mode primitive.Mode) {
	opts := []genOpts{
		{N: 10, Mode: mode},
		{N: 50, Mode: mode},
		{N: 100, Mode: mode},
		{N: 200, Mode: mode},
		{N: 500, Mode: mode},
	}

	imgs, err := genImages(rs, ext, opts...)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := `<html><body>
			{{range .}}
			<a href="/img/{{.Name}}?mode={{.Mode}}&n={{.NumShapes}}">
				<img style="width: 20%" src="/img/{{.Name}}" />
			</a>
			{{end}}
			</body></html>`
	templ := template.Must(template.New("").Parse(html))

	type dataStruct struct {
		Name      string
		Mode      primitive.Mode
		NumShapes int
	}

	var data []dataStruct

	for i, img := range imgs {
		data = append(data, dataStruct{
			Name:      filepath.Base(img),
			Mode:      opts[i].Mode,
			NumShapes: opts[i].N,
		})
	}

	templ.Execute(w, data)
}

func renderModeChoices(w http.ResponseWriter, r *http.Request, rs io.ReadSeeker, ext string) {
	opts := []genOpts{
		{N: 100, Mode: primitive.ModeTriangle},
		{N: 100, Mode: primitive.ModeRect},
		{N: 100, Mode: primitive.ModeEllipse},
		{N: 100, Mode: primitive.ModeCircle},
		{N: 100, Mode: primitive.ModeRotatedRect},
		{N: 100, Mode: primitive.ModeBeziers},
		{N: 100, Mode: primitive.ModeRotatedEllipse},
	}

	imgs, err := genImages(rs, ext, opts...)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := `<html><body>
			{{range .}}
			<a href="/modify/{{.Name}}?mode={{.Mode}}">
				<img style="width: 20%" src="/img/{{.Name}}" />
			</a>
			{{end}}
			</body></html>`
	templ := template.Must(template.New("").Parse(html))

	type dataStruct struct {
		Name string
		Mode primitive.Mode
	}

	var data []dataStruct

	for i, img := range imgs {
		data = append(data, dataStruct{
			Name: filepath.Base(img),
			Mode: opts[i].Mode,
		})

	}

	templ.Execute(w, data)
}

type genOpts struct {
	N    int
	Mode primitive.Mode
}

func genImages(rs io.ReadSeeker, ext string, opts ...genOpts) ([]string, error) {
	var out []string

	for _, opt := range opts {
		rs.Seek(0, 0)
		outFile, err := genImage(rs, ext, opt.N, opt.Mode)
		if err != nil {
			return nil, err
		}

		out = append(out, outFile)
	}

	return out, nil
}
func genImage(file io.Reader, ext string, numShapes int, mode primitive.Mode) (string, error) {
	output, err := primitive.Transform(file, ext, numShapes, primitive.WithMode(mode))

	if err != nil {
		return "", err

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
