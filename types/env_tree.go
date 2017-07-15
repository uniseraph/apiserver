package types

import (
	"gopkg.in/mgo.v2/bson"
)

/*
	参数目录
	zheng.cui
*/

//参数目录树元数据

//EnvTreeMeta has one EnvTreeNodeDir entry point
type EnvTreeMeta struct {
	Id          bson.ObjectId "_id"
	Name        string
	Description string
	CreatedTime int64 `json:",omitempty"`
	UpdatedTime int64 `json:",omitempty"`
}

//EnvTreeNodeDir has many sub EnvTreeNodeDirs and EnvTreeNodeParamKeys} pairs
//EnvTreeNodeDir belongs to EnvTreeMeta
type EnvTreeNodeDir struct {
	Id   bson.ObjectId "_id"
	Name string
	//一个父目录
	//最顶级的父目录为空，用于结合EnvTreeMeta查询该树的起点
	//EnvTreeNodeDir
	Parent bson.ObjectId `bson:",omitempty"`
	//多个子目录
	//EnvTreeNodeDir
	Children []bson.ObjectId
	//多个值
	//EnvTreeNodeParamKey
	Keys []bson.ObjectId
	//EnvTreeMeta
	Tree        bson.ObjectId
	CreatedTime int64 `json:",omitempty"`
	UpdatedTime int64 `json:",omitempty"`
}

//参数目录树节点的参数名称
//EnvTreeNodeParamKey has many EnvTreeNodeParamValue
//一棵树下，每个Key都要名字唯一
//mongo zanecloud --eval "db.env_tree_node_param_key.createIndex({name:1,tree:1}, {unique:true})"
type EnvTreeNodeParamKey struct {
	Id          bson.ObjectId "_id"
	Name        string
	Description string
	//默认值
	Default string
	//EnvTreeNodeDir
	Dir bson.ObjectId
	//EnvTreeMeta
	Tree        bson.ObjectId
	CreatedTime int64 `json:",omitempty"`
	UpdatedTime int64 `json:",omitempty"`
}

//参数目录树节点的参数值
//EnvTreeNodeParamValue belongs to EnvTreeNodeParamKey
//EnvTreeNodeParamValue belongs to Pool
//这其实是一个Key和Pool的关联关系表
//用来查询一个Key被哪些Pool所用，并且每个值都是什么
type EnvTreeNodeParamValue struct {
	Id    bson.ObjectId "_id"
	Value string
	//对应的参数名称
	//EnvTreeNodeParamKey
	Key bson.ObjectId
	//EnvTreeMeta
	Tree bson.ObjectId
	//PoolInfo
	Pool        bson.ObjectId `bson:",omitempty"`
	CreatedTime int64         `json:",omitempty"`
	UpdatedTime int64         `json:",omitempty"`
}
