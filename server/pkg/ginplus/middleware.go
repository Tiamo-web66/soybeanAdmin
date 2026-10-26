package ginplus

import "github.com/gin-gonic/gin"

// MiddlewareFunc 定义了中间件函数类型
type MiddlewareFunc func(*Context) bool

// GinMiddleware 将自定义中间件转换为 Gin 中间件
func GinMiddleware(middleware MiddlewareFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建 ginplus.Context
		ctx := NewContext(c)

		// 执行中间件
		if !middleware(ctx) {
			// 如果中间件返回 false，则终止请求处理
			c.Abort()
			return
		}

		// 继续处理下一个中间件或处理器
		c.Next()
	}
}

// Use 使用自定义中间件
func (g *RouterGroup) Use(middlewares ...interface{}) *RouterGroup {
	ginMiddlewares := make([]gin.HandlerFunc, 0, len(middlewares))

	for _, middleware := range middlewares {
		switch m := middleware.(type) {
		case MiddlewareFunc:
			// 如果是 ginplus.MiddlewareFunc 类型
			ginMiddlewares = append(ginMiddlewares, GinMiddleware(m))
		case gin.HandlerFunc:
			// 如果是 gin.HandlerFunc 类型
			ginMiddlewares = append(ginMiddlewares, m)
		case func(*Context) bool:
			// 如果是匿名函数但签名与 MiddlewareFunc 相同
			ginMiddlewares = append(ginMiddlewares, GinMiddleware(m))
		case func(*gin.Context):
			// 如果是匿名函数但签名与 gin.HandlerFunc 相同
			ginMiddlewares = append(ginMiddlewares, m)
		default:
			panic("不支持的中间件类型")
		}
	}

	g.RouterGroup.Use(ginMiddlewares...)
	return g
}

// Use 使用自定义中间件
func (e *Engine) Use(middlewares ...interface{}) *Engine {
	ginMiddlewares := make([]gin.HandlerFunc, 0, len(middlewares))

	for _, middleware := range middlewares {
		switch m := middleware.(type) {
		case MiddlewareFunc:
			// 如果是 ginplus.MiddlewareFunc 类型
			ginMiddlewares = append(ginMiddlewares, GinMiddleware(m))
		case gin.HandlerFunc:
			// 如果是 gin.HandlerFunc 类型
			ginMiddlewares = append(ginMiddlewares, m)
		case func(*Context) bool:
			// 如果是匿名函数但签名与 MiddlewareFunc 相同
			ginMiddlewares = append(ginMiddlewares, GinMiddleware(m))
		case func(*gin.Context):
			// 如果是匿名函数但签名与 gin.HandlerFunc 相同
			ginMiddlewares = append(ginMiddlewares, m)
		default:
			panic("不支持的中间件类型")
		}
	}

	e.Engine.Use(ginMiddlewares...)
	return e
}

// WrapGinMiddleware 将 gin 中间件包装为 ginplus 中间件
func WrapGinMiddleware(middleware gin.HandlerFunc) MiddlewareFunc {
	return func(c *Context) bool {
		middleware(c.Context)
		return !c.Context.IsAborted()
	}
}
