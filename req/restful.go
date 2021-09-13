/*
 * @Author: qiuling
 * @Date: 2019-06-25 20:44:57
 * @Last Modified by: qiuling
 * @Last Modified time: 2019-12-05 11:10:17
 */
package req

import (
	"encoding/json"
	"errors"

	. "github.com/wlxpkg/base"
	. "github.com/wlxpkg/base/config"
	"github.com/wlxpkg/base/log"
	"github.com/wlxpkg/base/req"
)

type Restful struct {
	client  *req.HttpClient
	exp     bool
	service map[string]string
}

func NewRestful(name string) *Restful {
	r := new(Restful)
	r.exp = true
	r.client = req.NewClient("")
	r.setService()
	r.GetService(name)
	return r
}

// setService 初始化时候设置服务map
func (r *Restful) setService() {
	config := Config.Server

	serverURL := make(map[string]string)

	serverURL["tools"] = "http://" + config.Tools
	serverURL["user"] = "http://" + config.User
	serverURL["course"] = "http://" + config.Course
	serverURL["discovery"] = "http://" + config.Discovery
	serverURL["common"] = "http://" + config.Common
	serverURL["grant"] = "http://" + config.Grant
	serverURL["shop"] = "http://" + config.Shop
	serverURL["message"] = "http://" + config.Message
	serverURL["game"] = "http://" + config.Game

	r.service = serverURL
}

// GetService 获取一个服务地址设置给 http 客户端
func (r *Restful) GetService(name string) *Restful {
	baseURL, exists := r.service[name]

	if !exists {
		log.Err("Restful.GetService 服务不存在, name:" + name)
	}
	r.client.SetBaseURL(baseURL)
	return r
}

// SetJwt 设置 jwt
func (r *Restful) SetJwt(jwt string) *Restful {
	r.client.AddHeader("Authorization", "Bearer "+jwt)
	return r
}

// SetExp 设置 exp, 为 true 则不返回原始数据直接异常
// 默认 true
func (r *Restful) SetExp(exp bool) *Restful {
	r.exp = exp
	return r
}

func (r *Restful) SetData(data map[string]string) *Restful {
	r.client.SetData(data)
	return r
}

func (r *Restful) SetHeader(data map[string]string) *Restful {
	r.client.AddHeaders(data)
	return r
}

// SetSecret 内部服务请求设置 ServiceSecret
func (r *Restful) SetSecret() *Restful {
	r.client.AddHeader("ServiceSecret", Config.Service.Secret)
	return r
}

// Req 发送请求
func (r *Restful) Req(method string, route string) (data interface{}, err error) {
	r = r.SetSecret() // 全部调用 SetSecret
	resp, err := r.client.Request(method, route)
	// resp := "{\"code\":1,\"message\":\"\",\"data\":{\"user_id\":\"1134660407147180032\",\"avatar\":\"http:\\/\\/thirdwx.qlogo.cn\\/mmopen\\/vi_32\\/Q3auHgzwzM48ybqIC8FzI2xAbkVEY4gsyL8XSSicX1R42woyg7sUEceXJesG1QL9BOH33B26DQsZZGKMsx6r0xA\\/132\",\"nickname\":\"阿Q\",\"jwt\":\"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiIsImp0aSI6Im12T3pAMVU5Qk8ifQ.eyJqdGkiOiJtdk96QDFVOUJPIiwiaWF0IjoxNTYxNzEzNDkyLCJleHAiOjE1NjE3NTY2OTIsImFydGlmYWN0IjoiUHowaVBORm9VNUlhSExtcFpRNHVTNU9PdHVwS2dxY1giLCJ0b2tlbiI6ImI3YTVhMWI5ODMzODBhY2U1ZmIxZjJmNjkwNzk1N2I0In0.eT6O-Y0etAuv1urK5lgsFWWHuM_x9bVr9Wief9uNDDw\"}}"

	if err != nil {
		reqStr, _ := json.Marshal(r.client.Data)
		log.Warn("微服务请求失败! service: " + r.client.BaseURL + ", reqData: " + Byte2String(reqStr) + ", method: " + method + ", route: " + route)
		return
	}

	data, err = r.serviceData(resp, method+"@"+route)
	return
}

// serviceData 解析数据
func (r *Restful) serviceData(resp string, req string) (resData interface{}, err error) {

	/* result, ok := gjson.Parse(resp).Value().(map[string]interface{})
	 if !ok {
		 reqStr, _ := json.Marshal(r.client.data)
		 log.Warn("微服务数据解析失败! service: " + r.client.baseUrl + " reqData: " + Byte2String(reqStr) + "resp: " + resp)
		 err = Excp("ERR_DATA_DECODE")
		 return
	 } */
	result, err := JsonDecode(resp)
	if err != nil {
		reqStr, _ := json.Marshal(r.client.Data)
		log.Warn("微服务数据解析失败! service: " + r.client.BaseURL + ", reqData: " + Byte2String(reqStr) + ", resp: " + resp + ", 请求的方法:" + req)
		return
	}

	// R(result, "result")
	code := result["code"].(float64)
	message := result["message"].(string)
	data := result["data"]

	if r.exp {
		if code != 1 {
			err = errors.New(message)
			return
		}
		resData = data
	} else {
		resData = resp
	}

	// result := gjson.GetMany(resp, "code", "message", "data")
	// code := gjson.Get(resp, "code")
	// message := gjson.Get(resp, "message")
	// data := gjson.Get(resp, "data")
	/* if r.exp {
		 if code.Int() != 1 {
			 err = errors.New(message.String())
			 return
		 }
		 resData = data.String()
	 } else {
		 resData = resp
	 } */

	return
}
