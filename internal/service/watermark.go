package service

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"materialcore/internal/dto"
	"materialcore/internal/model"

	"github.com/disintegration/imaging"
	"gorm.io/gorm"
)

const (
	wmPosTopLeft      = "top-left"
	wmPosTopRight     = "top-right"
	wmPosBottomLeft   = "bottom-left"
	wmPosBottomRight  = "bottom-right"
	wmPosCenter       = "center"
	wmPosTopCenter    = "top-center"
	wmPosBottomCenter = "bottom-center"
	wmPosCustom       = "custom"
)

// WatermarkMaterials 给一张或多张图片叠加 Logo/文字水印图（文字可前端渲染成透明 PNG）。
func (s *MaterialService) WatermarkMaterials(tenantID uint64, req dto.MaterialWatermarkReq) (*dto.MaterialWatermarkResult, error) {
	if s.store == nil {
		return nil, fmt.Errorf("%w: storage not configured", ErrBadRequest)
	}
	if len(req.IDs) == 0 {
		return nil, fmt.Errorf("%w: ids required", ErrBadRequest)
	}
	if len(req.IDs) > 50 {
		return nil, fmt.Errorf("%w: 单次最多 50 张", ErrBadRequest)
	}
	logoURL := strings.TrimSpace(req.LogoURL)
	if logoURL == "" {
		return nil, fmt.Errorf("%w: logoUrl required", ErrBadRequest)
	}

	pos := strings.ToLower(strings.TrimSpace(req.Position))
	if pos == "" {
		pos = wmPosBottomRight
	}
	if !validWatermarkPosition(pos) {
		return nil, fmt.Errorf("%w: 无效 position", ErrBadRequest)
	}

	scale := req.Scale
	if scale <= 0 {
		scale = 0.18
	}
	if scale > 0.8 {
		scale = 0.8
	}
	opacity := req.Opacity
	if opacity <= 0 {
		opacity = 0.7
	}
	if opacity > 1 {
		opacity = 1
	}
	margin := req.Margin
	if margin < 0 {
		margin = 0
	}
	if margin == 0 && pos != wmPosCustom {
		margin = 0.03
	}
	if margin > 0.3 {
		margin = 0.3
	}

	logoURL = s.store.ResolvePublicURL(logoURL)
	logoImg, _, err := downloadImage(logoURL)
	if err != nil {
		return nil, fmt.Errorf("%w: 读取水印图失败: %v", ErrBadRequest, err)
	}
	logoImg = imaging.Clone(logoImg)

	out := &dto.MaterialWatermarkResult{
		Success: make([]dto.MaterialDTO, 0, len(req.IDs)),
		Failed:  make([]dto.MaterialWatermarkFail, 0),
	}

	for _, id := range req.IDs {
		m, err := s.watermarkOne(tenantID, id, logoImg, pos, scale, opacity, margin, req.OffsetX, req.OffsetY)
		if err != nil {
			out.Failed = append(out.Failed, dto.MaterialWatermarkFail{ID: id, Error: err.Error()})
			continue
		}
		out.Success = append(out.Success, *m)
	}
	out.OK = len(out.Success)
	out.Fail = len(out.Failed)
	return out, nil
}

func (s *MaterialService) watermarkOne(
	tenantID, id uint64,
	logoImg image.Image,
	pos string,
	scale, opacity, margin, offsetX, offsetY float64,
) (*dto.MaterialDTO, error) {
	row, err := s.repos.Material.Get(tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if row.MediaType != model.MediaTypeImage {
		return nil, fmt.Errorf("%w: 仅支持图片", ErrBadRequest)
	}

	srcURL := s.store.ResolvePublicURL(row.URL)
	base, format, err := downloadImage(srcURL)
	if err != nil {
		return nil, fmt.Errorf("读取原图失败: %v", err)
	}
	base = imaging.Clone(base)
	dst := imaging.Clone(base)

	bw := dst.Bounds().Dx()
	bh := dst.Bounds().Dy()
	if bw < 8 || bh < 8 {
		return nil, fmt.Errorf("%w: 图片过小", ErrBadRequest)
	}

	targetW := int(math.Round(float64(bw) * scale))
	if targetW < 8 {
		targetW = 8
	}
	lw0, lh0 := logoImg.Bounds().Dx(), logoImg.Bounds().Dy()
	if lw0 < 1 || lh0 < 1 {
		return nil, fmt.Errorf("%w: 水印图无效", ErrBadRequest)
	}
	targetH := int(math.Round(float64(targetW) * float64(lh0) / float64(lw0)))
	if targetH < 8 {
		targetH = 8
	}
	resized := imaging.Resize(logoImg, targetW, targetH, imaging.Lanczos)
	if opacity < 0.999 {
		resized = applyOpacity(resized, opacity)
	}

	x, y := resolveWatermarkXY(bw, bh, targetW, targetH, pos, margin, offsetX, offsetY)
	overlayLogo(dst, resized, x, y)

	return s.rewriteMaterialImage(row, dst, format)
}

func validWatermarkPosition(pos string) bool {
	switch pos {
	case wmPosTopLeft, wmPosTopRight, wmPosBottomLeft, wmPosBottomRight,
		wmPosCenter, wmPosTopCenter, wmPosBottomCenter, wmPosCustom:
		return true
	default:
		return false
	}
}

// resolveWatermarkXY
// - 固定位：按 margin（相对短边比例）贴角/居中
// - custom：offsetX/Y 为水印左上角相对画布的百分比 (0-100)
func resolveWatermarkXY(bw, bh, lw, lh int, pos string, margin, offsetX, offsetY float64) (int, int) {
	mx := int(math.Round(float64(bw) * margin))
	my := int(math.Round(float64(bh) * margin))
	if mx < 0 {
		mx = 0
	}
	if my < 0 {
		my = 0
	}

	var x, y int
	switch pos {
	case wmPosTopLeft:
		x, y = mx, my
	case wmPosTopRight:
		x, y = bw-lw-mx, my
	case wmPosBottomLeft:
		x, y = mx, bh-lh-my
	case wmPosBottomRight:
		x, y = bw-lw-mx, bh-lh-my
	case wmPosTopCenter:
		x, y = (bw-lw)/2, my
	case wmPosBottomCenter:
		x, y = (bw-lw)/2, bh-lh-my
	case wmPosCenter:
		x, y = (bw-lw)/2, (bh-lh)/2
	case wmPosCustom:
		x = int(math.Round(float64(bw) * clamp01(offsetX/100)))
		y = int(math.Round(float64(bh) * clamp01(offsetY/100)))
	default:
		x, y = bw-lw-mx, bh-lh-my
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > bw-lw {
		x = max(0, bw-lw)
	}
	if y > bh-lh {
		y = max(0, bh-lh)
	}
	return x, y
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func applyOpacity(src image.Image, opacity float64) *image.NRGBA {
	b := src.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Bounds(), src, b.Min, draw.Src)
	aScale := opacity
	for i := 0; i < len(out.Pix); i += 4 {
		out.Pix[i+3] = uint8(math.Round(float64(out.Pix[i+3]) * aScale))
	}
	return out
}

func overlayLogo(dst *image.NRGBA, logo image.Image, x, y int) {
	lb := logo.Bounds()
	db := dst.Bounds()
	for py := 0; py < lb.Dy(); py++ {
		dy := y + py
		if dy < db.Min.Y || dy >= db.Max.Y {
			continue
		}
		for px := 0; px < lb.Dx(); px++ {
			dx := x + px
			if dx < db.Min.X || dx >= db.Max.X {
				continue
			}
			sr, sg, sb, sa := logo.At(lb.Min.X+px, lb.Min.Y+py).RGBA()
			if sa == 0 {
				continue
			}
			dr, dg, dbv, da := dst.At(dx, dy).RGBA()
			// src-over
			a := float64(sa) / 0xffff
			inv := 1 - a
			outR := uint8((float64(sr)/0xffff*a + float64(dr)/0xffff*inv) * 255)
			outG := uint8((float64(sg)/0xffff*a + float64(dg)/0xffff*inv) * 255)
			outB := uint8((float64(sb)/0xffff*a + float64(dbv)/0xffff*inv) * 255)
			outA := uint8(math.Min(255, (a+float64(da)/0xffff*inv)*255))
			dst.SetNRGBA(dx, dy, color.NRGBA{R: outR, G: outG, B: outB, A: outA})
		}
	}
}
