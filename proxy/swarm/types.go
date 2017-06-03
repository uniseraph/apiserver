package swarm


// mongodb中Container表，只记录容器创建时间和状态，具体信息需要从集群中获取，避免同步
type Container struct {
	Id          string
	IsDeleted   bool
	GmtDeleted  int64
	GmtCreated  int64
}

