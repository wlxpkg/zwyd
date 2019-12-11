/*
 * @Author: qiuling
 * @Date: 2019-06-20 16:58:11
 * @Last Modified by: qiuling
 * @Last Modified time: 2019-12-10 18:57:36
 */
package middleware

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"

	. "github.com/wlxpkg/base"
	"github.com/wlxpkg/base/model"
	. "github.com/wlxpkg/zwyd"

	"github.com/gin-gonic/gin"
)

func Member() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 后置数据准备
		bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()                                        //  must close
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // 重新赋值

		c.Set("bodyCopy", string(bodyBytes))

		userInfo, err := getUser(c)
		userID, _ := String2Int64(userInfo["user_id"])

		// R(userInfo, "userInfo")
		if err != nil {
			err = errors.New("ERR_INVALID_TOKEN")
			Abort(c, err)
			return
		}

		guestAllow := checkGuest(c, userID)
		if !guestAllow {
			err = errors.New("ERR_UNLOGIN")
			Abort(c, err)
			return
		}

		if userID == 0 || userID == 403 {
			c.Set("middleware", &Middleware{})
			c.Next()
		}

		permission := getPermission(c, userID)
		// R(permission, "permission")

		middleware := middlewareData(userInfo, permission)

		// 设置 example 变量
		c.Set("middleware", middleware)
		c.Next()
	}
}

func checkGuest(c *gin.Context, userID int64) bool {
	if userID == 0 || userID == 403 {
		path := c.Request.URL.Path
		method := c.Request.Method

		route := getRoute(path, method, 0)
		if route == "" {
			// 没匹配到白名单, 则无权限, 需要登录
			return false
		}
	}

	return true
}

// getPermission 检测会员是否有权限
func getPermission(c *gin.Context, userID int64) bool {
	if userID == 0 || userID == 403 {
		return false
	}
	path := c.Request.URL.Path
	method := c.Request.Method

	route := getRoute(path, method, 1)
	if route == "" {
		// 无需鉴权则直接返回 false
		return false
	}

	clientID := c.GetHeader("client-id")

	permission := checkRole(userID, clientID)
	return permission
}

// getRoute 获取本次请求匹配的路由
// rtype 路由类型, 0:guest, 1:member
func getRoute(path, method string, rtype int) (route string) {

	var allRoute []string
	switch rtype {
	case 0:
		allRoute = model.GuestRoute()
	case 1:
		allRoute = model.MemberRoute()
	}

	route = ""

	for _, routes := range allRoute {
		routeSli := strings.Split(routes, "@")
		if len(routeSli) < 2 {
			continue
		}

		if KeyMatch(path, routeSli[1]) && method == routeSli[0] {
			route = routeSli[1]
			break
		}
	}

	return
}

func checkRole(userID int64, clientID string) bool {
	time := model.MemberTime(userID, clientID)
	if time == "" {
		return false
	}

	return true
}
