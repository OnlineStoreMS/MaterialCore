package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"materialcore/internal/dto"
	"materialcore/internal/model"
	"materialcore/internal/repo"

	"gorm.io/gorm"
)

type MaterialService struct {
	repos *repo.Repos
}

func NewMaterialService(repos *repo.Repos) *MaterialService {
	return &MaterialService{repos: repos}
}

func (s *MaterialService) ListCategories(tenantID uint64) ([]dto.CategoryNode, error) {
	list, err := s.repos.Category.ListAll(tenantID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		if err := s.repos.EnsureDefaultCategories(tenantID); err != nil {
			return nil, err
		}
		list, err = s.repos.Category.ListAll(tenantID)
		if err != nil {
			return nil, err
		}
	}
	return buildCategoryTree(list), nil
}

func buildCategoryTree(list []model.MaterialCategory) []dto.CategoryNode {
	byParent := map[uint64][]model.MaterialCategory{}
	for _, c := range list {
		byParent[c.ParentID] = append(byParent[c.ParentID], c)
	}
	var walk func(parentID uint64) []dto.CategoryNode
	walk = func(parentID uint64) []dto.CategoryNode {
		children := byParent[parentID]
		out := make([]dto.CategoryNode, 0, len(children))
		for _, c := range children {
			node := dto.CategoryNode{
				ID:       c.ID,
				ParentID: c.ParentID,
				Name:     c.Name,
				Code:     c.Code,
				Sort:     c.Sort,
				Enabled:  c.Enabled,
				Children: walk(c.ID),
			}
			out = append(out, node)
		}
		return out
	}
	return walk(0)
}

func (s *MaterialService) CreateCategory(tenantID uint64, req dto.CategoryCreateReq) (*dto.CategoryNode, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrBadRequest
	}
	exists, err := s.repos.Category.ExistsName(tenantID, req.ParentID, name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: 同级分类名称已存在", ErrConflict)
	}
	if req.ParentID > 0 {
		if _, err := s.repos.Category.Get(tenantID, req.ParentID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("%w: 父分类不存在", ErrNotFound)
			}
			return nil, err
		}
	}
	enabled := 1
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	row := &model.MaterialCategory{
		TenantID: repo.NormalizeTenantID(tenantID),
		ParentID: req.ParentID,
		Name:     name,
		Sort:     req.Sort,
		Enabled:  enabled,
	}
	if err := s.repos.Category.Create(row); err != nil {
		return nil, err
	}
	return &dto.CategoryNode{
		ID: row.ID, ParentID: row.ParentID, Name: row.Name,
		Code: row.Code, Sort: row.Sort, Enabled: row.Enabled,
	}, nil
}

func (s *MaterialService) UpdateCategory(tenantID, id uint64, req dto.CategoryUpdateReq) (*dto.CategoryNode, error) {
	row, err := s.repos.Category.Get(tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrBadRequest
		}
		parentID := row.ParentID
		if req.ParentID != nil {
			parentID = *req.ParentID
		}
		exists, err := s.repos.Category.ExistsName(tenantID, parentID, name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("%w: 同级分类名称已存在", ErrConflict)
		}
		row.Name = name
	}
	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, fmt.Errorf("%w: 不能将分类设为自己的子节点", ErrBadRequest)
		}
		if *req.ParentID > 0 {
			if _, err := s.repos.Category.Get(tenantID, *req.ParentID); err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, fmt.Errorf("%w: 父分类不存在", ErrNotFound)
				}
				return nil, err
			}
			desc, err := s.repos.Category.ListDescendantIDs(tenantID, id)
			if err != nil {
				return nil, err
			}
			for _, did := range desc {
				if did == *req.ParentID {
					return nil, fmt.Errorf("%w: 不能移动到自己的子分类下", ErrBadRequest)
				}
			}
		}
		row.ParentID = *req.ParentID
	}
	if req.Sort != nil {
		row.Sort = *req.Sort
	}
	if req.Enabled != nil {
		row.Enabled = *req.Enabled
	}
	if err := s.repos.Category.Update(row); err != nil {
		return nil, err
	}
	return &dto.CategoryNode{
		ID: row.ID, ParentID: row.ParentID, Name: row.Name,
		Code: row.Code, Sort: row.Sort, Enabled: row.Enabled,
	}, nil
}

func (s *MaterialService) DeleteCategory(tenantID, id uint64) error {
	row, err := s.repos.Category.Get(tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		return err
	}
	if row.Code == model.CategoryCodeUncategorized {
		return fmt.Errorf("%w: 系统分类「未分类」不可删除", ErrBadRequest)
	}
	nChild, err := s.repos.Category.CountChildren(tenantID, id)
	if err != nil {
		return err
	}
	if nChild > 0 {
		return fmt.Errorf("%w: 请先删除或迁移子分类", ErrBadRequest)
	}
	nMat, err := s.repos.Material.CountByCategory(tenantID, id)
	if err != nil {
		return err
	}
	if nMat > 0 {
		return fmt.Errorf("%w: 分类下仍有素材，请先迁移", ErrBadRequest)
	}
	return s.repos.Category.Delete(tenantID, id)
}

func (s *MaterialService) ReorderCategories(tenantID uint64, items []dto.CategoryReorderItem) error {
	for _, it := range items {
		row, err := s.repos.Category.Get(tenantID, it.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrNotFound
			}
			return err
		}
		row.ParentID = it.ParentID
		row.Sort = it.Sort
		if err := s.repos.Category.Update(row); err != nil {
			return err
		}
	}
	return nil
}

func encodeTags(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	cleaned := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		cleaned = append(cleaned, t)
	}
	b, _ := json.Marshal(cleaned)
	return string(b)
}

func decodeTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return []string{}
	}
	return tags
}

func toMaterialDTO(m *model.Material) dto.MaterialDTO {
	return dto.MaterialDTO{
		ID: m.ID, CategoryID: m.CategoryID, Title: m.Title, MediaType: m.MediaType,
		URL: m.URL, CoverURL: m.CoverURL, FileName: m.FileName, Mime: m.Mime,
		SizeBytes: m.SizeBytes, DurationMs: m.DurationMs, Width: m.Width, Height: m.Height,
		Tags: decodeTags(m.TagsJSON), Remark: m.Remark, ProductID: m.ProductID, ProductSn: m.ProductSn,
		Sort: m.Sort, Status: m.Status, CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *MaterialService) resolveDefaultCategoryID(tenantID uint64, categoryID uint64) (uint64, error) {
	if categoryID > 0 {
		if _, err := s.repos.Category.Get(tenantID, categoryID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, fmt.Errorf("%w: 分类不存在", ErrNotFound)
			}
			return 0, err
		}
		return categoryID, nil
	}
	uncat, err := s.repos.Category.GetByCode(tenantID, model.CategoryCodeUncategorized)
	if err != nil {
		return 0, fmt.Errorf("%w: 未分类不存在，请重启服务初始化", ErrNotFound)
	}
	return uncat.ID, nil
}

func (s *MaterialService) CreateMaterial(tenantID, userID uint64, req dto.MaterialCreateReq) (*dto.MaterialDTO, error) {
	mt := strings.ToLower(strings.TrimSpace(req.MediaType))
	if mt != model.MediaTypeImage && mt != model.MediaTypeVideo {
		return nil, fmt.Errorf("%w: mediaType 须为 image 或 video", ErrBadRequest)
	}
	url := strings.TrimSpace(req.URL)
	if url == "" {
		return nil, ErrBadRequest
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = req.FileName
	}
	if title == "" {
		title = "未命名素材"
	}
	catID, err := s.resolveDefaultCategoryID(tenantID, req.CategoryID)
	if err != nil {
		return nil, err
	}
	row := &model.Material{
		TenantID:   repo.NormalizeTenantID(tenantID),
		CategoryID: catID,
		Title:      title,
		MediaType:  mt,
		URL:        url,
		CoverURL:   strings.TrimSpace(req.CoverURL),
		FileName:   strings.TrimSpace(req.FileName),
		Mime:       strings.TrimSpace(req.Mime),
		SizeBytes:  req.SizeBytes,
		DurationMs: req.DurationMs,
		Width:      req.Width,
		Height:     req.Height,
		TagsJSON:   encodeTags(req.Tags),
		Remark:     strings.TrimSpace(req.Remark),
		ProductID:  req.ProductID,
		ProductSn:  strings.TrimSpace(req.ProductSn),
		Sort:       req.Sort,
		Status:     model.MaterialStatusActive,
		CreatedBy:  userID,
	}
	if err := s.repos.Material.Create(row); err != nil {
		return nil, err
	}
	d := toMaterialDTO(row)
	return &d, nil
}

func (s *MaterialService) UpdateMaterial(tenantID, id uint64, req dto.MaterialUpdateReq) (*dto.MaterialDTO, error) {
	row, err := s.repos.Material.Get(tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if req.CategoryID != nil {
		catID, err := s.resolveDefaultCategoryID(tenantID, *req.CategoryID)
		if err != nil {
			return nil, err
		}
		row.CategoryID = catID
	}
	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return nil, ErrBadRequest
		}
		row.Title = t
	}
	if req.CoverURL != nil {
		row.CoverURL = strings.TrimSpace(*req.CoverURL)
	}
	if req.Tags != nil {
		row.TagsJSON = encodeTags(req.Tags)
	}
	if req.Remark != nil {
		row.Remark = strings.TrimSpace(*req.Remark)
	}
	if req.ClearProductID {
		row.ProductID = nil
	} else if req.ProductID != nil {
		row.ProductID = req.ProductID
	}
	if req.ProductSn != nil {
		row.ProductSn = strings.TrimSpace(*req.ProductSn)
	}
	if req.Sort != nil {
		row.Sort = *req.Sort
	}
	if req.Status != nil {
		row.Status = *req.Status
	}
	if err := s.repos.Material.Update(row); err != nil {
		return nil, err
	}
	d := toMaterialDTO(row)
	return &d, nil
}

func (s *MaterialService) GetMaterial(tenantID, id uint64) (*dto.MaterialDTO, error) {
	row, err := s.repos.Material.Get(tenantID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	d := toMaterialDTO(row)
	return &d, nil
}

func (s *MaterialService) DeleteMaterials(tenantID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return ErrBadRequest
	}
	return s.repos.Material.SoftDelete(tenantID, ids)
}

func (s *MaterialService) ListMaterials(tenantID uint64, q dto.MaterialListQuery) ([]dto.MaterialDTO, int64, error) {
	filter := repo.MaterialListFilter{
		MediaType:  strings.TrimSpace(q.MediaType),
		Keyword:    q.Keyword,
		ProductSn:  q.ProductSn,
		HasProduct: q.HasProduct,
		Tag:        q.Tag,
		Page:       q.Page,
		PageSize:   q.PageSize,
	}
	if q.CategoryID > 0 {
		if q.IncludeChildren {
			ids, err := s.repos.Category.ListDescendantIDs(tenantID, q.CategoryID)
			if err != nil {
				return nil, 0, err
			}
			filter.CategoryIDs = ids
		} else {
			filter.CategoryIDs = []uint64{q.CategoryID}
		}
	}
	list, total, err := s.repos.Material.List(tenantID, filter)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.MaterialDTO, 0, len(list))
	for i := range list {
		out = append(out, toMaterialDTO(&list[i]))
	}
	return out, total, nil
}

func (s *MaterialService) ListMaterialsByIDs(tenantID uint64, ids []uint64) ([]model.Material, error) {
	return s.repos.Material.GetByIDs(tenantID, ids)
}

func (s *MaterialService) DashboardStats(tenantID uint64) (*dto.DashboardStats, error) {
	total, images, videos, weekNew, err := s.repos.Material.Stats(tenantID)
	if err != nil {
		return nil, err
	}
	catCount, err := s.repos.Category.CountByTenant(tenantID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repos.Material.CountGroupByCategory(tenantID)
	if err != nil {
		return nil, err
	}
	cats, err := s.repos.Category.ListAll(tenantID)
	if err != nil {
		return nil, err
	}
	nameByID := map[uint64]string{}
	for _, c := range cats {
		nameByID[c.ID] = c.Name
	}
	byCat := make([]dto.CategoryCount, 0, len(rows))
	for _, r := range rows {
		byCat = append(byCat, dto.CategoryCount{
			CategoryID:   r.CategoryID,
			CategoryName: nameByID[r.CategoryID],
			Count:        r.Count,
		})
	}
	return &dto.DashboardStats{
		MaterialCount: total,
		ImageCount:    images,
		VideoCount:    videos,
		WeekNewCount:  weekNew,
		CategoryCount: catCount,
		ByCategory:    byCat,
	}, nil
}
