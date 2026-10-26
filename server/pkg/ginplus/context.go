package ginplus

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	SuccessCode   = "0000"
	ParamsErrCode = 9001           //参数解析失败错误码
	NormalErrCode = 0001           //通用业务错误码
	AutoErrCode   = 8889           //认证错误
	PageTag       = "^_^page^_^"   //存放page数据的key
	UserIdTag     = "^_^userId^_^" //用户id
)

// Context 是对gin.Context的包装，提供更多功能
type Context struct {
	*gin.Context
}

// NewContext 从 gin.Context 创建 ginplus.Context
func NewContext(c *gin.Context) *Context {
	return &Context{Context: c}
}

// Success 返回成功响应
func (c *Context) Success(data any) {
	c.SuccessMsg("成功", data)
}

// SuccessMsg 返回带自定义消息的成功响应
func (c *Context) SuccessMsg(msg string, data any) {
	c.JSON(http.StatusOK, Resp[any]{
		Code: SuccessCode,
		Msg:  msg,
		Data: data,
	})
}

// SuccessWithPage 带分页的响应
func (c *Context) SuccessWithPage(total int64, records any) {
	if total == 0 {
		records = make([]any, 0)
	}
	c.SuccessWithPageMsg(total, records, "成功")
}

// SuccessWithPageMsg 返回带分页和自定义消息的成功响应
func (c *Context) SuccessWithPageMsg(total int64, records any, msg string) {
	page := c.getPage()
	c.JSON(http.StatusOK, Resp[PageVO[any]]{
		Code: "0000",
		Msg:  msg,
		Data: PageVO[any]{
			Current: page.Current,
			Size:    page.Size,
			Total:   total,
			Records: records,
		},
	})
}

// FailAuth 认证失败
func (c *Context) FailAuth(msg string) {
	c.FailCode(AutoErrCode, msg)
}

// Fail 返回失败响应
func (c *Context) Fail(msg string) {
	c.FailCode(NormalErrCode, msg)
}

// FailCode 返回失败响应
func (c *Context) FailCode(code int, msg string) {
	c.JSON(http.StatusOK, Resp[any]{
		Code: strconv.Itoa(code),
		Msg:  msg,
		Data: nil,
	})
}

// FailWithData 返回带数据的失败响应
func (c *Context) FailWithData(code int, msg string, data any) {
	c.JSON(http.StatusOK, Resp[any]{
		Code: strconv.Itoa(code),
		Msg:  msg,
		Data: data,
	})
}

// ParseArrayParams 解析前端传递的数组参数（格式如：key[0]=value1&key[1]=value2）
func (c *Context) ParseArrayParams(paramName string) []string {
	var result []string
	// 获取所有查询参数
	queryMap := c.Request.URL.Query()

	// 遍历所有参数，查找符合 paramName[数字] 格式的参数
	prefix := paramName + "["
	for key, values := range queryMap {
		if strings.HasPrefix(key, prefix) && strings.HasSuffix(key, "]") {
			// 获取参数的第一个值（通常只有一个）
			if len(values) > 0 {
				result = append(result, values[0])
			}
		}
	}

	return result
}

// ParseArrayParamsInt 解析前端传递的整数数组参数
func (c *Context) ParseArrayParamsInt(paramName string) ([]int, error) {
	strValues := c.ParseArrayParams(paramName)
	result := make([]int, 0, len(strValues))

	for _, str := range strValues {
		val, err := strconv.Atoi(str)
		if err != nil {
			return nil, fmt.Errorf("参数 %s 包含非整数值: %s", paramName, str)
		}
		result = append(result, val)
	}

	return result, nil
}

// FailParams 参数解析失败
func (c *Context) FailParams(details string) {
	c.FailCode(ParamsErrCode, "请求参数解析失败: "+details)
}

// BindJSON 绑定 JSON 请求体，同时返回是否需要终止
func (c *Context) BindJSON(obj any) (needReturn bool) {
	if err := c.Context.ShouldBindJSON(obj); err != nil {
		c.FailParams(err.Error())
		return true
	}
	return false
}

// BindQuery 绑定查询参数，同时返回是否需要终止
func (c *Context) BindQuery(obj any) (needReturn bool) {
	if err := c.Context.ShouldBindQuery(obj); err != nil {
		c.FailParams(err.Error())
		return true
	}
	return false
}

// BindRoleId 绑定角色ID，返回是否需要终止
func (c *Context) BindRoleId(roleId *int) (needReturn bool) {
	roleIdStr := c.Param("roleId")
	var err error
	*roleId, err = strconv.Atoi(roleIdStr)
	if err != nil {
		c.FailParams("roleId必须为整数")
		return true
	}
	return false
}

// BindUserId 绑定用户ID，返回是否需要终止
func (c *Context) BindUserId(userId *int) (needReturn bool) {
	userIdStr := c.Param("userId")
	var err error
	*userId, err = strconv.Atoi(userIdStr)
	if err != nil {
		c.FailParams("userId必须为整数")
		return true
	}
	return false
}

// BindQueryPage 绑定分页查询参数，返回是否需要终止
func (c *Context) BindQueryPage(data IPage) (needReturn bool) {
	if err := c.Context.BindQuery(data); err != nil {
		c.FailParams(err.Error())
		return true
	}
	data.Constraint() //约束分页
	c.setPage(data.Copy())
	return false
}

// setPage 将page存入上下文
func (c *Context) setPage(page *PageDTO) {
	c.Set(PageTag, page)
}

// getPage 将page从上下文中获取
func (c *Context) getPage() *PageDTO {
	page, ok := c.Get(PageTag)
	if !ok {
		return nil
	}
	return page.(*PageDTO)
}

// Param 获取路径参数
func (c *Context) Param(key string) string {
	return c.Context.Param(key)
}

// Query 获取查询参数
func (c *Context) Query(key string) string {
	return c.Context.Query(key)
}

// DefaultQuery 获取查询参数，如果不存在则返回默认值
func (c *Context) DefaultQuery(key, defaultValue string) string {
	return c.Context.DefaultQuery(key, defaultValue)
}

// PostForm 获取表单参数
func (c *Context) PostForm(key string) string {
	return c.Context.PostForm(key)
}

// DefaultPostForm 获取表单参数，如果不存在则返回默认值
func (c *Context) DefaultPostForm(key, defaultValue string) string {
	return c.Context.DefaultPostForm(key, defaultValue)
}

// GetHeader 获取请求头
func (c *Context) GetHeader(key string) string {
	return c.Context.GetHeader(key)
}

// SetHeader 设置响应头
func (c *Context) SetHeader(key, value string) {
	c.Context.Header(key, value)
}

// SetUserId 设置用户ID
func (c *Context) SetUserId(userId int) {
	c.Set(UserIdTag, userId)
}

// GetUserId 获取用户ID
func (c *Context) GetUserId() (int, bool) {
	userIdAny, ok := c.Get(UserIdTag)
	if !ok || userIdAny == nil {
		c.Fail("获取不到操作者信息")
		return 0, false
	}
	userId, ok := userIdAny.(int)
	if !ok {
		c.Fail("获取不到操作者信息.")
		return 0, false
	}
	return userId, true
}

// GetOperateId 获取操作id,也就是用户id,跟GetUserId的区别是,调用方法不一样
func (c *Context) GetOperateId(operateId *int) (exists bool) {
	userId, exists := c.GetUserId()
	if exists {
		*operateId = userId
	}
	return exists
}
