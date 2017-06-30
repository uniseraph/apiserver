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

	fmt.Println("login success ....")
	fmt.Println(string(body))


	for _ , cookie := range resp.Cookies(){
		fmt.Println("cookie:", cookie)
	}


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

	fmt.Printf("create user success , user id is %s\n",userId)




	teamId , err := createTeam(client, &types.Team{
		Name : "team1",
		Description: "dev team1",
	} , resp.Cookies())
	if err!= nil {
		log.Errorf(err.Error())
		return
	}

	fmt.Printf("create team success , team id is %s\n",teamId)

	team,err:= inspectTeam(client,teamId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the team:%s err:%s",teamId,err.Error())
		return
	}
	fmt.Printf("inspect the team %#v", team)


	user,err:= inspectUser(client,userId,resp.Cookies())
	if err !=nil {
		log.Errorf("inspect the user:%s err:%s",userId,err.Error())
		return
	}
	fmt.Printf("inspect the user %#v \n", user)


	users , err := listUser(client,resp.Cookies())
	if err !=nil {
		log.Errorf("list  users err:%s",err.Error())
		return
	}

	fmt.Println("list all user ,,,,")
	for _ , u := range users {
		fmt.Printf("list the user %#v \n", u)
	}
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