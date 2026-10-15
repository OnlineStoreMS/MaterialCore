package dto

type CategoryNode struct {
	ID       uint64         `json:"id"`
	ParentID uint64         `json:"parentId"`
	Name     string         `json:"name"`
	Code     string         `json:"code"`
	Sort     int            `json:"sort"`
	Enabled  int            `json:"enabled"`
	Children []CategoryNode `json:"children,omitempty"`
}

type CategoryCreateReq struct {
	ParentID uint64 `json:"parentId"`
	Name     string `json:"name" binding:"required"`
	Sort     int    `json:"sort"`
	Enabled  *int   `json:"enabled"`
}

type CategoryUpdateReq struct {
	ParentID *uint64 `json:"parentId"`
	Name     *string `json:"name"`
	Sort     *int    `json:"sort"`
	Enabled  *int    `json:"enabled"`
}

type CategoryReorderReq struct {
	Items []CategoryReorderItem `json:"items" binding:"required"`
}

type CategoryReorderItem struct {
	ID       uint64 `json:"id" binding:"required"`
	ParentID uint64 `json:"parentId"`
	Sort     int    `json:"sort"`
}

type MaterialDTO struct {
	ID         uint64   `json:"id"`
	CategoryID uint64   `json:"categoryId"`
	Title      string   `json:"title"`
	MediaType  string   `json:"mediaType"`
	URL        string   `json:"url"`
	CoverURL   string   `json:"coverUrl"`
	FileName   string   `json:"fileName"`
	Mime       string   `json:"mime"`
	SizeBytes  int64    `json:"sizeBytes"`
	DurationMs int64    `json:"durationMs"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Tags       []string `json:"tags"`
	Remark     string   `json:"remark"`
	ProductID  *uint64  `json:"productId"`
	ProductSn  string   `json:"productSn"`
	Sort       int      `json:"sort"`
	Status     int      `json:"status"`
	CreatedBy  uint64   `json:"createdBy"`
	CreatedAt  string   `json:"createdAt"`
	UpdatedAt  string   `json:"updatedAt"`
}

type MaterialCreateReq struct {
	CategoryID uint64   `json:"categoryId"`
	Title      string   `json:"title" binding:"required"`
	MediaType  string   `json:"mediaType" binding:"required"`
	URL        string   `json:"url" binding:"required"`
	CoverURL   string   `json:"coverUrl"`
	FileName   string   `json:"fileName"`
	Mime       string   `json:"mime"`
	SizeBytes  int64    `json:"sizeBytes"`
	DurationMs int64    `json:"durationMs"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Tags       []string `json:"tags"`
	Remark     string   `json:"remark"`
	ProductID  *uint64  `json:"productId"`
	ProductSn  string   `json:"productSn"`
	Sort       int      `json:"sort"`
}

type MaterialUpdateReq struct {
	CategoryID *uint64  `json:"categoryId"`
	Title      *string  `json:"title"`
	URL        *string  `json:"url"`
	CoverURL   *string  `json:"coverUrl"`
	FileName   *string  `json:"fileName"`
	Mime       *string  `json:"mime"`
	SizeBytes  *int64   `json:"sizeBytes"`
	Width      *int     `json:"width"`
	Height     *int     `json:"height"`
	Tags       []string `json:"tags"`
	Remark     *string  `json:"remark"`
	ProductID  *uint64  `json:"productId"`
	ClearProductID bool `json:"clearProductId"`
	ProductSn  *string  `json:"productSn"`
	Sort       *int     `json:"sort"`
	Status     *int     `json:"status"`
}

// MaterialRotateReq 旋转并重写图片像素后重新入库（清除错误 EXIF 方向）
type MaterialRotateReq struct {
	// Degrees: 0=仅按 EXIF 纠正；90/-90/180/270 再旋转（顺时针为正）
	Degrees int `json:"degrees"`
}

// MaterialWatermarkReq 批量/单张叠加水印（文字请前端渲染为透明 PNG 作为 logoUrl）
type MaterialWatermarkReq struct {
	IDs []uint64 `json:"ids" binding:"required"`
	// LogoURL 水印图（Logo 或文字渲染图）
	LogoURL string `json:"logoUrl" binding:"required"`
	// Position: top-left | top-right | bottom-left | bottom-right | center | top-center | bottom-center | custom
	Position string `json:"position"`
	// OffsetX/Y：custom 时为水印左上角相对画布宽高的百分比 0-100
	OffsetX float64 `json:"offsetX"`
	OffsetY float64 `json:"offsetY"`
	// Margin 固定位边距，相对宽/高比例，默认 0.03
	Margin float64 `json:"margin"`
	// Scale 水印宽度占原图宽度比例，默认 0.18
	Scale float64 `json:"scale"`
	// Opacity 0-1，默认 0.7
	Opacity float64 `json:"opacity"`
}

type MaterialWatermarkFail struct {
	ID    uint64 `json:"id"`
	Error string `json:"error"`
}

type MaterialWatermarkResult struct {
	OK      int                      `json:"ok"`
	Fail    int                      `json:"fail"`
	Success []MaterialDTO            `json:"success"`
	Failed  []MaterialWatermarkFail  `json:"failed"`
}

type MaterialListQuery struct {
	CategoryID  uint64 `form:"categoryId"`
	IncludeChildren bool `form:"includeChildren"`
	MediaType   string `form:"mediaType"`
	Keyword     string `form:"keyword"`
	ProductSn   string `form:"productSn"`
	HasProduct  string `form:"hasProduct"` // yes | no | ""
	Tag         string `form:"tag"`
	Page        int    `form:"page"`
	PageSize    int    `form:"pageSize"`
}

type BatchIDsReq struct {
	IDs []uint64 `json:"ids" binding:"required"`
}

type DashboardStats struct {
	MaterialCount    int64            `json:"materialCount"`
	ImageCount       int64            `json:"imageCount"`
	VideoCount       int64            `json:"videoCount"`
	WeekNewCount     int64            `json:"weekNewCount"`
	CategoryCount    int64            `json:"categoryCount"`
	ByCategory       []CategoryCount  `json:"byCategory"`
}

type CategoryCount struct {
	CategoryID   uint64 `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Count        int64  `json:"count"`
}
