package handlers_test

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/handlers"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var cookies []*http.Cookie
var client *http.Client

func TestMain(m *testing.M) {
	//清理数据库
	cleanUpDatabase()
	//创建Http测试用的客户端实例
	client = &http.Client{}
	//登录root用户
	//获得测试使用的登录态cookie
	resp, _ := sessionCreate(client)
	cookies = resp.Cookies()

	//执行测试用例
	exitVal := m.Run()

	//清理数据库
	//cleanUpDatabase()

	os.Exit(exitVal)
}

func TestRoutes(t *testing.T) {
	actionsCheckResult, err := checkActions([]string{
		"/pools/{id:.*}/inspect",
		"/pools/register", //&MyHandler{h: postPoolsRegister, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/pools/ps",       //&MyHandler{h: getPoolsJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/pools/json",     //&MyHandler{h: getPoolsJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},

		"/users/{name:.*}/login",   //&MyHandler{h: getUserLogin},
		"/users/current",           //&MyHandler{h: getUserCurrent },
		"/users/create",            //&MyHandler{h: postUsersCreate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/inspect",   /// &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/detail",    // &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/ps",                //&MyHandler{h: getUsersJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/list",              //&MyHandler{h: getUsersJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/resetpass", //&MyHandler{h: postUserResetPass, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/remove",    //&MyHandler{h: postUserRemove, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/update",    // &MyHandler{h: postUserUpdate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/users/{id:.*}/join",      //&MyHandler{h: postUserJoin, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/users/{id:.*}/quit",      // &MyHandler{h: postUserQuit, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},

		"/teams/create",          // &MyHandler{h: postTeamsCreate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/inspect", // &MyHandler{h: getTeamJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/ps",              //&MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/list",            //             // &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/update",  //   &MyHandler{h: postTeamUpdate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/appoint",
		"/teams/{id:.*}/revoke",
		"/teams/{id:.*}/remove",
	})
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Check Routes Success.")
	}

	for action, b := range actionsCheckResult.Action2Result {
		t.Logf("%s:%#v\n", action, b)
	}
}

//"/actions/check" : &MyHandler{h: postActionsCheck } ,
func checkActions(actions []string) (*handlers.ActionCheckResponse, error) {

	url := fmt.Sprintf("http://localhost:8080/api/actions/check")

	r := handlers.ActionsCheckRequest{
		Actions: actions,
	}

	buf, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(buf)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("checkActions read body err:%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	u := handlers.ActionCheckResponse{
		Action2Result: map[string]bool{},
	}
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}

	return &u, nil

}

//清理数据库
//避免遗留的测试数据对测试结果造成干扰
func cleanUpDatabase() {
	cmds := [...][2]string{
		{"mongo", "zanecloud --eval \"db.user.remove({'name':'sadan'})\""},
		{"mongo", "zanecloud --eval \"db.team.remove({'name':'team1'})\""},
		{"mongo", "zanecloud --eval \"db.pool.remove({'name':'pool1'})\""},
		{"mongo", "zanecloud --eval \"db.env_tree_meta.remove({})\""},
		{"mongo", "zanecloud --eval \"db.env_tree_node_dir.remove({})\""},
	}

	for _, arr := range cmds {
		_, err := exec.Command("sh", "-c", fmt.Sprintf("%s %s", arr[0], arr[1])).Output()
		if err != nil {
			log.Fatal(err)
		}
	}
}
