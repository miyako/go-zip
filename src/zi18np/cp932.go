package zi18np

import (
	"archive/zip"
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Sjis2Utf8(str string) (string, error) {
	iostr := strings.NewReader(str)
	rio := transform.NewReader(iostr, japanese.ShiftJIS.NewDecoder())
	ret, err := ioutil.ReadAll(rio)
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func Utf82Sjis(str string) (string, error) {
	iostr := strings.NewReader(str)
	rio := transform.NewReader(iostr, japanese.ShiftJIS.NewEncoder())
	ret, err := ioutil.ReadAll(rio)
	if err != nil {
		return "", err
	}
	return string(ret), err
}

// Correct timestamp
func FileInfoHeader(fi os.FileInfo) (*zip.FileHeader, error) {
	size := fi.Size()
	fh := &zip.FileHeader{
		Name:               fi.Name(),
		UncompressedSize64: uint64(size),
	}

	local := time.Now().Local()
	_, offset := local.Zone()
	fh.SetModTime(fi.ModTime().Add(time.Duration(offset) * time.Second))
	fh.SetMode(fi.Mode())
	var uintmax = uint32((1 << 32) - 1)
	if fh.UncompressedSize64 > uint64(uintmax) {
		fh.UncompressedSize = uintmax
	} else {
		fh.UncompressedSize = uint32(fh.UncompressedSize64)
	}
	return fh, nil
}

// Zip file and, convert file name UTF8 to CP932
func Zip(source, output string) error {
	var f *os.File
	if "-" == output {
		f = os.Stdout
	} else {
		var err error
		f, err = os.Create(output)
		if err != nil {
			return err
		}
	}
	defer f.Close()

	archive := zip.NewWriter(f)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			//header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			name, err := Utf82Sjis(filepath.Join(baseDir, strings.TrimPrefix(path, source)))
			if err != nil {
				return err
			}
			header.Name = name
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			//header.Method = zip.Deflate
			header.Method = zip.Store
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// Unzip file, and convert file name CP932 to UTF8
func Unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		utf, err := Sjis2Utf8(path)
		if err != nil {
			fmt.Println(err)
			utf = path
		}
		if file.FileInfo().IsDir() {
			os.MkdirAll(utf, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		fmt.Println(utf)
		targetFile, err := os.OpenFile(utf, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}
