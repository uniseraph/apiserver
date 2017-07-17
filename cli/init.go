package cli

import (
	"github.com/codegangsta/cli"
	"fmt"
	"github.com/zanecloud/apiserver/utils"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/zanecloud/apiserver/types"
)

const initCommandName = "init"

func initCommand(c *cli.Context) {

	//#准备加盐计算
	name := "root"
	salt := "1234567891234567"
	pass := "hell05a"
	content := fmt.Sprintf("%s:%s", pass, salt) //"$pass:$salt"
	//#生成加盐后的密码
	encryptedPassword := utils.Md5(content)



	MgoDB:=     c.String(utils.KEY_MGO_DB)
	MgoURLs:=   c.String(utils.KEY_MGO_URLS)

	session, err := mgo.Dial(MgoURLs)
	if err != nil {
		logrus.Errorf("initCommand::dial mongodb %s  error: %s", MgoURLs, err.Error())
		return
	}
	defer  session.Close()

	err = session.DB(MgoDB).C("user").Insert(bson.M{"name":name,"pass":encryptedPassword,"salt":salt,"roleset":types.ROLESET_SYSADMIN})
	if err != nil {
		logrus.Fatal(err)
	}

}
