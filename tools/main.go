package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"errors"

	"encoding/json"
	"fmt"
	"github.com/zanecloud/apiserver/handlers"
	"github.com/zanecloud/apiserver/types"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

const startCommandName = "start"

type Options struct {
	User          string
	Pass          string
	ApplicationId string
	TemplateId    string
	ImageName     string
	ImageTag      string
	ServiceName   string
	APIServerHost string
}

func startCommand(c *cli.Context) {

	config, err := parseConfig(c)
	if err != nil {
		log.Fatal(err)
		return
	}

	client := &http.Client{}

	cookies, err := login(client, config.APIServerHost, config.User, config.Pass)

	if err != nil {
		log.Fatal(err)
		return
	}
	log.Info("login success...")

	template, err := inspectTemplate(client, config.APIServerHost, config.TemplateId, cookies)
	if err != nil {
		log.Fatal(err)
		return
	}

	mergeTemplate(template, config)

	if err := updateTemplate(client, config, template, cookies); err != nil {
		log.Fatal(err)
		return
	}

	log.Info("update template success..")

	if err := upgradeApplication(client, config.APIServerHost, config.ApplicationId, config.TemplateId, cookies); err != nil {
		log.Fatal(err)
		return
	}

	log.Info("upgrade succes...")
}

func upgradeApplication(client *http.Client, host, applicationId, templateId string, cookies []*http.Cookie) error {

	url := fmt.Sprintf("http://%s/api/applications/%s/upgrade", host, applicationId)

	r := handlers.ApplicationUpgradeRequest{
		ApplicationTemplateId: templateId,
	}

	buf, _ := json.Marshal(r)

	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(string(buf)))

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
		log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	//fmt.Println(string(body))
	//result := handlers.ApplicationCreateResponse{}
	//json.Unmarshal(body, &result)

	return nil
}

func mergeTemplate(template *types.Template, config *Options) {

	for i, _ := range template.Services {
		if template.Services[i].Name == config.ServiceName {
			template.Services[i].ImageName = config.ImageName
			template.Services[i].ImageTag = config.ImageTag

		}
	}

}

func updateTemplate(client *http.Client, config *Options, template *types.Template, cookies []*http.Cookie) error {
	url := fmt.Sprintf("http://%s/api/templates/%s/update", config.APIServerHost, template.Id.Hex())

	r := handlers.TemplateUpdateRequest{
		*template,
	}
	data, _ := json.Marshal(r)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(data)))
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

	result := &types.Template{}
	if err := json.Unmarshal(body, result); err != nil {
		return err
	}

	return nil
}

func inspectTemplate(client *http.Client, host string, id string, cookies []*http.Cookie) (*types.Template, error) {
	url := fmt.Sprintf("http://%s/api/templates/%s/inspect", host, id)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	result := &types.Template{}
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}

	return result, nil
}

//root用户登录
//获取登陆后的cookie
func login(client *http.Client, host string, user string, pass string) ([]*http.Cookie, error) {

	url := fmt.Sprintf("http://%s/api/users/%s/login?Pass=%s", host, user, pass)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Errorf(err.Error())
		return []*http.Cookie{}, nil
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("login post err:%s", err.Error())
		return []*http.Cookie{}, nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s", err.Error())
		return []*http.Cookie{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("login statuscode:%d err:%s", resp.StatusCode, string(body)))
	}

	fmt.Println(string(body))
	//if err := json.Unmarshal(body, user); err != nil {
	//	return nil, err
	//}
	//log.Infof("Resp body: %s", string(body))

	return resp.Cookies(), nil
}

func main() {

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "application daily build tools"
	app.Author = "zhengtao.wuzt"
	app.Email = "zhengtao.wuzt@ongo360"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-level, l",
			Value:  "info",
			EnvVar: "LOG_LEVEL",
			Usage:  "Log level (options: debug, info, warn, error, fatal, panic)",
		},
	}

	app.Before = func(c *cli.Context) error {
		log.SetOutput(os.Stderr)
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.SetLevel(level)
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  startCommandName,
			Usage: "更新模版的镜像并升级应用 ",
			Flags: []cli.Flag{

				cli.StringFlag{
					Name:   "apiserver_host",
					Value:  "localhost:8080",
					EnvVar: "APISERVER_URL",
					Usage: "apiserver	url",
				},
				cli.StringFlag{
					Name:   "template_uuid",
					EnvVar: "TEMPLATE_UUID",
					Usage:  "template uuid",
				},
				cli.StringFlag{
					Name:   "application_uuid",
					EnvVar: "",
					Usage:  "application_uuid",
				},
				cli.StringFlag{
					Name:  "service_name",
					Usage: "服务名",
				},
				cli.StringFlag{
					Name:   "image_name",
					EnvVar: "IMAGE_NAME",
					Usage:  "image name",
				},
				cli.StringFlag{
					Name:   "image_tag",
					EnvVar: "IMAGE_TAG",
					Usage:  "image_tag(0.1.0-xxxxx)",
				},
				cli.StringFlag{
					Name:  "user",
					Usage: "user",
				},
				cli.StringFlag{
					Name:  "pass",
					Usage: "pass",
				},
			},
			Action: startCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
func parseConfig(c *cli.Context) (*Options, error) {

	options := &Options{}

	if user := c.String("user"); user == "" {
		return nil, errors.New("user can't be empty")
	} else {
		options.User = user
	}

	if pass := c.String("pass"); pass == "" {
		return nil, errors.New("pass cant'bt empty")
	} else {
		options.Pass = pass
	}

	if image_tag := c.String("image_tag"); image_tag == "" {
		return nil, errors.New("image_tag cant'bt empty")
	} else {
		options.ImageTag = image_tag
	}
	if image_name := c.String("image_name"); image_name == "" {
		return nil, errors.New("image_name cant'bt empty")
	} else {
		options.ImageName = image_name
	}
	if application_uuid := c.String("application_uuid"); application_uuid == "" {
		return nil, errors.New("application-uuid cant'bt empty")
	} else {
		options.ApplicationId = application_uuid
	}
	if template_uuid := c.String("template_uuid"); template_uuid == "" {
		return nil, errors.New("template_uuid cant'bt empty")
	} else {
		options.TemplateId = template_uuid
	}

	if service_name := c.String("service_name"); service_name == "" {
		return nil, errors.New("service_name cant'bt empty")
	} else {
		options.ServiceName = service_name
	}

	if apiserver_host := c.String("apiserver_host"); apiserver_host == "" {
		return nil, errors.New("apiserver_host cant'bt empty")
	} else {
		options.APIServerHost = apiserver_host
	}
	return options, nil

}
