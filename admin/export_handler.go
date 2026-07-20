package admin

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"materialcore/internal/dto"
	"materialcore/internal/pkg/authcontext"
	"materialcore/internal/pkg/response"
	"materialcore/internal/service"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	svc *service.MaterialService
}

func NewExportHandler(svc *service.MaterialService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

func (h *ExportHandler) ExportZip(c *gin.Context) {
	var req dto.BatchIDsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.IDs) == 0 {
		response.Fail(c, http.StatusBadRequest, "ids required")
		return
	}
	if len(req.IDs) > 100 {
		response.Fail(c, http.StatusBadRequest, "最多打包 100 个素材")
		return
	}

	list, err := h.svc.ListMaterialsByIDs(authcontext.TenantID(c), req.IDs)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if len(list) == 0 {
		response.Fail(c, http.StatusNotFound, "未找到素材")
		return
	}

	filename := fmt.Sprintf("materials_%s.zip", time.Now().Format("20060102150405"))
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	client := &http.Client{Timeout: 60 * time.Second}
	usedNames := map[string]int{}

	for _, m := range list {
		name := m.FileName
		if name == "" {
			if u, err := url.Parse(m.URL); err == nil {
				name = path.Base(u.Path)
			}
		}
		if name == "" || name == "." || name == "/" {
			ext := ".bin"
			if m.MediaType == "video" {
				ext = ".mp4"
			} else {
				ext = ".jpg"
			}
			name = fmt.Sprintf("%d%s", m.ID, ext)
		}
		name = filepath.Base(name)
		if n, ok := usedNames[name]; ok {
			usedNames[name] = n + 1
			ext := filepath.Ext(name)
			base := strings.TrimSuffix(name, ext)
			name = fmt.Sprintf("%s_%d%s", base, n+1, ext)
		} else {
			usedNames[name] = 1
		}

		resp, err := client.Get(m.URL)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}
		w, err := zw.Create(name)
		if err != nil {
			resp.Body.Close()
			continue
		}
		_, _ = io.Copy(w, resp.Body)
		resp.Body.Close()
	}
}
