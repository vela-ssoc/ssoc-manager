package ywcmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Duty struct {
	Block      bool   `json:"block"`
	EmployeeID string `json:"employee_id"`
	Nickname   string `json:"nickname"`
	UID        int    `json:"uid"`
	Username   string `json:"username"`
	UType      string `json:"utype"`
}

type Duties []Duty

type DutyMain struct {
	Main Duties `json:"main,omitzero"`
}

type ResponseBody[S ~[]E, E any] struct {
	Code     int    `json:"code"`
	Page     int    `json:"page"`
	Total    int    `json:"total"`
	NumFound int    `json:"numFound"`
	Message  string `json:"message"`
	Result   S      `json:"result"`
}

func (r ResponseBody[S, E]) checkError() *CMDBError {
	if r.Code == 20000 {
		return nil
	}

	return &CMDBError{Code: r.Code, Message: r.Message}
}

type CMDBError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *CMDBError) Error() string {
	code := strconv.Itoa(e.Code)
	return "CMDB 服务器错误码: " + code + ", 错误信息: " + e.Message
}

type HTTPError struct {
	Code    int
	Body    []byte
	Request *http.Request
}

func (e *HTTPError) Error() string {
	var method, strURL string
	if req := e.Request; req != nil {
		method = req.Method
		strURL = req.URL.Redacted()
	}
	return fmt.Sprintf("HTTP 请求失败 %s %s: %d %s", method, strURL, e.Code, e.Body)
}

func (e *HTTPError) JSON(v any) error {
	return json.Unmarshal(e.Body, v)
}
