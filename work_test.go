package wechat

import (
	"flag"
	"strconv"
	"testing"
)

var corpID = flag.String("corpid", "", "corpid")
var corpSecret = flag.String("corpsecret", "", "corpsecret")
var agentid = flag.String("agentid", "", "agentid")
var code = flag.String("code", "", "corpsecret")

func TestAppParse(t *testing.T) {
	flag.Parse()

	client := Client{
		CorpID:     *corpID,
		CorpSecret: *corpSecret,
		AgentID:    *agentid,
	}

	accessToken, err := client.AccessToken()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("accessToken is: ", accessToken)

	if *code != "" {
		userInfo, err := client.GetUserId(*code)
		if err != nil || userInfo == "" {
			t.Fatal(err.Error())
		}
		t.Log("userInfo is: ", userInfo)
	}

	departments, err := client.ListDepartment()
	if err != nil {
		t.Fatal(err.Error())
	}
	for _, department := range departments.DepartmentList {
		t.Log(department)
		users, err := client.ListUser(strconv.Itoa(department.ID), true)
		if err != nil {
			t.Fatal(err.Error())
		}
		for _, user := range users.Users {
			t.Log(user)
		}
	}

	agent, err := client.GetAgent()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(agent)

	for _, item := range agent.AllowUsers.User {
		user, err := client.GetUser(item["userid"])
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Log(user)
	}

	resp, err := client.Message("blazehu|blazehu", "测试")
	t.Log(resp)
	if err != nil {
		t.Fatal(err.Error())
	}
}
