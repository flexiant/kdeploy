package utils

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// CheckRequiredFlags checks for required flags, and show usage if requirements not met
func CheckRequiredFlags(c *cli.Context, flags []string) {
	missing := ""
	for _, flag := range flags {
		if c.String(flag) != "" {
			missing = fmt.Sprintf("%s\t--%s\n", missing, flag)
		}
	}

	if missing != "" {
		fmt.Printf("Incorrect usage. Please use parameters:\n%s\n", missing)
		cli.ShowCommandHelp(c, c.Command.Name)
		os.Exit(2)
	}
}

func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func CheckError(err error) {
	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			log.Fatal(err)
		} else {
			log.Fatal(err)
		}
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

// Keys returns the keys in a map[string]string
func Keys(m map[string]string) []string {
	names := make([]string, len(m))
	i := 0
	for key := range m {
		names[i] = key
		i++
	}
	return names
}

// Values returns the values in a map[string]string
func Values(m map[string]string) []string {
	vals := make([]string, len(m))
	i := 0
	for _, val := range m {
		vals[i] = val
		i++
	}
	return vals
}
