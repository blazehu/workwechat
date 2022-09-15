# workwechat
企业微信服务端API，[相关文档](https://developer.work.weixin.qq.com/document/path/90664)

## Test
```shell
go test -v --corpid xxx --corpsecret xxx --agentid xxx
```

test case console log:

```
=== RUN   TestAppParse
    work_test.go:27: accessToken is:  xxxxxxxxxx
    work_test.go:42: Department Id: 2, Name: 运维支持, ParentId: 1, Order: 100000000
    work_test.go:48: User Id: zhangsan, Name: 张三
    work_test.go:48: User Id: lisi, Name: 李四
    work_test.go:48: User Id: wangwu, Name: 王五
    work_test.go:56: {{0 ok} {[map[userid:zhaoliu]]} {[2]}}
    work_test.go:63: User Id: blazehu, Name: blazehu
    work_test.go:67: {0 ok}
--- PASS: TestAppParse (2.98s)
PASS
ok  	workwechat	3.676s
```

### Documentation for API Endpoints

| Func           | HTTP request          | Description     |
|----------------|-----------------------|-----------------|
| AccessToken    | GET /gettoken         | 获取 access token |
| GetAgent       | GET /agent/get        | 获取指定的应用详情       |
| GetUserId      | GET /user/getuserinfo | 获取访问用户身份        |
| GetUser        | GET /user/get         | 读取成员            |
| ListUser       | GET /user/simplelist  | 获取部门成员          |
| ListDepartment | GET /department/list  | 获取部门列表          |
| Message        | GET /message/send     | 发送应用消息          |
