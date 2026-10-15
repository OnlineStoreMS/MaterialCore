package service

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"strings"
	"time"

	"materialcore/internal/dto"
	"materialcore/internal/model"

	"github.com/disintegration/imaging"
	"gorm.io/gorm"
)

// RotateMaterial 下载原图 → 按角度旋转像素并重编码 → 上传覆盖素材 URL。
// 会丢掉原 EXIF Orientation，把正确方向写进像素，避免各端显示不一致。
func (s *MaterialService) RotateMaterial(tenantID, id uint64, degrees int) (*dto.MaterialDTO, error) {
	degrees, ok := normalizeRotateDegrees(degrees)
	if !ok {
		return nil, fmt.Errorf("%w: degrees 须为 0 / ±90 / 180 / 270（0=仅按 EXIF 纠正方向）", ErrBadRequest)
	}
	if s.store == nil {
		return nil, fmt.Errorf("%w: storage not configured", ErrBadRequest)
	}

	row, err := s.repos.Material.Get(tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if row.MediaType != model.MediaTypeImage {
		return nil, fmt.Errorf("%w: 仅支持旋转图片素材", ErrBadRequest)
	}

	srcURL := s.store.ResolvePublicURL(row.URL)
	img, format, err := downloadImage(srcURL)
	if err != nil {
		return nil, fmt.Errorf("%w: 读取原图失败: %v", ErrBadRequest, err)
	}

	img = imaging.Clone(img)
	rotated := applyRotate(img, degrees)
	return s.rewriteMaterialImage(row, rotated, format)
}

func normalizeRotateDegrees(d int) (int, bool) {
	d %= 360
	if d < 0 {
		d += 360
	}
	switch d {
	case 0, 90, 180, 270:
		return d, true
	default:
		return 0, false
	}
}

func applyRotate(img image.Image, degrees int) image.Image {
	switch degrees {
	case 90:
		return imaging.Rotate90(img)
	case 180:
		return imaging.Rotate180(img)
	case 270:
		return imaging.Rotate270(img)
	default:
		return img
	}
}

func downloadImage(url string) (image.Image, string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("http %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 40<<20)) // 40MB
	if err != nil {
		return nil, "", err
	}
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, "", err
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		ct := resp.Header.Get("Content-Type")
		if strings.Contains(ct, "png") {
			format = "png"
		} else {
			format = "jpeg"
		}
	}
	return img, format, nil
}
