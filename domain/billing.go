package domain

import (
	"context"
	"errors"
	"time"
)

// 合記记账域：账本（Book）→ 账单（Bill）/ 分类（Category）/ 账单图片（BillImage）。
// 对外 JSON 字段名沿用客户端契约（_id / crt_time / upd_time / money 分为单位），
// 与 heji Android 客户端及历史数据保持一致。

// bill 类型：收入 +1 / 支出 -1
const (
	BillTypeIncome      = 1
	BillTypeExpenditure = -1

	// BillImageBucket 是账单图片在 FileRepository 中的桶名（磁盘实现下即子目录）。
	// 上传、下载与删除必须使用同一个值，因此定义在领域层。
	BillImageBucket = "heji"
)

var (
	ErrBookNotFound      = errors.New("账本不存在")
	ErrBillNotFound      = errors.New("账单不存在")
	ErrCategoryNotFound  = errors.New("分类不存在")
	ErrBillImageNotFound = errors.New("账单图片不存在")
	ErrNotBookMember     = errors.New("无权访问该账本")
	ErrBookNameExists    = errors.New("同名账本已存在")
	ErrInvalidBillTime   = errors.New("账单日期格式必须为 YYYY-MM-DD")
	ErrImageBillMismatch = errors.New("图片不属于该账单")
)

// Book 账本实体
type Book struct {
	ID        string     `json:"_id"`
	Name      string     `json:"name"`
	Type      string     `json:"type,omitempty"`
	Banner    string     `json:"banner,omitempty"`
	CrtUserID string     `json:"crt_user_id"`
	IsInitial bool       `json:"is_initial"`
	Members   []string   `json:"members,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"crt_time"`
	UpdatedAt time.Time  `json:"upd_time"`
}

// CreateBookRequest 创建账本请求
type CreateBookRequest struct {
	Name      string `json:"name" binding:"required,max=100"`
	Type      string `json:"type,omitempty" binding:"max=50"`
	Banner    string `json:"banner,omitempty" binding:"max=255"`
	IsInitial bool   `json:"is_initial,omitempty"`
}

// UpdateBookRequest 更新账本请求（指针字段区分“未提供”与“置空”）
type UpdateBookRequest struct {
	Name   *string `json:"name,omitempty" binding:"omitempty,max=100"`
	Type   *string `json:"type,omitempty" binding:"omitempty,max=50"`
	Banner *string `json:"banner,omitempty" binding:"omitempty,max=255"`
}

// Bill 账单实体
type Bill struct {
	ID          string     `json:"_id"`
	BookID      string     `json:"book_id"`
	Money       int64      `json:"money"` // 金额（分）
	Type        int        `json:"type"`  // 收入 1 / 支出 -1
	Category    string     `json:"category,omitempty"`
	CrtUser     string     `json:"crt_user"`
	CrtUserName string     `json:"crt_user_name,omitempty"`
	Time        string     `json:"time"` // 账单日期 YYYY-MM-DD，兼容客户端带时分秒的写法
	Remark      string     `json:"remark,omitempty"`
	Images      []string   `json:"images,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"crt_time"`
	UpdatedAt   time.Time  `json:"upd_time"`
}

// CreateBillRequest 创建账单请求。ID 由客户端生成（离线创建后同步），缺省时服务端生成。
type CreateBillRequest struct {
	ID       string   `json:"_id,omitempty"`
	BookID   string   `json:"book_id" binding:"required"`
	Money    int64    `json:"money" binding:"gte=0"`
	Type     int      `json:"type" binding:"required,oneof=1 -1"`
	Category string   `json:"category,omitempty" binding:"max=100"`
	Time     string   `json:"time" binding:"required"`
	Remark   string   `json:"remark,omitempty" binding:"max=500"`
	Images   []string `json:"images,omitempty"`
}

// UpdateBillRequest 更新账单请求
type UpdateBillRequest struct {
	Money    *int64   `json:"money,omitempty" binding:"omitempty,gte=0"`
	Type     *int     `json:"type,omitempty" binding:"omitempty,oneof=1 -1"`
	Category *string  `json:"category,omitempty" binding:"omitempty,max=100"`
	Time     *string  `json:"time,omitempty"`
	Remark   *string  `json:"remark,omitempty" binding:"omitempty,max=500"`
	Images   []string `json:"images,omitempty"`
}

// BillQueryFilter 账单查询过滤器。BookID 必填，由用例层校验调用者是否为账本成员。
type BillQueryFilter struct {
	QueryParams
	BookID   string `json:"book_id" form:"book_id"`
	Type     *int   `json:"type,omitempty" form:"type"`
	Category string `json:"category,omitempty" form:"category"`
	TimeFrom string `json:"time_from,omitempty" form:"time_from"`
	TimeTo   string `json:"time_to,omitempty" form:"time_to"`
}

// Category 分类实体
type Category struct {
	ID         string `json:"_id"`
	BookID     string `json:"book_id"`
	Name       string `json:"name"`
	Type       int    `json:"type"`
	Level      int    `json:"level"`
	Index      int    `json:"index"`
	ParentID   string `json:"parentId,omitempty"`
	CreateUser string `json:"create_user,omitempty"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	BookID   string `json:"book_id" binding:"required"`
	Name     string `json:"name" binding:"required,max=50"`
	Type     int    `json:"type" binding:"required,oneof=1 -1"`
	Level    int    `json:"level,omitempty"`
	Index    int    `json:"index,omitempty"`
	ParentID string `json:"parentId,omitempty"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,max=50"`
	Type     *int    `json:"type,omitempty" binding:"omitempty,oneof=1 -1"`
	Level    *int    `json:"level,omitempty"`
	Index    *int    `json:"index,omitempty"`
	ParentID *string `json:"parentId,omitempty"`
}

// BillImage 账单图片实体
type BillImage struct {
	ID     string `json:"_id"`
	BillID string `json:"bill_id"`
	Key    string `json:"key"`
	Ext    string `json:"ext,omitempty"`
	Size   int64  `json:"length,omitempty"`
	MD5    string `json:"md5,omitempty"`
}

// SyncChangesResponse 增量拉取响应；since 之后有变更的账本与账单（含已软删除的记录）。
type SyncChangesResponse struct {
	Books     []*Book `json:"books"`
	Bills     []*Bill `json:"bills"`
	HasMore   bool    `json:"has_more"`
	NextSince int64   `json:"next_since"`
}

// BillingNotifier 账本变更推送（由 internal/notifier 实现，未配置时为空实现）。
// 推送失败不影响业务写入：用例层只负责发出，不等待投递结果。
type BillingNotifier interface {
	NotifyBookChange(ctx context.Context, bookID, actorID, changeType, entity string, payload any)
}

// BookRepository 账本数据访问接口
type BookRepository interface {
	Create(ctx context.Context, book *Book) error
	GetByID(ctx context.Context, id string) (*Book, error)
	ListByUser(ctx context.Context, userID string) ([]*Book, error)
	Update(ctx context.Context, book *Book) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, bookID, userID string) error
	IsMember(ctx context.Context, bookID, userID string) (bool, error)
}

// BillRepository 账单数据访问接口
type BillRepository interface {
	Create(ctx context.Context, bill *Bill) error
	GetByID(ctx context.Context, id string) (*Bill, error)
	Query(ctx context.Context, filter BillQueryFilter) (*PagedResult[*Bill], error)
	Update(ctx context.Context, bill *Bill) error
	Delete(ctx context.Context, id string) error
}

// CategoryRepository 分类数据访问接口
type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id string) (*Category, error)
	ListByBookID(ctx context.Context, bookID string) ([]*Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id string) error
}

// BillImageRepository 账单图片数据访问接口
type BillImageRepository interface {
	Create(ctx context.Context, img *BillImage) error
	GetByID(ctx context.Context, id string) (*BillImage, error)
	ListByBillID(ctx context.Context, billID string) ([]*BillImage, error)
	Delete(ctx context.Context, id string) error
}

// SyncRepository 增量同步查询接口
type SyncRepository interface {
	QueryChanges(ctx context.Context, userID string, sinceMs int64, limit int) (*SyncChangesResponse, error)
}

// BillPagedResult 账单分页结果（Swagger 别名；账本列表不分页）
type BillPagedResult = PagedResult[*Bill]

// BookUseCase 账本业务逻辑接口
type BookUseCase interface {
	CreateBook(ctx context.Context, userID string, req *CreateBookRequest) (*Book, error)
	GetBook(ctx context.Context, userID, id string) (*Book, error)
	ListBooks(ctx context.Context, userID string) ([]*Book, error)
	UpdateBook(ctx context.Context, userID, id string, req *UpdateBookRequest) (*Book, error)
	DeleteBook(ctx context.Context, userID, id string) error
	JoinBook(ctx context.Context, userID, id string) error
	ShareBook(ctx context.Context, userID, id string) (string, error)
}

// BillUseCase 账单业务逻辑接口
type BillUseCase interface {
	CreateBill(ctx context.Context, userID string, req *CreateBillRequest) (*Bill, error)
	GetBill(ctx context.Context, userID, id string) (*Bill, error)
	ListBills(ctx context.Context, userID string, filter BillQueryFilter) (*BillPagedResult, error)
	UpdateBill(ctx context.Context, userID, id string, req *UpdateBillRequest) (*Bill, error)
	DeleteBill(ctx context.Context, userID, id string) error
}

// CategoryUseCase 分类业务逻辑接口
type CategoryUseCase interface {
	CreateCategory(ctx context.Context, userID string, req *CreateCategoryRequest) (*Category, error)
	CreateCategories(ctx context.Context, userID string, reqs []*CreateCategoryRequest) ([]*Category, error)
	GetCategory(ctx context.Context, userID, id string) (*Category, error)
	ListCategories(ctx context.Context, userID, bookID string) ([]*Category, error)
	UpdateCategory(ctx context.Context, userID, id string, req *UpdateCategoryRequest) (*Category, error)
	DeleteCategory(ctx context.Context, userID, id string) error
}

// BillImageUseCase 账单图片业务逻辑接口。上传/下载走 domain.FileRepository 抽象（磁盘或 MinIO）。
type BillImageUseCase interface {
	UploadImage(ctx context.Context, userID, billID string, img *BillImage, upload *UploadRequest) (*BillImage, error)
	GetImage(ctx context.Context, userID, id string) (*BillImage, error)
	ListByBillID(ctx context.Context, userID, billID string) ([]*BillImage, error)
	DeleteImage(ctx context.Context, userID, billID, imageID string) error
}

// SyncUseCase 增量同步业务逻辑接口
type SyncUseCase interface {
	GetChanges(ctx context.Context, userID string, sinceMs int64, limit int) (*SyncChangesResponse, error)
}
