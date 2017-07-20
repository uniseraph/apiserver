package utils

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/types"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
	校验用户身份
	根据用户提供的密码，对比需要校验的用户实例
	通过对密码和盐做MD5运算，和user.Pass字段比较是否相同
*/
func ValidatePassword(user types.User, pass string) (ok bool, err error) {
	if (len(user.Pass) == 0) || (len(user.Salt) == 0) {
		ok = false
		err = errors.New("Password or Salt is empty!")
		return
	}

	ok = Md5(fmt.Sprintf("%s:%s", pass, user.Salt)) == user.Pass
	logrus.Debugf("getUserLogin::get the user %#v, password is %s, rlt:%d", user, pass, ok)
	return ok, nil
}

//根据用户输入的密码
//生成密码密文和盐
func EncryptedPassword(pass string) (enc string, salt string) {
	//为用户密码加盐
	s := RandomStr(16)
	//生成加密后的密码，数据库中不保存明文密码
	encryptedPassword := Md5(fmt.Sprintf("%s:%s", pass, s))

	return encryptedPassword, s
}

//根据用户查找该用户有权访问的集群ID
func PoolIdsOfUser(c_pool *mgo.Collection, c_team *mgo.Collection, user *types.User) ([]bson.ObjectId, error) {

	poolIds := make([]bson.ObjectId, 0, 20)

	//检查当前用户是否有权限操作该容器
	if user.RoleSet&types.ROLESET_SYSADMIN == types.ROLESET_SYSADMIN {
		//如果用户是系统管理员
		//则不需要校验用户对该机器的权限
		pools := make([]types.PoolInfo, 0, 20)
		if err := c_pool.Find(bson.M{}).All(&pools); err != nil {
			return nil, err
		}
		for _, pool := range pools {
			poolIds = append(poolIds, pool.Id)
		}
		return poolIds, nil
	}

	//已经给当前用户授权过的集群，可以查看
	poolIds = append(poolIds, user.PoolIds...)

	//如果该用户加入过某些团队
	//则该团队能查看的pool
	//该用户也可以查看
	//则验证通过
	if len(user.TeamIds) > 0 {
		teams := make([]types.Team, 0, 10)
		selector := bson.M{
			"_id": bson.M{
				"$in": user.TeamIds,
			},
		}
		//查找该用户所在Team
		if err := c_team.Find(selector).All(&teams); err != nil {
			return nil, err
		}

		//用户所在的某个TEAM
		//拥有授权的集群
		for _, team := range teams {
			poolIds = append(poolIds, team.PoolIds...)
		}
	}

	return poolIds, nil
}
