package controller

import (
	"net/http"

	"shadmin/domain"
	"shadmin/internal/constants"

	"github.com/gin-gonic/gin"
)

// BookController 账本接口。
type BookController struct {
	BookUseCase domain.BookUseCase
}

// CreateBook 创建账本
//
//	@Summary		Create a book
//	@Description	Create a ledger book owned by the current user
//	@Tags			Book
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CreateBookRequest			true	"Book data"
//	@Success		201		{object}	domain.Response{data=domain.Book}	"Book created"
//	@Failure		400		{object}	domain.Response						"Invalid request"
//	@Failure		409		{object}	domain.Response						"Duplicate book name"
//	@Router			/book [post]
func (bc *BookController) CreateBook(c *gin.Context) {
	var req domain.CreateBookRequest
	if !MustBindJSON(c, &req) {
		return
	}

	book, err := bc.BookUseCase.CreateBook(c.Request.Context(), c.GetString(constants.UserID), &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, domain.RespSuccess(book))
}

// ListBooks 当前用户可访问的账本列表
//
//	@Summary		List books of the current user
//	@Description	Return every book the current user is a member of, newest first
//	@Tags			Book
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	domain.Response{data=[]domain.Book}	"Books retrieved"
//	@Failure		500	{object}	domain.Response						"Internal server error"
//	@Router			/book [get]
func (bc *BookController) ListBooks(c *gin.Context) {
	books, err := bc.BookUseCase.ListBooks(c.Request.Context(), c.GetString(constants.UserID))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(books))
}

// GetBook 账本详情
//
//	@Summary		Get a book
//	@Description	Retrieve one book; the caller must be a member
//	@Tags			Book
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string								true	"Book ID"
//	@Success		200	{object}	domain.Response{data=domain.Book}	"Book retrieved"
//	@Failure		403	{object}	domain.Response						"Not a member of this book"
//	@Failure		404	{object}	domain.Response						"Book not found"
//	@Router			/book/{id} [get]
func (bc *BookController) GetBook(c *gin.Context) {
	book, err := bc.BookUseCase.GetBook(c.Request.Context(), c.GetString(constants.UserID), c.Param("id"))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(book))
}

// UpdateBook 更新账本
//
//	@Summary		Update a book
//	@Description	Update name, type or banner of a book; the caller must be a member
//	@Tags			Book
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string								true	"Book ID"
//	@Param			request	body		domain.UpdateBookRequest			true	"Book update data"
//	@Success		200		{object}	domain.Response{data=domain.Book}	"Book updated"
//	@Failure		400		{object}	domain.Response						"Invalid request"
//	@Failure		403		{object}	domain.Response						"Not a member of this book"
//	@Failure		404		{object}	domain.Response						"Book not found"
//	@Router			/book/{id} [put]
func (bc *BookController) UpdateBook(c *gin.Context) {
	var req domain.UpdateBookRequest
	if !MustBindJSON(c, &req) {
		return
	}

	book, err := bc.BookUseCase.UpdateBook(c.Request.Context(), c.GetString(constants.UserID), c.Param("id"), &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(book))
}

// DeleteBook 删除账本（软删除，仅创建人）
//
//	@Summary		Delete a book
//	@Description	Soft delete a book; only its creator may delete it
//	@Tags			Book
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string			true	"Book ID"
//	@Success		200	{object}	domain.Response	"Book deleted"
//	@Failure		403	{object}	domain.Response	"Only the creator may delete this book"
//	@Failure		404	{object}	domain.Response	"Book not found"
//	@Router			/book/{id} [delete]
func (bc *BookController) DeleteBook(c *gin.Context) {
	if err := bc.BookUseCase.DeleteBook(c.Request.Context(), c.GetString(constants.UserID), c.Param("id")); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(nil))
}

// JoinBook 加入账本
//
//	@Summary		Join a book
//	@Description	Join a book using its share code (the book ID)
//	@Tags			Book
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string			true	"Book ID (share code)"
//	@Success		200	{object}	domain.Response	"Joined"
//	@Failure		404	{object}	domain.Response	"Book not found"
//	@Router			/book/{id}/join [post]
func (bc *BookController) JoinBook(c *gin.Context) {
	if err := bc.BookUseCase.JoinBook(c.Request.Context(), c.GetString(constants.UserID), c.Param("id")); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(nil))
}

// ShareBook 生成分享码
//
//	@Summary		Share a book
//	@Description	Return the share code for a book the caller is a member of
//	@Tags			Book
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"Book ID"
//	@Success		200	{object}	domain.Response{data=string}	"Share code"
//	@Failure		403	{object}	domain.Response		"Not a member of this book"
//	@Failure		404	{object}	domain.Response		"Book not found"
//	@Router			/book/{id}/share [post]
func (bc *BookController) ShareBook(c *gin.Context) {
	code, err := bc.BookUseCase.ShareBook(c.Request.Context(), c.GetString(constants.UserID), c.Param("id"))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(code))
}
