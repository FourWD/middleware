package common

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func ZipFile(pathFile string, filename string) (string, error) {
	type fileMeta struct {
		Path  string
		IsDir bool
	}

	var files []fileMeta
	println(pathFile)
	err := filepath.Walk("/tmp/"+pathFile, func(path string, info os.FileInfo, err error) error {
		files = append(files, fileMeta{Path: path, IsDir: info.IsDir()})
		return nil
	})
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	z, err := os.Create("/tmp/" + filename + ".zip")
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	defer z.Close()

	zw := zip.NewWriter(z)
	defer zw.Close()

	for _, f := range files {
		path := f.Path
		if f.IsDir {
			path = fmt.Sprintf("%s%c", path, os.PathSeparator)
		}

		w, err := zw.Create(path)
		if err != nil {
			log.Fatalln(err)
			return "", err
		}

		if !f.IsDir {
			file, err := os.Open(f.Path)
			if err != nil {
				log.Fatalln(err)
				return "", err
			}
			defer file.Close()

			if _, err = io.Copy(w, file); err != nil {
				log.Fatalln(err)
				return "", err
			}
		}
	}

	return "/tmp/" + filename + ".zip", nil
}
