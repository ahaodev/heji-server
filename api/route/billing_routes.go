package route

import (
	"shadmin/api/middleware"

	"github.com/gin-gonic/gin"
)

// setupBillingRoutes 注册合記记账域路由：账本 / 账单 / 分类 / 图片 / 增量同步。
// 所有路由都用 Casbin 校验 (userID, path, method)，资源在启动时随路由扫描入库。
func (pr *ProtectedRoutes) setupBillingRoutes(router *gin.RouterGroup, casbinMiddleware *middleware.CasbinMiddleware) {
	pr.setupBookRoutes(router, casbinMiddleware)
	pr.setupBillRoutes(router, casbinMiddleware)
	pr.setupCategoryRoutes(router, casbinMiddleware)
	pr.setupImageRoutes(router, casbinMiddleware)
	pr.setupSyncRoutes(router, casbinMiddleware)
}

func (pr *ProtectedRoutes) setupBookRoutes(router *gin.RouterGroup, casbinMiddleware *middleware.CasbinMiddleware) {
	group := router.Group("/book")
	group.Use(casbinMiddleware.CheckAPIPermission())
	controller := pr.factory.CreateBookController()

	group.GET("", controller.ListBooks)          // GET /api/v1/book
	group.POST("", controller.CreateBook)        // POST /api/v1/book
	group.GET("/:id", controller.GetBook)        // GET /api/v1/book/:id
	group.PUT("/:id", controller.UpdateBook)     // PUT /api/v1/book/:id
	group.DELETE("/:id", controller.DeleteBook)  // DELETE /api/v1/book/:id
	group.POST("/:id/join", controller.JoinBook) // POST /api/v1/book/:id/join
	group.POST("/:id/share", controller.ShareBook)
}

func (pr *ProtectedRoutes) setupBillRoutes(router *gin.RouterGroup, casbinMiddleware *middleware.CasbinMiddleware) {
	group := router.Group("/bill")
	group.Use(casbinMiddleware.CheckAPIPermission())
	controller := pr.factory.CreateBillController()

	group.GET("", controller.ListBills)          // GET /api/v1/bill
	group.POST("", controller.CreateBill)        // POST /api/v1/bill
	group.GET("/export", controller.ExportBills) // GET /api/v1/bill/export
	group.GET("/:id", controller.GetBill)        // GET /api/v1/bill/:id
	group.PUT("/:id", controller.UpdateBill)     // PUT /api/v1/bill/:id
	group.DELETE("/:id", controller.DeleteBill)  // DELETE /api/v1/bill/:id
}

func (pr *ProtectedRoutes) setupCategoryRoutes(router *gin.RouterGroup, casbinMiddleware *middleware.CasbinMiddleware) {
	group := router.Group("/category")
	group.Use(casbinMiddleware.CheckAPIPermission())
	controller := pr.factory.CreateCategoryController()

	group.GET("", controller.ListCategories)          // GET /api/v1/category
	group.POST("", controller.CreateCategory)         // POST /api/v1/category
	group.POST("/batch", controller.CreateCategories) // POST /api/v1/category/batch
	group.GET("/:id", controller.GetCategory)         // GET /api/v1/category/:id
	group.PUT("/:id", controller.UpdateCategory)      // PUT /api/v1/category/:id
	group.DELETE("/:id", controller.DeleteCategory)   // DELETE /api/v1/category/:id
}

func (pr *ProtectedRoutes) setupImageRoutes(router *gin.RouterGroup, casbinMiddleware *middleware.CasbinMiddleware) {
	group := router.Group("/image")
	group.Use(casbinMiddleware.CheckAPIPermission())
	controller := pr.factory.CreateImageController()

	group.POST("/upload", controller.UploadImage) // POST /api/v1/image/upload
	group.GET("/list", controller.ListBillImages) // GET /api/v1/image/list
	group.GET("/:id", controller.GetImage)        // GET /api/v1/image/:id
	group.DELETE("", controller.DeleteImage)      // DELETE /api/v1/image
}

func (pr *ProtectedRoutes) setupSyncRoutes(router *gin.RouterGroup, casbinMiddleware *middleware.CasbinMiddleware) {
	group := router.Group("/sync")
	group.Use(casbinMiddleware.CheckAPIPermission())
	controller := pr.factory.CreateSyncController()

	group.GET("/changes", controller.GetChanges) // GET /api/v1/sync/changes
}

// setupClientCompatRoutes 注册合記 Android 客户端使用的兼容接口：
// 免鉴权注册（/Register）与手机号登录（/Login），以及 MQTT broker 地址下发。
// 客户端以 tel+password 登录并期望 data 直接是 JWT 字符串，与新前端的登录契约不同，
// 因此单独保留这两个端点而不改动 /auth/*。
func (pr *PublicRoutes) setupClientCompatRoutes(router *gin.RouterGroup) {
	controller := pr.factory.CreateClientAuthController()
	router.POST("/Register", controller.Register)
	router.POST("/Login", controller.Login)
	router.GET("/mqtt/broker", controller.MQTTBroker)
}
