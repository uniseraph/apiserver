export default {
  install(Vue, options) {
    Vue.prototype.constants = {
    	ROLE_NORMAL_USER: 0x01,
    	ROLE_APP_ADMIN: 0x02,
    	ROLE_SYS_ADMIN: 0x04,

    	MONTH_LIST: ['一月', '二月', '三月', '四月', '五月', '六月', '七月', '八月', '九月', '十月', '十一月', '十二月'],

    	DAY_LIST: ['日', '一', '二', '三', '四', '五', '六'],

    	APPLICATION_STATUS_MAP: {
    	  "running": "运行中",
    	  "stopped": "已停止",
    	  "*": "未知错误"
    	},

    	APPLICATION_CLASS_MAP: {
    	  "running": "green--text",
    	  "stopped": "orange--text",
    	  "*": "red--text"
    	},

    	CONTAINER_STATUS_MAP: {
    	  "running": "运行中",
    	  "stopped": "已停止",
    	  "*": "未知错误"
    	},

    	CONTAINER_CLASS_MAP: {
    	  "running": "green--text",
    	  "stopped": "orange--text",
    	  "*": "red--text"
    	},

    	TIME_LIST: [
          "00:00","01:00","02:00","03:00","04:00","05:00","06:00","07:00","08:00","09:00","10:00","11:00",
          "12:00","13:00","14:00","15:00","16:00","17:00","18:00","19:00","20:00","21:00","22:00","23:00"
        ],

        SSH_OP_LIST: [
          { Id: 'LoginFailed', Name: '登录失败' },
	      { Id: 'Logined', Name: '登录成功' },
	      { Id: 'ExecCmd', Name: '命令执行' }
        ],

    	MODULE_LIST: [
    	  { Id: 'User', Name: '用户'},
          { Id: 'Team', Name: '团队'},
          { Id: 'Pool', Name: '集群'},
          { Id: 'Env', Name: '参数目录'},
          { Id: 'ApplicationTemplate', Name: '应用模板'},
          { Id: 'Application', Name: '应用管理'}
    	],

    	MODULE_MAP: {
    	  User: { Id: 'User', Name: '用户'},
          Team: { Id: 'Team', Name: '团队'},
          Pool: { Id: 'Pool', Name: '集群'},
          Env: { Id: 'Env', Name: '参数目录'},
          ApplicationTemplate: { Id: 'ApplicationTemplate', Name: '应用模板'},
          Application: { Id: 'Application', Name: '应用管理'}
    	},

    	OPERATION_MAP: {
		  User: [
	        { Id: 'Create', Name: '新增' },
	        { Id: 'Update', Name: '修改' },
	        { Id: 'Delete', Name: '删除' },
	        { Id: 'LoginFailed', Name: '登录失败' },
	        { Id: 'Logined', Name: '登录成功' }
	      ],
	      Team: [
	        { Id: 'Create', Name: '新增' },
	        { Id: 'Update', Name: '修改' },
	        { Id: 'Delete', Name: '删除' },
	        { Id: 'AddUser', Name: '添加成员' },
	        { Id: 'RemoveUser', Name: '删除成员' }
	      ],
	      Pool: [
	        { Id: 'Create', Name: '新增' },
	        { Id: 'Update', Name: '修改' },
	        { Id: 'Delete', Name: '删除' },
	        { Id: 'AuthTeam', Name: '授权团队' },
	        { Id: 'RevokerTeam', Name: '取消授权团队' },
	        { Id: 'AuthUser', Name: '授权用户' },
	        { Id: 'RevokerUser', Name: '取消授权用户' }
	      ],
	      Env: [
	      	{ Id: 'UpdateEnvValue', Name: '修改参数默认值' },
	      	{ Id: 'UpdatePoolValue', Name: '修改集群参数当前值' },
	      ],
	      ApplicationTemplate: [
	      	{ Id: 'Create', Name: '新增' },
	        { Id: 'Update', Name: '修改' },
	        { Id: 'Delete', Name: '删除' }
	      ],
	      Application: [
	      	{ Id: 'Create', Name: '新增' },
	        { Id: 'Delete', Name: '删除' },
	        { Id: 'UpdateServiceReplicaCount', Name: '修改容器个数' },
	        { Id: 'RestartContainer', Name: '重启容器' },
	        { Id: 'Upgrade', Name: '升级应用' },
	        { Id: 'Rollback', Name: '回滚应用' },
	        { Id: 'AuthTeam', Name: '授权团队' },
	        { Id: 'RevokerTeam', Name: '取消授权团队' },
	        { Id: 'AuthUser', Name: '授权用户' },
	        { Id: 'RevokerUser', Name: '取消授权用户' }
	      ]
	    }
    }
  }
}