package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"path"
	"strings"
	"time"

	"materialcore/internal/dto"
	"materialcore/internal/model"
	"materialcore/internal/repo"
)

func encodeImage(img image.Image, preferFormat string) (data []byte, mime, ext string, err error) {
	outFmt := strings.ToLower(preferFormat)
	if outFmt != "png" && outFmt != "jpeg" && outFmt != "jpg" {
		outFmt = "jpeg"
	}
	if outFmt == "jpg" {
		outFmt = "jpeg"
	}
	var buf bytes.Buffer
	if outFmt == "png" {
		mime, ext = "image/png", ".png"
		err = png.Encode(&buf, img)
	} else {
		mime, ext = "image/jpeg", ".jpg"
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 92})
	}
	if err != nil {
		return nil, "", "", err
	}
	return buf.Bytes(), mime, ext, nil
}

func (s *MaterialService) rewriteMaterialImage(row *model.Material, img image.Image, preferFormat string) (*dto.MaterialDTO, error) {
	if s.store == nil {
		return nil, fmt.Errorf("%w: storage not configured", ErrBadRequest)
	}
	data, mime, ext, err := encodeImage(img, preferFormat)
	if err != nil {
		return nil, err
	}

	baseName := row.FileName
	if baseName == "" {
		baseName = path.Base(strings.Split(row.URL, "?")[0])
	}
	if baseName == "" || baseName == "." || baseName == "/" {
		baseName = fmt.Sprintf("material-%d%s", row.ID, ext)
	} else {
		baseName = strings.TrimSuffix(baseName, path.Ext(baseName)) + ext
	}

	tenant := repo.NormalizeTenantID(row.TenantID)
	month := time.Now().Format("200601")
	subdir := fmt.Sprintf("materials/%d/%s", tenant, month)

	newURL, err := s.store.UploadBytes(data, baseName, mime, subdir)
	if err != nil {
		return nil, err
	}

	b := img.Bounds()
	row.URL = newURL
	row.Mime = mime
	row.FileName = baseName
	row.SizeBytes = int64(len(data))
	row.Width = b.Dx()
	row.Height = b.Dy()
	if err := s.repos.Material.Update(row); err != nil {
		return nil, err
	}
	d := s.toMaterialDTO(row)
	return &d, nil
}
