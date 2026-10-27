package controller

import (
	"fmt"
	"net/http"

	"shadmin/domain"
	"shadmin/internal/constants"

	"github.com/gin-gonic/gin"
)

// BillController 账单接口。
type BillController struct {
	BillUseCase domain.BillUseCase
}

// CreateBill 创建账单
//
//	@Summary		Create a bill
//	@Description	Create a bill inside a book the caller belongs to; client-generated IDs are accepted for offline sync
//	@Tags			Bill
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CreateBillRequest			true	"Bill data"
//	@Success		201		{object}	domain.Response{data=domain.Bill}	"Bill created"
//	@Failure		400		{object}	domain.Response						"Invalid request"
//	@Failure		403		{object}	domain.Response						"Not a member of this book"
//	@Router			/bill [post]
func (bc *BillController) CreateBill(c *gin.Context) {
	var req domain.CreateBillRequest
	if !MustBindJSON(c, &req) {
		return
	}

	bill, err := bc.BillUseCase.CreateBill(c.Request.Context(), c.GetString(constants.UserID), &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, domain.RespSuccess(bill))
}

// ListBills 账单列表
//
//	@Summary		List bills of a book
//	@Description	Paginated bills of one book, filterable by type, category and date range
//	@Tags			Bill
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			book_id		query		string												true	"Book ID"
//	@Param			type		query		int													false	"1 income, -1 expenditure"
//	@Param			category	query		string												false	"Category name"
//	@Param			time_from	query		string												false	"Start date (YYYY-MM-DD)"
//	@Param			time_to		query		string												false	"End date (YYYY-MM-DD)"
//	@Param			page		query		int													false	"Page number (default: 1)"
//	@Param			page_size	query		int													false	"Page size (default: 10)"
//	@Success		200			{object}	domain.Response{data=domain.BillPagedResult}		"Bills retrieved"
//	@Failure		400			{object}	domain.Response										"Missing book_id"
//	@Failure		403			{object}	domain.Response										"Not a member of this book"
//	@Router			/bill [get]
func (bc *BillController) ListBills(c *gin.Context) {
	var filter domain.BillQueryFilter
	if !MustBindQuery(c, &filter) {
		return
	}
	if filter.BookID == "" {
		c.JSON(http.StatusBadRequest, domain.RespError("book_id is required"))
		return
	}

	result, err := bc.BillUseCase.ListBills(c.Request.Context(), c.GetString(constants.UserID), filter)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(result))
}

// GetBill 账单详情
//
//	@Summary		Get a bill
//	@Description	Retrieve one bill; the caller must be a member of its book
//	@Tags			Bill
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string								true	"Bill ID"
//	@Success		200	{object}	domain.Response{data=domain.Bill}	"Bill retrieved"
//	@Failure		403	{object}	domain.Response						"Not a member of this book"
//	@Failure		404	{object}	domain.Response						"Bill not found"
//	@Router			/bill/{id} [get]
func (bc *BillController) GetBill(c *gin.Context) {
	bill, err := bc.BillUseCase.GetBill(c.Request.Context(), c.GetString(constants.UserID), c.Param("id"))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(bill))
}

// UpdateBill 更新账单
//
//	@Summary		Update a bill
//	@Description	Partial update of a bill; the caller must be a member of its book
//	@Tags			Bill
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string								true	"Bill ID"
//	@Param			request	body		domain.UpdateBillRequest			true	"Bill update data"
//	@Success		200		{object}	domain.Response{data=domain.Bill}	"Bill updated"
//	@Failure		400		{object}	domain.Response						"Invalid request"
//	@Failure		403		{object}	domain.Response						"Not a member of this book"
//	@Failure		404		{object}	domain.Response						"Bill not found"
//	@Router			/bill/{id} [put]
func (bc *BillController) UpdateBill(c *gin.Context) {
	var req domain.UpdateBillRequest
	if !MustBindJSON(c, &req) {
		return
	}

	bill, err := bc.BillUseCase.UpdateBill(c.Request.Context(), c.GetString(constants.UserID), c.Param("id"), &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(bill))
}

// DeleteBill 删除账单（软删除）
//
//	@Summary		Delete a bill
//	@Description	Soft delete a bill; the caller must be a member of its book
//	@Tags			Bill
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string			true	"Bill ID"
//	@Success		200	{object}	domain.Response	"Bill deleted"
//	@Failure		403	{object}	domain.Response	"Not a member of this book"
//	@Failure		404	{object}	domain.Response	"Bill not found"
//	@Router			/bill/{id} [delete]
func (bc *BillController) DeleteBill(c *gin.Context) {
	if err := bc.BillUseCase.DeleteBill(c.Request.Context(), c.GetString(constants.UserID), c.Param("id")); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(nil))
}

// ExportBills 导出账单 CSV
//
//	@Summary		Export bills as CSV
//	@Description	Download a book's bills as UTF-8 CSV, optionally limited to a year or a year+month
//	@Tags			Bill
//	@Produce		text/csv
//	@Security		BearerAuth
//	@Param			book_id	query		string			true	"Book ID"
//	@Param			year	query		string			false	"Year (YYYY)"
//	@Param			month	query		string			false	"Month (MM), requires year"
//	@Success		200		{file}		binary			"CSV file"
//	@Failure		400		{object}	domain.Response	"Missing book_id"
//	@Failure		403		{object}	domain.Response	"Not a member of this book"
//	@Router			/bill/export [get]
func (bc *BillController) ExportBills(c *gin.Context) {
	bookID := c.Query("book_id")
	if bookID == "" {
		c.JSON(http.StatusBadRequest, domain.RespError("book_id is required"))
		return
	}

	// 导出不分页：一次取名单本账本的全部账单（上限为 QueryParams 的最大页大小）
	timeFrom, timeTo := exportRange(c.Query("year"), c.Query("month"))
	filter := domain.BillQueryFilter{
		BookID:   bookID,
		TimeFrom: timeFrom,
		TimeTo:   timeTo,
	}
	filter.PageSize = domain.MaxPageSize

	result, err := bc.BillUseCase.ListBills(c.Request.Context(), c.GetString(constants.UserID), filter)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="bills.csv"`)
	// UTF-8 BOM，便于 Excel 正确识别中文
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	_, _ = c.Writer.WriteString("日期,类型,分类,金额,备注\n")
	for _, bill := range result.List {
		typeName := "支出"
		if bill.Type == domain.BillTypeIncome {
			typeName = "收入"
		}
		_, _ = fmt.Fprintf(c.Writer, "%s,%s,%s,%.2f,%s\n",
			bill.Time, typeName, bill.Category, float64(bill.Money)/100, csvSafe(bill.Remark))
	}
}

// exportRange 把 year / month 转成日期区间；缺省时返回空串（即不过滤）。
func exportRange(year, month string) (string, string) {
	if year == "" {
		return "", ""
	}
	if month == "" {
		return year + "-01-01", year + "-12-31"
	}
	start := year + "-" + month + "-01"
	// 用字符串上界而不是计算月末天数：YYYY-MM-31 覆盖所有月份，且与 time 列同为字符串比较
	return start, year + "-" + month + "-31"
}

// csvSafe 去掉会破坏 CSV 结构的字符，避免一条账单拆成多行。
func csvSafe(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\n' || r == '\r' || r == ',' {
			r = ' '
		}
		out = append(out, r)
	}
	return string(out)
}
