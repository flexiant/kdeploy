package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	CheckError(err)
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		CheckError(err)
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			CheckError(err)
			f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			CheckError(err)
			defer f.Close()

			_, err = io.Copy(f, rc)
			CheckError(err)
		}
	}
	return nil
}
