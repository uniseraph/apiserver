package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/types"
	"encoding/json"
	"strings"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/handlers"
)




func main() {

	client := &http.Client{}

	//登录root用户
	//获得测试使用的登录态cookie
	resp := sessionCreate(client)

	fmt.Println("\n current user is ....")
	currentUser , err := currentUser(client,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the current user err:%s",err.Error())
		return
	}
	fmt.Printf("the current user is  %#v \n", currentUser)



	userId , err := createUser(client,  &types.User{
		Name : "sadan",
		Pass : "1234",
		RoleSet: types.ROLESET_NORMAL ,
		Email : "zhengtao.wuzt@gmail.com",
		Tel        : "18167189863",
	}, resp.Cookies())
	if err!= nil {
		log.Errorf(err.Error())
		return
	}
	fmt.Printf("\ncreate user success , user id is %s\n",userId)
	user,err:= inspectUser(client,userId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the user:%s err:%s",userId,err.Error())
		return
	}
	fmt.Printf("inspect the user %#v \n", user)


	teamId , err := createTeam(client, &types.Team{
		Name : "team1",
		Description: "dev team1",
	} , resp.Cookies())
	if err!= nil {
		log.Errorf(err.Error())
		return
	}
	fmt.Printf("\ncreate team success , team id is %s\n",teamId)
	team,err:= inspectTeam(client,teamId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the team:%s err:%s",teamId,err.Error())
		return
	}
	fmt.Printf("inspect the team %#v\n", team)



	fmt.Println("\nlist all users ,,,,")
	users , err := listUser(client,resp.Cookies())
	if err !=nil {
		log.Errorf("list  users err:%s",err.Error())
		return
	}
	for _ , u := range users {
		fmt.Printf("user: %#v \n", u)
	}

	fmt.Println("\nlist all teams ......")
	teams , err := listTeam(client,resp.Cookies())
	if err !=nil {
		log.Errorf("list  teams err:%s",err.Error())
		return
	}
	for _ , u := range teams {
		fmt.Printf("team: %#v \n", u)
	}


	fmt.Printf("\nuser %s join the team %s \n", userId, teamId)
	if err := joinTeam(client, userId, teamId , resp.Cookies()) ; err != nil {
		fmt.Printf("join fail : %s\n", err.Error())
		return
	}
	user,err= inspectUser(client,userId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the user:%s err:%s",userId,err.Error())
		return
	}
	fmt.Printf("inspect the user %#v \n", user)
	team ,err= inspectTeam(client,teamId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the team:%s err:%s",teamId,err.Error())
		return
	}
	fmt.Printf("inspect the team %#v \n", team)

	fmt.Printf("join success ......\n")

	email :="76577126@qq.com"
	fmt.Printf("\nupdate user %s email:%s\n", userId , email)
	if err := updateUser(client , userId ,email,
		resp.Cookies()); err != nil {
		fmt.Printf("update user:%s email err:%s",userId,err.Error())
		return
	}
	user,err= inspectUser(client,userId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the user:%s err:%s",userId,err.Error())
		return
	}
	fmt.Printf("inspect the user %#v \n", user)

	if user.Email != "76577126@qq.com" {
		return
	}
	fmt.Printf("update user:%s email success....\n",userId)


	fmt.Printf("\nuser %s quit the team %s  ......\n",userId,teamId)
	if err := quitTeam(client , userId , teamId , resp.Cookies()) ; err != nil {
		fmt.Printf("quit fail : %s\n", err.Error())
		return
	}
	user,err = inspectUser(client,userId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the user:%s err:%s",userId,err.Error())
		return
	}
	fmt.Printf("inspect the user %#v \n", user)


	team ,err= inspectTeam(client,teamId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the team:%s err:%s",teamId,err.Error())
		return
	}
	fmt.Printf("inspect the team %#v \n", team)



	fmt.Printf("\nappoint user:%s as the team:%s leader \n", userId,teamId)
	if err := appointTeam(client,teamId,userId,resp.Cookies()) ; err !=nil{
		log.Errorf("appoint the team:%s err:%s",teamId,err.Error())
		return
	}
	fmt.Printf("appoint  the team %s success \n", teamId)

	team,err= inspectTeam(client,teamId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the team:%s err:%s",teamId,err.Error())
		return
	}
	fmt.Printf("inspect the team %#v \n", team)

	fmt.Printf("\nrevoke user:%s as the team:%s leader \n", userId,teamId)
	if err := revokeTeam(client,teamId,userId,resp.Cookies()) ; err !=nil{
		log.Errorf("revoke the team:%s err:%s",teamId,err.Error())
		return
	}
	fmt.Printf("revoke  the team %s success \n", teamId)

	cteam,err:= inspectTeam(client,teamId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the team:%s err:%s",teamId,err.Error())
		return
	}
	fmt.Printf("inspect the team %#v \n", cteam)



	fmt.Printf("\ncheck current root user's actions permission\n")
	actionsCheckResult , err := checkActions(client,[]string{
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
	},resp.Cookies())
	if err !=nil {
		log.Errorf("actionsCheck   err:%s",err.Error())
		return
	}

	for action, b := range actionsCheckResult.Action2Result{
		fmt.Printf("%s:%#v\n" , action , b)
	}
	fmt.Printf("checkaction success...\n" )


	fmt.Println("\ncreate a pool....")

	result , err := registerPool(client,"pool1", &handlers.PoolsRegisterRequest{
		Name: "pool1",
		Driver: "swarm",
		DriverOpts: types.DriverOpts{
			Version:"v1.0",
			EndPoint:"unix:///var/run/docker.sock",
			APIVersion:"v1.23",
		},
	},resp.Cookies())

	if err!=nil {
		log.Errorf("register the pool:%s err:%s","pool1",err.Error())
		return
	}

	fmt.Printf("register success , result is %#v",result)

	//测试退出当前用户功能
	fmt.Println("\nsession destory....")
	err = sessionLogout(client, resp.Cookies())
	if err != nil {
		log.Errorf("session destroy failed!", err.Error())
	}else{
		log.Infof("session destroy success.")
	}

	//登录root用户
	//获得测试使用的登录态cookie
	resp = sessionCreate(client)

	//继续以下测试，依然可以使用resp.Cookies()保持会话
}

//root用户登录
//获取登陆后的cookie
func sessionCreate(client *http.Client) (response *http.Response)  {
	req, err := http.NewRequest("POST", "http://localhost:8080/api/users/root/login?Pass=hell05a" ,nil )
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
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

		log.Errorf("login statuscode:%d err:%s" , resp.StatusCode, string(body))
		return
	}

	fmt.Println(string(body))
	rootUser:=&types.User{}
	json.Unmarshal(body,rootUser)
	fmt.Printf("\nlogin success , the root  user is %#v....",rootUser)

	for _ , cookie := range resp.Cookies(){
		fmt.Println("cookie:", cookie)
	}

	return resp
}


//"/actions/check" : &MyHandler{h: postActionsCheck } ,
func checkActions(client *http.Client , actions []string , cookies []*http.Cookie) (*handlers.ActionCheckResponse ,error) {

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

func joinTeam(client *http.Client , userId string , teamId string , cookies []*http.Cookie)( error){

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




func quitTeam(client *http.Client , userId string , teamId string , cookies []*http.Cookie)( error){

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

func createUser(client *http.Client  , user *types.User , cookies []*http.Cookie) (string ,  error){

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


func createTeam(client *http.Client  , team *types.Team , cookies []*http.Cookie) (string ,  error){

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


func inspectTeam(client *http.Client  , teamId string , cookies []*http.Cookie) (*types.Team ,  error){

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
func currentUser(client *http.Client  ,  cookies []*http.Cookie) (*types.User ,  error){

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


func inspectUser(client *http.Client  , userId string , cookies []*http.Cookie) (*types.User ,  error){

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

//"/teams/{id:.*}/remove":  checkUserPermission(postTeamRemove,types.ROLESET_SYSADMIN),
func revokeTeam(client *http.Client   ,teamId string , userId string , cookies []*http.Cookie) ( error){
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
func appointTeam(client *http.Client   ,teamId string , userId string , cookies []*http.Cookie) ( error){
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

func listUser(client *http.Client   , cookies []*http.Cookie) ([]types.User ,  error){

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

func listTeam(client *http.Client   , cookies []*http.Cookie) ([]types.Team ,  error){

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


	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s",resp.StatusCode,err.Error())
			return nil ,err
		}


		return nil , errors.New(string(body))
	}

	var result []types.Team
	if err := json.NewDecoder(resp.Body).Decode(&result) ; err !=nil {
		//log.Errorf("decode the users buf : %s error:%s" , string(body) , err.Error() )
		return nil , err
	}

	return result, nil

}

//"/users/{id:.*}/update":    &MyHandler{h: postUserUpdate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},

func  updateUser(client *http.Client , userId string ,  email string , cookies []*http.Cookie) error {

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


func registerPool(client *http.Client , name string , request * handlers.PoolsRegisterRequest , cookies []*http.Cookie) (interface{} , error) {

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

func sessionLogout(client *http.Client  ,  cookies []*http.Cookie) (error){

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
	log.Infof("logout success.")

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