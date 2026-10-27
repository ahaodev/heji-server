package controller

import (
	"net/http"

	"shadmin/domain"
	"shadmin/internal/constants"

	"github.com/gin-gonic/gin"
)

// CategoryController 分类接口。
type CategoryController struct {
	CategoryUseCase domain.CategoryUseCase
}

// CreateCategory 创建分类
//
//	@Summary		Create a category
//	@Description	Create one category inside a book the caller belongs to
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CreateCategoryRequest			true	"Category data"
//	@Success		201		{object}	domain.Response{data=domain.Category}	"Category created"
//	@Failure		400		{object}	domain.Response							"Invalid request"
//	@Failure		403		{object}	domain.Response							"Not a member of this book"
//	@Router			/category [post]
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	var req domain.CreateCategoryRequest
	if !MustBindJSON(c, &req) {
		return
	}

	category, err := cc.CategoryUseCase.CreateCategory(c.Request.Context(), c.GetString(constants.UserID), &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, domain.RespSuccess(category))
}

// CreateCategories 批量创建分类
//
//	@Summary		Create categories in bulk
//	@Description	Create several categories in one request (used by client-side sync)
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		[]domain.CreateCategoryRequest		true	"Category list"
//	@Success		201		{object}	domain.Response{data=[]domain.Category}	"Categories created"
//	@Failure		400		{object}	domain.Response							"Invalid request"
//	@Failure		403		{object}	domain.Response							"Not a member of this book"
//	@Router			/category/batch [post]
func (cc *CategoryController) CreateCategories(c *gin.Context) {
	var reqs []*domain.CreateCategoryRequest
	if !MustBindJSON(c, &reqs) {
		return
	}
	if len(reqs) == 0 {
		c.JSON(http.StatusBadRequest, domain.RespError("category list is empty"))
		return
	}

	categories, err := cc.CategoryUseCase.CreateCategories(c.Request.Context(), c.GetString(constants.UserID), reqs)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, domain.RespSuccess(categories))
}

// ListCategories 账本下的分类列表
//
//	@Summary		List categories of a book
//	@Description	Return every category of one book, ordered by index
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			book_id	query		string									true	"Book ID"
//	@Success		200		{object}	domain.Response{data=[]domain.Category}	"Categories retrieved"
//	@Failure		400		{object}	domain.Response							"Missing book_id"
//	@Failure		403		{object}	domain.Response							"Not a member of this book"
//	@Router			/category [get]
func (cc *CategoryController) ListCategories(c *gin.Context) {
	bookID := c.Query("book_id")
	if bookID == "" {
		c.JSON(http.StatusBadRequest, domain.RespError("book_id is required"))
		return
	}

	categories, err := cc.CategoryUseCase.ListCategories(c.Request.Context(), c.GetString(constants.UserID), bookID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(categories))
}

// GetCategory 分类详情
//
//	@Summary		Get a category
//	@Description	Retrieve one category; the caller must be a member of its book
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string									true	"Category ID"
//	@Success		200	{object}	domain.Response{data=domain.Category}	"Category retrieved"
//	@Failure		403	{object}	domain.Response							"Not a member of this book"
//	@Failure		404	{object}	domain.Response							"Category not found"
//	@Router			/category/{id} [get]
func (cc *CategoryController) GetCategory(c *gin.Context) {
	category, err := cc.CategoryUseCase.GetCategory(c.Request.Context(), c.GetString(constants.UserID), c.Param("id"))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(category))
}

// UpdateCategory 更新分类
//
//	@Summary		Update a category
//	@Description	Partial update of a category; the caller must be a member of its book
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string									true	"Category ID"
//	@Param			request	body		domain.UpdateCategoryRequest			true	"Category update data"
//	@Success		200		{object}	domain.Response{data=domain.Category}	"Category updated"
//	@Failure		400		{object}	domain.Response							"Invalid request"
//	@Failure		403		{object}	domain.Response							"Not a member of this book"
//	@Failure		404		{object}	domain.Response							"Category not found"
//	@Router			/category/{id} [put]
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	var req domain.UpdateCategoryRequest
	if !MustBindJSON(c, &req) {
		return
	}

	category, err := cc.CategoryUseCase.UpdateCategory(c.Request.Context(), c.GetString(constants.UserID), c.Param("id"), &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(category))
}

// DeleteCategory 删除分类
//
//	@Summary		Delete a category
//	@Description	Hard delete a category; the caller must be a member of its book
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string			true	"Category ID"
//	@Success		200	{object}	domain.Response	"Category deleted"
//	@Failure		403	{object}	domain.Response	"Not a member of this book"
//	@Failure		404	{object}	domain.Response	"Category not found"
//	@Router			/category/{id} [delete]
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	if err := cc.CategoryUseCase.DeleteCategory(c.Request.Context(), c.GetString(constants.UserID), c.Param("id")); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(nil))
}
