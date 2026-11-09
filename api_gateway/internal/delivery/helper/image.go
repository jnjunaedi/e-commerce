package helper

import (
	"e-commerce/pkg/utils/apperror"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

func ProcessImage(files []*multipart.FileHeader) ([]string, error) {
	paths := make([]string, 0, len(files))

	for _, header := range files {
		err := func() error {
			log.Println("1")
			file, err := header.Open()
			if err != nil {
				return apperror.New(apperror.ErrInternalServer, "gagal menyimpan gambar")
			}
			defer file.Close()
			log.Println("2")

			buf := make([]byte, 512)
			n, err := file.Read(buf)
			var detectType string

			log.Println("3")
			if n > 0 {
				detectType = http.DetectContentType(buf[:n])
			}

			log.Println("4")
			if err != nil {
				return apperror.New(apperror.ErrInternalServer, "gagal menyimpan gambar")
			}

			log.Println("5")
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				return apperror.New(apperror.ErrInternalServer, "gagal menyimpan gambar")
			}

			log.Println("6")
			ext := GetExtFilename(detectType)
			if ext == "" {
				return apperror.New(apperror.ErrBadRequest, "Format gambar tidak didukung")
			}

			log.Println("7")
			genUUID, err := uuid.NewV7()
			if err != nil {
				return apperror.New(apperror.ErrInternalServer, "gagal menyimpan gambar")
			}

			log.Println("8")
			safeFilename := genUUID.String() + ext
			outputPath := filepath.Join("api_gateway", "uploads", "product_images", safeFilename)

			log.Println("9")
			path, err := CompressAndResizeImage(file, outputPath, 800)

			log.Println("10")
			paths = append(paths, path)
			return nil
		}()

		if err != nil {
			return []string{}, err
		}
	}

	return paths, nil
}

func CompressAndResizeImage(src io.Reader, outputPath string, maxWidth int) (string, error) {
	imgSrc, err := imaging.Decode(src)
	if err != nil {
		return "", apperror.New(apperror.ErrInternalServer, "gagal menyimpan gambar")
	}

	var processImage = imgSrc
	if imgSrc.Bounds().Dx() > maxWidth {
		processImage = imaging.Resize(imgSrc, maxWidth, 0, imaging.Lanczos)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", apperror.New(apperror.ErrInternalServer, "gagal membuat folder penyimpanan fisik")
	}

	err = imaging.Save(processImage, outputPath)
	if err != nil {
		return "", apperror.New(apperror.ErrInternalServer, "gagal menyimpan gambar")
	}

	return outputPath, nil
}

func GetExtFilename(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpeg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}
