package cli

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const initCommandName = "init"

func initCommand(c *cli.Context) {

	//#准备加盐计算
	name := "root"
	pass := "hell05a"

	MgoDB := c.String(utils.KEY_MGO_DB)
	MgoURLs := c.String(utils.KEY_MGO_URLS)

	session, err := mgo.Dial(MgoURLs)
	if err != nil {
		logrus.Errorf("initCommand::dial mongodb %s  error: %s", MgoURLs, err.Error())
		return
	}
	defer session.Close()

	enc, salt := utils.EncryptedPassword(pass)

	user := types.User{
		Id:      bson.NewObjectId(),
		Name:    name,
		Pass:    enc,
		Salt:    salt,
		RoleSet: types.ROLESET_SYSADMIN,
	}

	err = session.DB(MgoDB).C("user").Insert(user)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("init user:%s success ..." , name)

}
