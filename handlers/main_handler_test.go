package handlers_test

import (
	"testing"
	"net/http"
	"io/ioutil"
	"fmt"
	log "github.com/Sirupsen/logrus"
	dockerclient "github.com/docker/docker/client"
	"github.com/zanecloud/apiserver/types"
	"encoding/json"
	"strings"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/handlers"
	"os"
	"os/exec"
	"context"
)

var cookies []*http.Cookie
var client *http.Client

func TestMain(m *testing.M)  {
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
	cleanUpDatabase()

	os.Exit(exitVal)
}

func TestSession(t *testing.T)  {
	err := sessionLogout()
	if err != nil {
		t.Error(err)
	}else{
		t.Log("Logout success")
	}

	//登录root用户
	//获得测试使用的登录态cookie
	resp, err := sessionCreate(client)
	if err != nil {
		t.Error(err)
	}else{
		cookies = resp.Cookies()
		t.Log(cookies)
	}
}

func TestTeam(t *testing.T)  {
	var teamId string
	t.Run("TEAM=1", func(t *testing.T) {
		var err error
		teamId , err = createTeam(&types.Team{
			Name : "team1",
			Description: "dev team1",
		})
		if err != nil {
			t.Error(err)
		}else {
			t.Log(teamId)
		}
	})
	t.Run("TEAM=2", func(t *testing.T) {
		team , err:= inspectTeam(teamId)
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(team)
		}
	})
	t.Run("TEAM=3", func(t *testing.T) {
		teams , err := listTeam()
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(teams)
		}

		if len(teams) != 1 {
			t.Error("Team count is not correct!")
		}
	})
}


func TestUser(t *testing.T){
	var userId string
	var teamId string
	t.Run("USER=1", func(t *testing.T) {
		currentUser , err := currentUser()
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(currentUser)
		}
		userId = currentUser.Id.String()
	})

	t.Run("USER=2", func(t *testing.T) {
		uid , err := createUser(&types.User{
			Name : "sadan",
			Pass : "1234",
			RoleSet: types.ROLESET_NORMAL ,
			Email : "zhengtao.wuzt@gmail.com",
			Tel        : "18167189863",
		})
		if err!= nil {
			t.Error(err)
		}else{
			t.Log(uid)
		}

		userId = uid
	})

	t.Run("USER=3", func(t *testing.T) {
		user,err:= inspectUser(userId)
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(user)
		}
		if user.Id.Hex() != userId {
			t.Error("Inspect User error! user id not correct!")
		}
	})

	t.Run("USER=4", func(t *testing.T) {
		users , err := listUser()
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(users)
		}

		if len(users) != 2 {
			t.Error("User count not correct")
		}
	})

	//加入team
	t.Run("USER=5", func(t *testing.T) {
		teams , err := listTeam()
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(teams)
		}

		if len(teams) != 1 {
			t.Error("Team count is not correct!")
		}

		tx := teams[0]
		teamId = tx.Id.Hex()
		if err := joinTeam(userId, teamId) ; err != nil {
			t.Error(err)
		}
		user, err := inspectUser(userId)
		if err !=nil {
			t.Error(err)
		}else {
			t.Log(user)
		}
		//TODO
		//检查某个用户是否已经退出某个TEAM

		team ,err := inspectTeam(teamId)
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(team)
		}
	})

	t.Run("USER=6", func(t *testing.T) {
		email :="76577126@qq.com"
		if err := updateUser(userId ,email); err != nil {
			t.Error(err)
		}
		user, err := inspectUser(userId)
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(user)
		}

		if user.Email != email {
			t.Error("User update error for email!")
		}
	})

	t.Run("USER=6", func(t *testing.T) {

		if err := quitTeam(userId , teamId ) ; err != nil {
			t.Error(err)
		}
		user, err := inspectUser(userId)
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(user)
		}
		//TOOD
		//检查该用户是否已经退出某个team

		team ,err := inspectTeam(teamId)
		if err !=nil {
			t.Error(err)
		}else{
			t.Log(team)
		}
	})

	t.Run("USER=7", func(t *testing.T) {
		if err := appointTeam(teamId,userId) ; err !=nil{
			t.Error(err)
		}

		team,err := inspectTeam(teamId)
		if err !=nil {
			t.Error(err)
		}else {
			t.Log(team)
		}
		if err := revokeTeam(teamId,userId) ; err !=nil{
			t.Error(err)
		}

		team, err = inspectTeam(teamId)
		if err !=nil {
			t.Error(err)
		}else {
			t.Log(team)
		}
	})
}

func TestRoutes(t *testing.T)  {
	actionsCheckResult , err := checkActions([]string{
		"/pools/{id:.*}/inspect",
		"/pools/register",        //&MyHandler{h: postPoolsRegister, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/pools/ps",              //&MyHandler{h: getPoolsJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/pools/json",            //&MyHandler{h: getPoolsJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},

		"/users/{name:.*}/login",   //&MyHandler{h: getUserLogin},
		"/users/current",           //&MyHandler{h: getUserCurrent },
		"/users/create",           //&MyHandler{h: postUsersCreate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/inspect",  /// &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/detail",   // &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/ps",                //&MyHandler{h: getUsersJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/list",              //&MyHandler{h: getUsersJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/resetpass", //&MyHandler{h: postUserResetPass, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/remove",    //&MyHandler{h: postUserRemove, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/update",   // &MyHandler{h: postUserUpdate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/users/{id:.*}/join",      //&MyHandler{h: postUserJoin, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/users/{id:.*}/quit",     // &MyHandler{h: postUserQuit, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},

		"/teams/create",           // &MyHandler{h: postTeamsCreate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/inspect",  // &MyHandler{h: getTeamJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/ps",                //&MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/list",//             // &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/update", //   &MyHandler{h: postTeamUpdate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/appoint",
		"/teams/{id:.*}/revoke",
		"/teams/{id:.*}/remove",
	})
	if err !=nil {
		t.Error(err)
	}else{
		t.Log("Check Routes Success.")
	}

	for action, b := range actionsCheckResult.Action2Result{
		t.Logf("%s:%#v\n" , action , b)
	}
}

func TestPool(t *testing.T)  {
	result , err := registerPool("pool1", &handlers.PoolsRegisterRequest{
		Name: "pool1",
		Driver: "swarm",
		DriverOpts: types.DriverOpts{
			Version:"v1.0",
			EndPoint:"47.92.49.245:2375",
			APIVersion:"v1.23",
		},
	})

	if err!=nil {
		t.Error(err)
	}else{
		t.Log(result)
	}

	r , _ := result.(handlers.PoolsRegisterResponse)
	dockerclient, err:=dockerclient.NewClient(r.Proxy,"v1.23",nil,map[string]string{})
	if err!=nil {
		t.Error(err)
	}

	info , err :=dockerclient.Info(context.Background())
	if err!=nil {
		t.Error(err)
	}else{
		t.Log(info)
	}
}

//root用户登录
//获取登陆后的cookie
func sessionCreate(client *http.Client) (resp *http.Response, err error)  {
	req, err := http.NewRequest("POST", "http://localhost:8080/api/users/root/login?Pass=hell05a" ,nil )
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	defer resp.Body.Close()
	if err !=nil {
		log.Errorf("login post err:%s", err.Error())
		return
	}


	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s",err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		return resp, errors.Errorf("login statuscode:%d err:%s" , resp.StatusCode, string(body))
	}

	//fmt.Println(string(body))
	rootUser:=&types.User{}
	json.Unmarshal(body,rootUser)
	//fmt.Printf("\nlogin success , the root  user is %#v....",rootUser)

	//for _ , cookie := range resp.Cookies(){
	//	fmt.Println("cookie:", cookie)
	//}

	return
}


//"/actions/check" : &MyHandler{h: postActionsCheck } ,
func checkActions(actions []string) (*handlers.ActionCheckResponse ,error) {

	url := fmt.Sprintf("http://localhost:8080/api/actions/check")

	r:=handlers.ActionsCheckRequest{
		Actions:actions,
	}

	buf , err := json.Marshal(r)
	if err!=nil {
		return nil,err
	}

	req , err := http.NewRequest(http.MethodPost , url,  strings.NewReader(string(buf)))
	if err != nil {
		return  nil,err
	}
	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}



	resp, err := client.Do(req)
	if err!=nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("checkActions read body err:%s",err.Error())
		return nil,err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	u:=handlers.ActionCheckResponse{
		Action2Result: map[string]bool{},
	}
	if err := json.Unmarshal(body,&u); err !=nil {
		return nil, err
	}


	return &u , nil


}

func joinTeam(userId string , teamId string)( error){

	url := fmt.Sprintf("http://localhost:8080/api/users/%s/join?TeamId=%s",userId,teamId)

	req , err := http.NewRequest(http.MethodPost , url,nil)
	if err != nil {
		return  err
	}
	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err!=nil {
		return  err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s",err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return  errors.New(string(body))
	}


	return  nil
}




func quitTeam(userId string , teamId string)( error){

	url := fmt.Sprintf("http://localhost:8080/api/users/%s/quit?TeamId=%s",userId,teamId)

	req , err := http.NewRequest(http.MethodPost , url,nil)
	if err != nil {
		return  err
	}
	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err!=nil {
		return  err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("quitTeam read body err:%s",err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return  errors.New(string(body))
	}


	return  nil
}

func createTeam(team *types.Team) (string ,  error){
	buf , err := json.Marshal(team)
	if err!=nil {
		return "" , err
	}

	req , err := http.NewRequest(http.MethodPost , "http://localhost:8080/api/teams/create",strings.NewReader(string(buf)))
	if err != nil {
		return "" , err
	}

	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return "" ,err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s",err.Error())
		return "" ,err
	}

	if resp.StatusCode != http.StatusOK {
		return "" , errors.New(string(body))
	}

	result:= handlers.TeamsCreateResponse{}
	if err := json.Unmarshal(body,&result) ; err !=nil {
		return "" , err
	}


	return result.Id, nil

}


func inspectTeam(teamId string) (*types.Team ,  error){

	url := fmt.Sprintf("http://localhost:8080/api/teams/%s/inspect",teamId)

	req , err := http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		return nil , err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return nil ,err
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return nil ,err
		}

		return nil , errors.New(string(body))
	}

	result:= &types.Team{}
	if err := json.NewDecoder(resp.Body).Decode(result) ; err !=nil {
		return nil , err
	}

	return result, nil

}

//"/users/current":           &MyHandler{h: getUserCurrent ,opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
func TestCurrentUser(t *testing.T){

	url := fmt.Sprintf("http://localhost:8080/api/users/current")

	req , err := http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		t.Error(err)
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			t.Error(err)
		}
		t.Error(string(body))
	}

	result:= &types.User{}
	if err := json.NewDecoder(resp.Body).Decode(result) ; err !=nil {
		t.Error(err)
	}

	if err !=nil {
		t.Errorf("inspect the current user err:%s",err.Error())
	}else {
		t.Logf("the current user is  %#v \n", result)
	}

}

/*

	以下是测试用例的辅助方法

*/

//"/teams/{id:.*}/remove":  checkUserPermission(postTeamRemove,types.ROLESET_SYSADMIN),
func revokeTeam(teamId string , userId string) ( error){
	url := fmt.Sprintf("http://localhost:8080/api/teams/%s/revoke?UserId=%s",teamId,userId)

	req , err := http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("appoint read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return err
		}


		return  errors.New(string(body))
	}

	return nil

}


//"/teams/{id:.*}/appoint?UserId=xxx": checkUserPermission(postTeamAppoint,types.ROLESET_SYSADMIN),
func appointTeam(teamId string , userId string) ( error){
	url := fmt.Sprintf("http://localhost:8080/api/teams/%s/appoint?UserId=%s",teamId,userId)

	req , err := http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("appoint read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return err
		}


		return  errors.New(string(body))
	}

	return nil

}

func listUser() ([]types.User ,  error){

	url := fmt.Sprintf("http://localhost:8080/api/users/ps")

	req , err := http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		return nil , err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return nil ,err
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return nil ,err
		}


		return nil , errors.New(string(body))
	}

	var result []types.User
	if err := json.NewDecoder(resp.Body).Decode(&result) ; err !=nil {
		//log.Errorf("decode the users buf : %s error:%s" , string(body) , err.Error() )
		return nil , err
	}

	return result, nil

}

func listTeam() ([]types.Team ,  error){

	url := fmt.Sprintf("http://localhost:8080/api/teams/ps")

	req , err := http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		return nil , err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return nil ,err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	//fmt.Printf("listTeam repos body is %s\n", string(body))
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return nil ,err
		}


		return nil , errors.New(string(body))
	}

	var result []types.Team
	if err := json.Unmarshal(body,&result) ; err !=nil {
		//log.Errorf("decode the users buf : %s error:%s" , string(body) , err.Error() )
		return nil , err
	}

	return result, nil

}

//"/users/{id:.*}/update":    &MyHandler{h: postUserUpdate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},

func  updateUser(userId string ,  email string) error {

	url := fmt.Sprintf("http://localhost:8080/api/users/%s/update" , userId)

	r:=handlers.UserUpdateRequest{}
	r.Email = email
	r.Roleset = types.ROLESET_APPADMIN

	buf , _ := json.Marshal(&r)

	req , err := http.NewRequest(http.MethodPost , url , strings.NewReader(string(buf)))
	if err != nil {
		return  err
	}


	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return err
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return err
		}

		return errors.New(string(body))
	}

	return nil
}


func registerPool(name string , request * handlers.PoolsRegisterRequest) (interface{} , error) {

	url := fmt.Sprintf("http://localhost:8080/api/pools/register")

	buf , _ := json.Marshal(request)

	req , _ := http.NewRequest(http.MethodPost , url , strings.NewReader(string(buf)))

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return nil , err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
		return err.Error(), err
	}

	if resp.StatusCode != http.StatusOK {
		return string(body), errors.New(string(body))
	}

	//fmt.Println(string(body))
	result := handlers.PoolsRegisterResponse{}
	json.Unmarshal(body,&result)

	return result , nil
}

func sessionLogout() (error){

	//先调用登出接口
	//是的cookie失效
	url := fmt.Sprintf("http://localhost:8080/api/session/logout")

	req , err := http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return err
		}

		return errors.New(string(body))
	}
	resp.Body.Close()
	//log.Infof("logout success.")

	//再调用current user接口
	//应该看到返回状态是401

	url = fmt.Sprintf("http://localhost:8080/api/users/current")

	req , err = http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err = client.Do(req)
	if err!=nil {
		return err
	}
	defer resp.Body.Close()

	//如果是未授权，则退出成功
	if resp.StatusCode != http.StatusUnauthorized {
		return errors.New(string("After logout, current user api is not 401"))
	}else {
		return nil
	}

	return nil

}

//"/users/current":           &MyHandler{h: getUserCurrent ,opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
func currentUser() (*types.User ,  error){

	url := fmt.Sprintf("http://localhost:8080/api/users/current")

	req , err := http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		return nil , err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return nil ,err
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return nil ,err
		}

		return nil , errors.New(string(body))
	}

	result:= &types.User{}
	if err := json.NewDecoder(resp.Body).Decode(result) ; err !=nil {
		return nil , err
	}

	return result, nil

}


func inspectUser(userId string) (*types.User ,  error){

	url := fmt.Sprintf("http://localhost:8080/api/users/%s/inspect",userId)

	req , err := http.NewRequest(http.MethodPost , url , nil)
	if err != nil {
		return nil , err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}


	resp, err := client.Do(req)
	if err!=nil {
		return nil ,err
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return nil ,err
		}

		return nil , errors.New(string(body))
	}

	result:= &types.User{}
	if err := json.NewDecoder(resp.Body).Decode(result) ; err !=nil {
		return nil , err
	}

	return result, nil

}


func createUser(user *types.User) (string ,  error){

	buf , err := json.Marshal(user)
	if err!=nil {
		return "" , err
	}

	req , err := http.NewRequest(http.MethodPost , "http://localhost:8080/api/users/create",strings.NewReader(string(buf)))
	if err != nil {
		return "" , err
	}

	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}



	resp, err := client.Do(req)
	if err!=nil {
		return "" ,err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s",err.Error())
		return "" ,err
	}

	if resp.StatusCode != http.StatusOK {
		return "" , errors.New(string(body))
	}

	u:=handlers.UsersCreateResponse{}
	if err := json.Unmarshal(body,&u) ; err !=nil {
		return "" , err
	}


	return u.Id, nil

}

//清理数据库
//避免遗留的测试数据对测试结果造成干扰
func cleanUpDatabase()  {
	cmds := [...][2]string {
		{"mongo", "zanecloud --eval \"db.user.remove({'name':'sadan'})\""},
		{"mongo", "zanecloud --eval \"db.team.remove({'name':'team1'})\""},
		{"mongo", "zanecloud --eval \"db.pool.remove({'name':'pool1'})\""},
	}

	for _, arr := range cmds {
		_, err := exec.Command("sh", "-c", fmt.Sprintf("%s %s", arr[0], arr[1])).Output()
		if err != nil {
			log.Fatal(err)
		}
	}

}