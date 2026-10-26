package ginplus

import "github.com/gin-gonic/gin"

// HandlerFunc 定义了处理请求的函数类型，使用 ginplus.Context 而不是 gin.Context
type HandlerFunc func(*Context)

// Router 定义了路由注册方法接口
type Router interface {
	// HTTP方法
	GET(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes
	POST(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes
	PUT(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes
	DELETE(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes
	PATCH(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes
	HEAD(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes
	OPTIONS(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes
	Any(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes

	// 组路由
	Group(relativePath string, handlers ...HandlerFunc) *RouterGroup
}

// RouterRegistrar 路由注册器接口
type RouterRegistrar interface {
	Register(router Router)
}

// RegisterRouters 注册多个路由注册器
func RegisterRouters(r Router, registrars ...RouterRegistrar) {
	for _, registrar := range registrars {
		registrar.Register(r)
	}
}

// HandlerFuncToGin 将 ginplus.HandlerFunc 转换为 gin.HandlerFunc
func HandlerFuncToGin(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(NewContext(c))
	}
}

// ConvertHandlers 将多个 ginplus.HandlerFunc 转换为多个 gin.HandlerFunc
func ConvertHandlers(handlers ...HandlerFunc) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = HandlerFuncToGin(handler)
	}
	return ginHandlers
}

// GinHandlerToHandlerFunc 将 gin.HandlerFunc 转换为 ginplus.HandlerFunc
// 用于兼容现有的 gin 处理器
func GinHandlerToHandlerFunc(handler gin.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		handler(c.Context)
	}
}

// WrapGinHandler 将 gin.HandlerFunc 包装为 ginplus.HandlerFunc
func WrapGinHandler(handler gin.HandlerFunc) HandlerFunc {
	return GinHandlerToHandlerFunc(handler)
}
