package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"shadmin/domain"
	"shadmin/internal/constants"

	"github.com/gin-gonic/gin"
)

// ImageController 账单图片接口。图片内容经 domain.FileRepository 读写，
// 存储后端（磁盘 / MinIO）由配置决定。
type ImageController struct {
	ImageUseCase domain.BillImageUseCase
	FileStorage  domain.FileRepository
}

// UploadImage 上传账单图片
//
//	@Summary		Upload a bill image
//	@Description	Store an image file for a bill; the caller must be a member of the bill's book
//	@Tags			Image
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file	formData	file								true	"Image file"
//	@Param			billId	query		string								true	"Bill ID"
//	@Param			_id		query		string								false	"Client-generated image ID"
//	@Success		201		{object}	domain.Response{data=domain.BillImage}	"Image stored"
//	@Failure		400		{object}	domain.Response						"Missing file or billId"
//	@Failure		403		{object}	domain.Response						"Not a member of this book"
//	@Router			/image/upload [post]
func (ic *ImageController) UploadImage(c *gin.Context) {
	billID := c.Query("billId")
	if billID == "" {
		c.JSON(http.StatusBadRequest, domain.RespError("billId is required"))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.RespError(fmt.Errorf("read uploaded file: %w", err)))
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	upload := &domain.UploadRequest{
		Bucket:      domain.BillImageBucket,
		ObjectName:  fmt.Sprintf("bills/%s/%s", billID, filename),
		Reader:      file,
		Size:        header.Size,
		ContentType: header.Header.Get("Content-Type"),
	}

	img := &domain.BillImage{
		ID:  c.Query("_id"),
		Ext: strings.ToLower(filepath.Ext(filename)),
	}

	result, err := ic.ImageUseCase.UploadImage(c.Request.Context(), c.GetString(constants.UserID), billID, img, upload)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, domain.RespSuccess(result))
}

// ListBillImages 账单图片列表
//
//	@Summary		List images of a bill
//	@Description	Return image metadata for one bill
//	@Tags			Image
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			bill_id	query		string									true	"Bill ID"
//	@Success		200		{object}	domain.Response{data=[]domain.BillImage}	"Images retrieved"
//	@Failure		400		{object}	domain.Response							"Missing bill_id"
//	@Failure		403		{object}	domain.Response							"Not a member of this book"
//	@Router			/image/list [get]
func (ic *ImageController) ListBillImages(c *gin.Context) {
	billID := c.Query("bill_id")
	if billID == "" {
		c.JSON(http.StatusBadRequest, domain.RespError("bill_id is required"))
		return
	}

	images, err := ic.ImageUseCase.ListByBillID(c.Request.Context(), c.GetString(constants.UserID), billID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(images))
}

// GetImage 下载图片内容
//
//	@Summary		Download a bill image
//	@Description	Stream the stored image bytes; the caller must be a member of the bill's book
//	@Tags			Image
//	@Produce		image/jpeg
//	@Produce		image/png
//	@Security		BearerAuth
//	@Param			id	path		string			true	"Image ID"
//	@Success		200	{file}		binary			"Image bytes"
//	@Failure		403	{object}	domain.Response	"Not a member of this book"
//	@Failure		404	{object}	domain.Response	"Image not found"
//	@Router			/image/{id} [get]
func (ic *ImageController) GetImage(c *gin.Context) {
	img, err := ic.ImageUseCase.GetImage(c.Request.Context(), c.GetString(constants.UserID), c.Param("id"))
	if err != nil {
		RespondError(c, err)
		return
	}

	reader, err := ic.FileStorage.Download(c.Request.Context(), &domain.DownloadRequest{
		Bucket:     domain.BillImageBucket,
		ObjectName: img.Key,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, domain.RespError("image content not found"))
		return
	}
	defer reader.Close()

	c.DataFromReader(http.StatusOK, img.Size, imageContentType(img.Ext), reader, nil)
}

// DeleteImage 删除账单图片
//
//	@Summary		Delete a bill image
//	@Description	Remove an image record and its stored file
//	@Tags			Image
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			billId	query		string			true	"Bill ID"
//	@Param			imageId	query		string			true	"Image ID"
//	@Success		200		{object}	domain.Response	"Image deleted"
//	@Failure		400		{object}	domain.Response	"Missing billId or imageId"
//	@Failure		403		{object}	domain.Response	"Not a member of this book"
//	@Router			/image [delete]
func (ic *ImageController) DeleteImage(c *gin.Context) {
	billID := c.Query("billId")
	imageID := c.Query("imageId")
	if billID == "" || imageID == "" {
		c.JSON(http.StatusBadRequest, domain.RespError("billId and imageId are required"))
		return
	}

	if err := ic.ImageUseCase.DeleteImage(c.Request.Context(), c.GetString(constants.UserID), billID, imageID); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(nil))
}

// imageContentType 按扩展名给出 Content-Type；未知类型回落到二进制流。
func imageContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
