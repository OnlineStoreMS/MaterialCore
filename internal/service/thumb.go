package service

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"path"
	"strings"
	"time"

	"materialcore/internal/model"
	"materialcore/internal/repo"

	"github.com/disintegration/imaging"
	"gorm.io/gorm"
)

const listThumbMaxEdge = 400

// EnsureImageThumb 为图片素材生成列表缩略图（约 400px 边长 JPEG），写入 ThumbURL。
func (s *MaterialService) EnsureImageThumb(tenantID, id uint64) (string, error) {
	if s.store == nil {
		return "", fmt.Errorf("%w: storage not configured", ErrBadRequest)
	}
	row, err := s.repos.Material.Get(tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", ErrNotFound
		}
		return "", err
	}
	if row.MediaType != model.MediaTypeImage {
		return "", fmt.Errorf("%w: 仅图片可生成缩略图", ErrBadRequest)
	}
	if strings.TrimSpace(row.ThumbURL) != "" {
		return s.store.ResolvePublicURL(row.ThumbURL), nil
	}

	srcURL := s.store.ResolvePublicURL(row.URL)
	img, _, err := downloadImage(srcURL)
	if err != nil {
		return "", fmt.Errorf("%w: 读取原图失败: %v", ErrBadRequest, err)
	}
	img = imaging.Clone(img)
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 1 || h < 1 {
		return "", fmt.Errorf("%w: 图片无效", ErrBadRequest)
	}
	tw, th := w, h
	if w >= h {
		if w > listThumbMaxEdge {
			tw = listThumbMaxEdge
			th = h * listThumbMaxEdge / w
		}
	} else if h > listThumbMaxEdge {
		th = listThumbMaxEdge
		tw = w * listThumbMaxEdge / h
	}
	if tw != w || th != h {
		img = imaging.Resize(img, tw, th, imaging.Lanczos)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 78}); err != nil {
		return "", err
	}

	baseName := row.FileName
	if baseName == "" {
		baseName = path.Base(strings.Split(row.URL, "?")[0])
	}
	baseName = strings.TrimSuffix(baseName, path.Ext(baseName)) + "_thumb.jpg"
	tenant := repo.NormalizeTenantID(tenantID)
	month := time.Now().Format("200601")
	subdir := fmt.Sprintf("thumbs/%d/%s", tenant, month)

	thumbURL, err := s.store.UploadBytes(buf.Bytes(), baseName, "image/jpeg", subdir)
	if err != nil {
		return "", err
	}
	row.ThumbURL = thumbURL
	if err := s.repos.Material.Update(row); err != nil {
		return "", err
	}
	return s.store.ResolvePublicURL(thumbURL), nil
}

// EnsureThumbs 批量补齐缩略图（跳过已有 / 非图片）
func (s *MaterialService) EnsureThumbs(tenantID uint64, ids []uint64) (ok int, fail int) {
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, err := s.EnsureImageThumb(tenantID, id); err != nil {
			fail++
			continue
		}
		ok++
	}
	return
}

func (s *MaterialService) maybeGenerateThumbAsync(tenantID, id uint64, mediaType, coverURL string) {
	if mediaType != model.MediaTypeImage {
		return
	}
	if strings.TrimSpace(coverURL) != "" {
		// 已有封面则列表可用，仍尝试生成独立 thumb（后台）
	}
	go func() {
		_, _ = s.EnsureImageThumb(tenantID, id)
	}()
}
