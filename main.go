package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	zip "github.com/hillu/go-archive-zip-crypto"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func Utf82Sjis(str string) (string, error) {
	iostr := strings.NewReader(str)
	rio := transform.NewReader(iostr, japanese.ShiftJIS.NewEncoder())
	ret, err := io.ReadAll(rio)
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func Sjis2Utf8(str string) (string, error) {
	iostr := strings.NewReader(str)
	rio := transform.NewReader(iostr, japanese.ShiftJIS.NewDecoder())
	ret, err := io.ReadAll(rio)
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func unzipFile(reader *zip.ReadCloser, file *zip.File, dest string, cp932 bool) error {
	name := file.Name
	if cp932 {
		name_utf, err := Sjis2Utf8(name)
		if err != nil {
			log.Fatalln(err)
		} else {
			name = name_utf
		}
	}

	path := filepath.Join(dest, name)

	fmt.Printf("%s\n", name)

	if file.FileInfo().IsDir() {
		os.MkdirAll(path, file.Mode())
		return nil
	}

	dir := filepath.Dir(path)
	if dir != "." && !Exists(dir) {
		os.MkdirAll(dir, 0755)
	}

	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	fmt.Println(path)
	targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, fileReader); err != nil {
		return err
	}
	return nil
}

func main() {

	cp932 := flag.Bool("cp932", false, "cp932")
	src := flag.String("src", "", "src")
	dst := flag.String("dst", "", "dst")
	m := flag.String("method", "deflate", "method")
	e := flag.String("encryption", "zipcrypto", "encryption")
	password := flag.String("password", "", "password")
	unzip := flag.Bool("unzip", false, "unzip")

	flag.Parse()

	// fmt.Printf("cp932: %t\n", *cp932)

	var encryption zip.EncryptionMethod
	switch *e {
	case "zipcrypto":
		encryption = zip.StandardEncryption
	case "aes128":
		encryption = zip.AES128Encryption
	case "aes192":
		encryption = zip.AES192Encryption
	default:
		encryption = zip.AES256Encryption
	}

	// fmt.Printf("encryption: %v\n", encryption)

	var method uint16
	switch *m {
	case "store":
		method = zip.Store
	default:
		method = zip.Deflate
	}

	// fmt.Printf("method: %v\n", method)

	if *unzip {
		if Exists(*src) {
			var err error
			reader, err := zip.OpenReader(*src)
			if err != nil {
				log.Fatalln(err)
			}
			defer reader.Close()

			if err := os.MkdirAll(*dst, 0755); err != nil {
				log.Fatalln(err)
			}

			for _, file := range reader.File {
				if *password != "" {
					file.SetPassword(*password)
				}

				err := unzipFile(reader, file, *dst, *cp932)
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	} else {
		var zipFile *os.File
		if *dst == "-" {
			zipFile = os.Stdout
		} else {
			var err error
			zipFile, err = os.Create(*dst)
			if err != nil {
				log.Fatalln(err)
			}
		}
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		info, err := os.Stat(*src)
		if err != nil {
			log.Fatalln(err)
		}

		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Base(*src)
		}

		filepath.Walk(*src, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			var name string
			if baseDir != "" {
				//src is folder
				name = filepath.Join(baseDir, strings.TrimPrefix(path, *src))
			} else {
				//src is file
				name = filepath.Base(path)
			}

			if !info.IsDir() {
				header.Method = method
			} else {
				header.Name += "/"
			}

			fmt.Printf("%s\n", name)

			if *cp932 {
				name_cp932, err := Utf82Sjis(name)
				if err != nil {
					log.Fatalln(err)
				} else {
					name = name_cp932
				}
			}

			header.Name = name

			if info.IsDir() {
				_, err := zipWriter.CreateHeader(header)
				if err != nil {
					return err
				}
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			if *password != "" {
				w, err := zipWriter.Encrypt(header.Name, *password, encryption)
				if err != nil {
					log.Fatal(err)
				}
				_, err = io.Copy(w, file)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				writer, err := zipWriter.CreateHeader(header)
				if err != nil {
					return err
				}
				_, err = io.Copy(writer, file)
				if err != nil {
					log.Fatal(err)
				}
			}
			file.Close()
			return err
		})

		zipWriter.Flush()
	}
}
