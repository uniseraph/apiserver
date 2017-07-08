package application

import "github.com/zanecloud/apiserver/types"

type Application interface {
	Create() error
	Start()  error

}




func CreateApplication(app *types.Application , pool *types.PoolInfo ) error {




	return nil


}