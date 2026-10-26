package ginplus

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
)

// APIInfo 存储API的基本信息
type APIInfo struct {
	Path        string // API路径
	Method      string // HTTP方法
	Name        string // API名称
	Desc        string // API描述
	RequireAuth bool   // 是否需要认证
}

// Engine 是对gin.Engine的包装，增加API自动收集功能
type Engine struct {
	*gin.Engine
	apis          []APIInfo          // 用于收集API信息
	apisMux       sync.Mutex         // 用于保护apis在并发场景下的安全
	nameMap       map[string]APIInfo // 用于检查API名称唯一性
	pathMethodMap map[string]APIInfo // 用于检查path+method组合唯一性
}

// New 创建一个新的GinPlus引擎
func New(opts ...gin.OptionFunc) *Engine {
	return &Engine{
		Engine:        gin.New(opts...),
		apis:          make([]APIInfo, 0),
		nameMap:       make(map[string]APIInfo),
		pathMethodMap: make(map[string]APIInfo),
	}
}

// Default 创建一个带有默认中间件的GinPlus引擎
func Default(opts ...gin.OptionFunc) *Engine {
	e := gin.Default(opts...)
	e.Use(gin.Logger(), gin.Recovery())
	return &Engine{
		Engine:        e,
		apis:          make([]APIInfo, 0),
		nameMap:       make(map[string]APIInfo),
		pathMethodMap: make(map[string]APIInfo),
	}
}

// GET 是对gin.Engine.GET的增强，增加API信息收集
func (e *Engine) GET(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	e.collectAPI("GET", path, name, desc, false) // 默认不需要认证
	return e.Engine.GET(path, ConvertHandlers(handlers...)...)
}

// POST 是对gin.Engine.POST的增强，增加API信息收集
func (e *Engine) POST(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	e.collectAPI("POST", path, name, desc, false) // 默认不需要认证
	return e.Engine.POST(path, ConvertHandlers(handlers...)...)
}

// PUT 是对gin.Engine.PUT的增强，增加API信息收集
func (e *Engine) PUT(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	e.collectAPI("PUT", path, name, desc, false) // 默认不需要认证
	return e.Engine.PUT(path, ConvertHandlers(handlers...)...)
}

// DELETE 是对gin.Engine.DELETE的增强，增加API信息收集
func (e *Engine) DELETE(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	e.collectAPI("DELETE", path, name, desc, false) // 默认不需要认证
	return e.Engine.DELETE(path, ConvertHandlers(handlers...)...)
}

// PATCH 是对gin.Engine.PATCH的增强，增加API信息收集
func (e *Engine) PATCH(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	e.collectAPI("PATCH", path, name, desc, false) // 默认不需要认证
	return e.Engine.PATCH(path, ConvertHandlers(handlers...)...)
}

// HEAD 是对gin.Engine.HEAD的增强，增加API信息收集
func (e *Engine) HEAD(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	e.collectAPI("HEAD", path, name, desc, false) // 默认不需要认证
	return e.Engine.HEAD(path, ConvertHandlers(handlers...)...)
}

// OPTIONS 是对gin.Engine.OPTIONS的增强，增加API信息收集
func (e *Engine) OPTIONS(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	e.collectAPI("OPTIONS", path, name, desc, false) // 默认不需要认证
	return e.Engine.OPTIONS(path, ConvertHandlers(handlers...)...)
}

// Any 是对gin.Engine.Any的增强，增加API信息收集
func (e *Engine) Any(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	e.collectAPI("GET", path, name, desc, false)
	e.collectAPI("POST", path, name, desc, false)
	e.collectAPI("PUT", path, name, desc, false)
	e.collectAPI("DELETE", path, name, desc, false)
	e.collectAPI("PATCH", path, name, desc, false)
	e.collectAPI("HEAD", path, name, desc, false)
	e.collectAPI("OPTIONS", path, name, desc, false)
	return e.Engine.Any(path, ConvertHandlers(handlers...)...)
}

// Group 创建一个路由分组
func (e *Engine) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		RouterGroup: e.Engine.Group(relativePath, ConvertHandlers(handlers...)...),
		engine:      e,
		basePath:    relativePath,
		requireAuth: false, // 默认不需要认证
	}
}

// RouterGroup 是对gin.RouterGroup的包装，增加API自动收集功能
type RouterGroup struct {
	*gin.RouterGroup
	engine      *Engine
	basePath    string
	requireAuth bool // 是否需要认证
}

// Group 创建嵌套路由分组
func (g *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		RouterGroup: g.RouterGroup.Group(relativePath, ConvertHandlers(handlers...)...),
		engine:      g.engine,
		basePath:    g.basePath + relativePath,
		requireAuth: g.requireAuth, // 继承父分组的认证设置
	}
}

// RequireAuth 设置该路由组下的所有路由都需要认证
func (g *RouterGroup) RequireAuth() *RouterGroup {
	g.requireAuth = true
	return g
}

// GET 是对gin.RouterGroup.GET的增强，增加API信息收集
func (g *RouterGroup) GET(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	absolutePath := g.basePath + path
	g.engine.collectAPI("GET", absolutePath, name, desc, g.requireAuth)
	return g.RouterGroup.GET(path, ConvertHandlers(handlers...)...)
}

// POST 是对gin.RouterGroup.POST的增强，增加API信息收集
func (g *RouterGroup) POST(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	absolutePath := g.basePath + path
	g.engine.collectAPI("POST", absolutePath, name, desc, g.requireAuth)
	return g.RouterGroup.POST(path, ConvertHandlers(handlers...)...)
}

// PUT 是对gin.RouterGroup.PUT的增强，增加API信息收集
func (g *RouterGroup) PUT(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	absolutePath := g.basePath + path
	g.engine.collectAPI("PUT", absolutePath, name, desc, g.requireAuth)
	return g.RouterGroup.PUT(path, ConvertHandlers(handlers...)...)
}

// DELETE 是对gin.RouterGroup.DELETE的增强，增加API信息收集
func (g *RouterGroup) DELETE(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	absolutePath := g.basePath + path
	g.engine.collectAPI("DELETE", absolutePath, name, desc, g.requireAuth)
	return g.RouterGroup.DELETE(path, ConvertHandlers(handlers...)...)
}

// PATCH 是对gin.RouterGroup.PATCH的增强，增加API信息收集
func (g *RouterGroup) PATCH(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	absolutePath := g.basePath + path
	g.engine.collectAPI("PATCH", absolutePath, name, desc, g.requireAuth)
	return g.RouterGroup.PATCH(path, ConvertHandlers(handlers...)...)
}

// HEAD 是对gin.RouterGroup.HEAD的增强，增加API信息收集
func (g *RouterGroup) HEAD(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	absolutePath := g.basePath + path
	g.engine.collectAPI("HEAD", absolutePath, name, desc, g.requireAuth)
	return g.RouterGroup.HEAD(path, ConvertHandlers(handlers...)...)
}

// OPTIONS 是对gin.RouterGroup.OPTIONS的增强，增加API信息收集
func (g *RouterGroup) OPTIONS(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	absolutePath := g.basePath + path
	g.engine.collectAPI("OPTIONS", absolutePath, name, desc, g.requireAuth)
	return g.RouterGroup.OPTIONS(path, ConvertHandlers(handlers...)...)
}

// Any 是对gin.RouterGroup.Any的增强，增加API信息收集
func (g *RouterGroup) Any(path string, name, desc string, handlers ...HandlerFunc) gin.IRoutes {
	absolutePath := g.basePath + path
	g.engine.collectAPI("GET", absolutePath, name, desc, g.requireAuth)
	g.engine.collectAPI("POST", absolutePath, name, desc, g.requireAuth)
	g.engine.collectAPI("PUT", absolutePath, name, desc, g.requireAuth)
	g.engine.collectAPI("DELETE", absolutePath, name, desc, g.requireAuth)
	g.engine.collectAPI("PATCH", absolutePath, name, desc, g.requireAuth)
	g.engine.collectAPI("HEAD", absolutePath, name, desc, g.requireAuth)
	g.engine.collectAPI("OPTIONS", absolutePath, name, desc, g.requireAuth)
	return g.RouterGroup.Any(path, ConvertHandlers(handlers...)...)
}

// 内部函数，收集API信息
func (e *Engine) collectAPI(method, path, name, desc string, requireAuth bool) {
	e.apisMux.Lock()
	defer e.apisMux.Unlock()

	// 检查API名称唯一性
	if existingAPI, exists := e.nameMap[name]; exists {
		panic(fmt.Sprintf("路由名称重复: '%s' 已经被用于 %s %s, 不能再用于 %s %s",
			name, existingAPI.Method, existingAPI.Path, method, path))
	}

	// 检查path+method组合唯一性
	pathMethodKey := method + ":" + path
	if existingAPI, exists := e.pathMethodMap[pathMethodKey]; exists {
		panic(fmt.Sprintf("路由冲突: %s %s 已经被定义为名称 '%s', 不能再定义为 '%s'",
			method, path, existingAPI.Name, name))
	}

	// 创建新的API信息
	apiInfo := APIInfo{
		Method:      method,
		Path:        path,
		Name:        name,
		Desc:        desc,
		RequireAuth: requireAuth,
	}

	// 添加到映射和列表中
	e.nameMap[name] = apiInfo
	e.pathMethodMap[pathMethodKey] = apiInfo
	e.apis = append(e.apis, apiInfo)
}

// GetAPIs 返回收集到的所有API信息
func (e *Engine) GetAPIs() []APIInfo {
	e.apisMux.Lock()
	defer e.apisMux.Unlock()

	result := make([]APIInfo, len(e.apis))
	copy(result, e.apis)
	return result
}

// GetAuthAPIs 返回需要认证的API信息
func (e *Engine) GetAuthAPIs() []APIInfo {
	e.apisMux.Lock()
	defer e.apisMux.Unlock()

	var result []APIInfo
	for _, api := range e.apis {
		if api.RequireAuth {
			result = append(result, api)
		}
	}
	return result
}

// GetPublicAPIs 返回不需要认证的API信息
func (e *Engine) GetPublicAPIs() []APIInfo {
	e.apisMux.Lock()
	defer e.apisMux.Unlock()

	var result []APIInfo
	for _, api := range e.apis {
		if !api.RequireAuth {
			result = append(result, api)
		}
	}
	return result
}
