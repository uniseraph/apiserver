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

func TestTeam(t *testing.T) {
	var teamId string
	t.Run("TEAM=1", func(t *testing.T) {
		var err error
		teamId, err = createTeam(&types.Team{
			Name:        "team1",
			Description: "dev team1",
		})
		if err != nil {
			t.Error(err)
		} else {
			t.Log(teamId)
		}
	})
	t.Run("TEAM=2", func(t *testing.T) {
		team, err := inspectTeam(teamId)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(team)
		}
	})
	t.Run("TEAM=3", func(t *testing.T) {
		teams, err := listTeam()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(teams)
		}

		if len(teams) != 1 {
			t.Error("Team count is not correct!")
		}
	})
}

//"/teams/{id:.*}/remove":  checkUserPermission(postTeamRemove,types.ROLESET_SYSADMIN),
func revokeTeam(teamId string, userId string) error {
	url := fmt.Sprintf("http://localhost:8080/api/teams/%s/revoke?UserId=%s", teamId, userId)

	req, err := http.NewRequest(http.MethodPost, url, nil)
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
			log.Debugf("appoint read body statuscode:%d err:%s", resp.StatusCode, err.Error())
			return err
		}

		return errors.New(string(body))
	}

	return nil

}

//"/teams/{id:.*}/appoint?UserId=xxx": checkUserPermission(postTeamAppoint,types.ROLESET_SYSADMIN),
func appointTeam(teamId string, userId string) error {
	url := fmt.Sprintf("http://localhost:8080/api/teams/%s/appoint?UserId=%s", teamId, userId)

	req, err := http.NewRequest(http.MethodPost, url, nil)
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
			log.Debugf("appoint read body statuscode:%d err:%s", resp.StatusCode, err.Error())
			return err
		}

		return errors.New(string(body))
	}

	return nil

}

func createTeam(team *types.Team) (string, error) {
	buf, err := json.Marshal(team)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/teams/create", strings.NewReader(string(buf)))
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

	result := handlers.TeamsCreateResponse{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Id, nil

}

func inspectTeam(teamId string) (*types.Team, error) {

	url := fmt.Sprintf("http://localhost:8080/api/teams/%s/inspect", teamId)

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

	result := &types.Team{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil

}

/*

	以下是测试用例的辅助方法

*/

func listTeam() ([]types.Team, error) {

	url := fmt.Sprintf("http://localhost:8080/api/teams/ps")

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

	body, _ := ioutil.ReadAll(resp.Body)

	//fmt.Printf("listTeam repos body is %s\n", string(body))
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
			return nil, err
		}

		return nil, errors.New(string(body))
	}

	var result []types.Team
	if err := json.Unmarshal(body, &result); err != nil {
		//log.Errorf("decode the users buf : %s error:%s" , string(body) , err.Error() )
		return nil, err
	}

	return result, nil

}

func joinTeam(userId string, teamId string) error {

	url := fmt.Sprintf("http://localhost:8080/api/users/%s/join?TeamId=%s", userId, teamId)

	req, err := http.NewRequest(http.MethodPost, url, nil)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s", err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	return nil
}

func quitTeam(userId string, teamId string) error {

	url := fmt.Sprintf("http://localhost:8080/api/users/%s/quit?TeamId=%s", userId, teamId)

	req, err := http.NewRequest(http.MethodPost, url, nil)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("quitTeam read body err:%s", err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	return nil
}
