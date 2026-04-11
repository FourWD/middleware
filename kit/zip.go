package kit

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func ZipFile(pathFile, filename string) (string, error) {
	archivePath := filepath.Join(os.TempDir(), filename+".zip")

	archive, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	root := filepath.Join(os.TempDir(), pathFile)
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(filepath.Dir(root), path)
		if err != nil {
			return err
		}

		headerName := filepath.ToSlash(relativePath)
		if info.IsDir() {
			_, err = zipWriter.Create(headerName + "/")
			return err
		}

		writer, err := zipWriter.Create(headerName)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		return "", err
	}

	return archivePath, nil
}
