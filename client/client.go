package main

import (
	"net/http"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"fmt"
)

func main() {

	client := &http.Client{}

	req, err := http.NewRequest("POST", "http://localhost:8080/users/root/login?Pass=hell05a" ,nil )
	if err != nil {
		logrus.Errorf(err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err !=nil {
		logrus.Errorf("login post err:%s", err.Error())
		return
	}


	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Debugf("login read body err:%s",err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {

		logrus.Errorf("login statuscode:%s err:%s" , resp.StatusCode, err.Error())
		return
	}

	fmt.Println("login success ....")

	fmt.Println(string(body))
}
