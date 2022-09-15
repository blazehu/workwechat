package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	httpTimeout = 10 * time.Minute
	HOST        = "https://qyapi.weixin.qq.com/cgi-bin"
)

// Client
// CorpID:  Enterprise WeChat ID
// CorpSecret:  Enterprise WeChat application Secret
// AgentID:  Enterprise WeChat App ID
type Client struct {
	CorpID     string
	CorpSecret string
	AgentID    string
}

type Resp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type AccessTokenResp struct {
	Resp
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type User struct {
	UserId string `json:"userid"`
	Name   string `json:"name"`
}

func (u User) String() string {
	return fmt.Sprintf("User Id: %s, Name: %s", u.UserId, u.Name)
}

type UserInfoResp struct {
	Resp
	UserId   string `json:"UserId"`
	DeviceId string `json:"DeviceId"`
}

type Department struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentId int    `json:"parentid"`
	Order    int    `json:"order"`
}

func (d Department) String() string {
	return fmt.Sprintf("Department Id: %d, Name: %s, ParentId: %d, Order: %d", d.ID, d.Name, d.ParentId, d.Order)
}

type DepartmentResp struct {
	Resp
	DepartmentList []Department `json:"department"`
}

type ListUserResp struct {
	Resp
	Users []User `json:"userlist"`
}

type AllowDepartment struct {
	Partyid []int `json:"partyid"`
}

type AllowUsers struct {
	User []map[string]string `json:"user"`
}

type AgentResp struct {
	Resp
	AllowUsers      AllowUsers      `json:"allow_userinfos"`
	AllowDepartment AllowDepartment `json:"allow_partys"`
}

type UserResp struct {
	Resp
	User
}

// GetAPI assemble api
func (c Client) GetAPI(path string) string {
	return strings.TrimSpace(HOST) + "/" + strings.TrimSpace(path)
}

// Request base http client handle
func (c Client) request(path, method string, data map[string]interface{}, params map[string]string) ([]byte, error) {
	// NOTE: parse data
	var (
		body io.Reader
	)
	if data != nil {
		bytesData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bytesData)
	} else {
		body = nil
	}

	// NOTE: new http client
	client := &http.Client{
		Timeout: time.Duration(httpTimeout),
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	url := c.GetAPI(path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// NOTE: set header
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Connection", "Close")
	req.Header.Set("Accept", "*/*")

	// NOTE: set params
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// NOTE: send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// NOTE: parse resp data
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		errMsg := "StatusCode is " + strconv.Itoa(resp.StatusCode) + ",Error: " + string(respData)
		return nil, errors.New(errMsg)
	}
	return respData, nil
}

// ParseRespData parse response data to map[string]interface{}
func (c Client) ParseRespData(respData []byte) (map[string]interface{}, bool) {
	var isOk bool
	data := make(map[string]interface{})
	if err := json.Unmarshal(respData, &data); err != nil {
		isOk = false
		return nil, isOk
	}
	// NOTE: WeChat Api errCode
	if value, ok := data["errcode"]; ok {
		switch reflect.ValueOf(value).Kind() {
		case reflect.Float64:
			isOk = value == 0.0
		case reflect.Int:
			isOk = value == 0
		case reflect.String:
			isOk = value == "0"
		}
	}
	return data, isOk
}

// AccessToken get access token
func (c Client) AccessToken() (string, error) {
	path := "gettoken"
	params := map[string]string{
		"corpid":     c.CorpID,
		"corpsecret": c.CorpSecret,
	}
	resp := AccessTokenResp{}
	respData, err := c.request(path, "GET", nil, params)
	if err != nil {
		return resp.AccessToken, err
	}

	if err := json.Unmarshal(respData, &resp); err != nil {
		return resp.AccessToken, errors.New(fmt.Sprintf("get access_token failed"))
	}
	return resp.AccessToken, nil
}

// GetUserId get user info by WeChat code
func (c Client) GetUserId(code string) (string, error) {
	path := "user/getuserinfo"

	// NOTE: get accessToken
	accessToken, err := c.AccessToken()
	if err != nil {
		return "", err
	}
	params := map[string]string{
		"access_token": accessToken,
		"code":         code,
	}
	respData, err := c.request(path, "GET", nil, params)
	if err != nil {
		return "", err
	}

	resp := UserInfoResp{}
	if err := json.Unmarshal(respData, &resp); err != nil {
		return resp.UserId, errors.New(fmt.Sprintf("get userId failed"))
	}
	return resp.UserId, nil
}

// ListDepartment within the scope of app permissions
func (c Client) ListDepartment() (DepartmentResp, error) {
	path := "department/list"

	resp := DepartmentResp{}
	// NOTE: get accessToken
	accessToken, err := c.AccessToken()
	if err != nil {
		return resp, err
	}
	params := map[string]string{
		"access_token": accessToken,
	}
	respData, err := c.request(path, "GET", nil, params)
	if err != nil {
		return resp, err
	}

	if err := json.Unmarshal(respData, &resp); err != nil {
		return resp, errors.New(fmt.Sprintf("get department failed"))
	}
	return resp, nil
}

// ListUser get a list of users under a department, can be retrieved recursively
func (c Client) ListUser(departmentID string, fetchChild bool) (ListUserResp, error) {
	path := "user/simplelist"

	resp := ListUserResp{}
	// NOTE: get accessToken
	accessToken, err := c.AccessToken()
	if err != nil {
		return resp, err
	}
	var fetchChildParams string
	if fetchChild {
		fetchChildParams = "1"
	} else {
		fetchChildParams = "0"
	}
	params := map[string]string{
		"access_token":  accessToken,
		"department_id": departmentID,
		"fetch_child":   fetchChildParams,
	}
	respData, err := c.request(path, "GET", nil, params)
	if err != nil {
		return resp, err
	}

	if err := json.Unmarshal(respData, &resp); err != nil {
		return resp, errors.New(fmt.Sprintf("list users failed"))
	}
	return resp, nil
}

// GetAgent get a agent
func (c Client) GetAgent() (AgentResp, error) {
	path := "agent/get"

	resp := AgentResp{}
	// NOTE: get accessToken
	accessToken, err := c.AccessToken()
	if err != nil {
		return resp, err
	}
	params := map[string]string{
		"access_token": accessToken,
		"agentid":      c.AgentID,
	}
	respData, err := c.request(path, "GET", nil, params)
	if err != nil {
		return resp, err
	}

	if err := json.Unmarshal(respData, &resp); err != nil {
		return resp, errors.New(fmt.Sprintf("get agent failed"))
	}
	return resp, nil
}

// GetUser get a user
func (c Client) GetUser(userid string) (UserResp, error) {
	path := "user/get"

	resp := UserResp{}
	// NOTE: get accessToken
	accessToken, err := c.AccessToken()
	if err != nil {
		return resp, err
	}
	params := map[string]string{
		"access_token": accessToken,
		"userid":       userid,
	}
	respData, err := c.request(path, "GET", nil, params)
	if err != nil {
		return resp, err
	}

	if err := json.Unmarshal(respData, &resp); err != nil {
		return resp, errors.New(fmt.Sprintf("get users failed"))
	}
	return resp, nil
}

// Message send application message notifications to enterprise WeChat users
func (c Client) Message(toUser, content string) (Resp, error) {
	path := "message/send"

	resp := Resp{}
	// NOTE: get accessToken
	accessToken, err := c.AccessToken()
	if err != nil {
		return resp, err
	}
	params := map[string]string{
		"access_token": accessToken,
	}
	data := map[string]interface{}{
		"touser":  toUser,
		"msgtype": "text",
		"agentid": c.AgentID,
		"text": map[string]string{
			"content": content,
		},
	}
	respData, err := c.request(path, "POST", data, params)
	if err != nil {
		return resp, err
	}

	if err := json.Unmarshal(respData, &resp); err != nil {
		return resp, errors.New(fmt.Sprintf("message failed"))
	}
	return resp, nil
}
