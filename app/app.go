package app

import "github.com/zanecloud/apiserver/types"

type App interface {
	Create() error
	Start()  error

}




func CreateApplication(app *types.Application , pool *types.PoolInfo ) error {




	return nil


}