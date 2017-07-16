package handlers_test

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/handlers"
	"github.com/zanecloud/apiserver/types"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestUser(t *testing.T) {
	var userId string
	var teamId string
	t.Run("USER=1", func(t *testing.T) {
		currentUser, err := currentUser()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(currentUser)
		}
		userId = currentUser.Id.String()
	})

	t.Run("USER=2", func(t *testing.T) {
		uid, err := createUser(&types.User{
			Name:    "sadan",
			Pass:    "1234",
			RoleSet: types.ROLESET_NORMAL,
			Email:   "zhengtao.wuzt@gmail.com",
			Tel:     "18167189863",
		})
		if err != nil {
			t.Error(err)
		} else {
			t.Log(uid)
		}

		userId = uid
	})

	t.Run("USER=3", func(t *testing.T) {
		user, err := inspectUser(userId)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(user)
		}
		if user.Id.Hex() != userId {
			t.Error("Inspect User error! user id not correct!")
		}
	})

	t.Run("USER=4", func(t *testing.T) {
		users, err := listUser()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(users)
		}

		if len(users) != 2 {
			t.Error("User count not correct")
		}
	})

	//加入team
	t.Run("USER=5", func(t *testing.T) {
		teams, err := listTeam()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(teams)
		}

		if len(teams) != 1 {
			t.Error("Team count is not correct!")
		}

		fmt.Printf("Team: %#v", teams)
		tx := teams[0]
		teamId = tx.Id.Hex()
		if err := joinTeam(userId, teamId); err != nil {
			t.Error(err)
		}
		user, err := inspectUser(userId)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(user)
		}
		//TODO
		//检查某个用户是否已经退出某个TEAM

		team, err := inspectTeam(teamId)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(team)
		}
	})

	t.Run("USER=6", func(t *testing.T) {
		email := "76577126@qq.com"
		if err := updateUser(userId, email); err != nil {
			t.Error(err)
		}
		user, err := inspectUser(userId)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(user)
		}

		if user.Email != email {
			t.Error("User update error for email!")
		}
	})

	t.Run("USER=6", func(t *testing.T) {

		if err := quitTeam(userId, teamId); err != nil {
			t.Error(err)
		}
		user, err := inspectUser(userId)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(user)
		}
		//TOOD
		//检查该用户是否已经退出某个team

		team, err := inspectTeam(teamId)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(team)
		}
	})

	t.Run("USER=7", func(t *testing.T) {
		if err := appointTeam(teamId, userId); err != nil {
			t.Error(err)
		}

		team, err := inspectTeam(teamId)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(team)
		}
		if err := revokeTeam(teamId, userId); err != nil {
			t.Error(err)
		}

		team, err = inspectTeam(teamId)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(team)
		}
	})
}

//"/users/current":           &MyHandler{h: getUserCurrent ,opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
func TestCurrentUser(t *testing.T) {

	url := fmt.Sprintf("http://localhost:8080/api/users/current")

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
			t.Error(err)
		}
		t.Error(string(body))
	}

	result := &types.User{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		t.Error(err)
	}

	if err != nil {
		t.Errorf("inspect the current user err:%s", err.Error())
	} else {
		t.Logf("the current user is  %#v \n", result)
	}
}

//"/users/current":           &MyHandler{h: getUserCurrent ,opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
func currentUser() (*types.User, error) {

	url := fmt.Sprintf("http://localhost:8080/api/users/current")

	req, err := http.NewRequest(http.MethodPost, url, nil)
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

	if resp.StatusCode != http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
			return nil, err
		}

		return nil, errors.New(string(body))
	}

	result := &types.User{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil

}

func inspectUser(userId string) (*types.User, error) {

	url := fmt.Sprintf("http://localhost:8080/api/users/%s/inspect", userId)

	req, err := http.NewRequest(http.MethodPost, url, nil)
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

	if resp.StatusCode != http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
			return nil, err
		}

		return nil, errors.New(string(body))
	}

	result := &types.User{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil

}

func createUser(user *types.User) (string, error) {

	buf, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/users/create", strings.NewReader(string(buf)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s", err.Error())
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}

	u := handlers.UsersCreateResponse{}
	if err := json.Unmarshal(body, &u); err != nil {
		return "", err
	}

	return u.Id, nil

}

func listUser() ([]types.User, error) {

	url := fmt.Sprintf("http://localhost:8080/api/users/ps")

	req, err := http.NewRequest(http.MethodPost, url, nil)
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

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
			return nil, err
		}

		return nil, errors.New(string(body))
	}

	var result []types.User
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		//log.Errorf("decode the users buf : %s error:%s" , string(body) , err.Error() )
		return nil, err
	}

	return result, nil

}

//"/users/{id:.*}/update":    &MyHandler{h: postUserUpdate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},

func updateUser(userId string, email string) error {

	url := fmt.Sprintf("http://localhost:8080/api/users/%s/update", userId)

	r := handlers.UserUpdateRequest{}
	r.Email = email
	r.Roleset = types.ROLESET_APPADMIN

	buf, _ := json.Marshal(&r)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(buf)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
			return err
		}

		return errors.New(string(body))
	}

	return nil
}
