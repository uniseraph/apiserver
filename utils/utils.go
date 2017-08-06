package utils

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/zanecloud/apiserver/types"
	"gopkg.in/mgo.v2"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//计算MD5
func Md5(data string) string {
	h := md5.New()
	io.WriteString(h, data)
	str := fmt.Sprintf("%x", h.Sum(nil))
	logrus.Debugf("MD5 for string: %s, hash is %s", data, str)
	return str
}

func HttpError(w http.ResponseWriter, err string, status int) {
	logrus.WithField("status", status).Errorf("HTTP error: %v", err)
	http.Error(w, err, status)
}

//请求处理结果成功的标准操作
func HttpOK(w http.ResponseWriter, result interface{}) {
	//body := map[string]interface{}{
	//	"status" : "0",
	//	"msg"	 : "Success",
	//}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	//if result != nil {
	//	body["data"] = result
	//	json.NewEncoder(w).Encode(body)
	//}else{
	//	json.NewEncoder(w).Encode(body)
	//}
	if result != nil {
		json.NewEncoder(w).Encode(result)
	}
}

//生成随机字符串，长度为n
func RandomStr(n int) string {
	if n > 0 {
		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}
		s := fmt.Sprintf("%X", b)

		return s
	}
	return ""
}

//当前时间的int64格式返回值
func TimeNow64() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//将解析HTTP请求的body解析成JSON对象
//存储到对应的req模型中
//func HttpRequestBodyJsonParse(w http.ResponseWriter, r *http.Request, req interface{})  {
//	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
//		HttpError(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//}

//得到数据库表连接后的HTTP请求回调函数
type mgoCollectionsCallback func(cs map[string]*mgo.Collection)

//统一管理数据库
//批量获取表连接
//使用闭包处理API的业务逻辑
func GetMgoCollections(ctx context.Context, w http.ResponseWriter, names []string, cb mgoCollectionsCallback) {
	mgoSession, err := GetMgoSessionClone(ctx)
	if err != nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := GetAPIServerConfig(ctx).MgoDB

	var cs = make(map[string]*mgo.Collection)
	for _, name := range names {
		c := mgoSession.DB(mgoDB).C(name)
		cs[name] = c
	}

	cb(cs)
}

type execCallback func(cs map[string]*mgo.Collection) error

func ExecMgoCollections(ctx context.Context, names []string, cb execCallback) error {
	mgoSession, err := GetMgoSessionClone(ctx)
	if err != nil {
		//走不到这里的,ctx中必然有mgoSesson
		return err
	}
	defer mgoSession.Close()

	mgoDB := GetAPIServerConfig(ctx).MgoDB

	var cs = make(map[string]*mgo.Collection)
	for _, name := range names {
		c := mgoSession.DB(mgoDB).C(name)
		cs[name] = c
	}

	return cb(cs)
}

//

func GetClusterInfo(ctx context.Context, endpoint string) (*types.ClusterInfo, error) {

	_, addr, _, err := client.ParseHost(endpoint)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s/info", addr)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	swarmClient := http.Client{}

	resp, err := swarmClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, _ := ioutil.ReadAll(resp.Body)
	//logrus.Debugf("docker info result is %s", string(buf))

	result := &types.ClusterInfo{}
	if err := json.Unmarshal(buf, result); err != nil {
		return nil, err
	}

	return result, nil

}

//"SystemStatus": [
//[
//"Role",
//"primary"
//],
//[
//"Strategy",
//"spread"
//],
//[
//"Filters",
//"health, port, containerslots, dependency, affinity, constraint, whitelist"
//],
//[
//"Nodes",
//"3"
//],
//[
//" iZ8vbcm5gh7j8t4gwe8oh1Z",
//"172.16.200.35:2376"
//],
//[
//"  └ ID",
//"Y4QM:7IVB:SVGT:FI2F:AK44:ROZM:N5HW:V7TM:EHBA:QTX6:A5HK:WOGC"
//],
//[
//"  └ Status",
//"Healthy"
//],
//[
//"  └ Containers",
//"4 (4 Running, 0 Paused, 0 Stopped)"
//],
//[
//"  └ Reserved CPUs",
//"0 / 1"
//],
//[
//"  └ Reserved Memory",
//"0 B / 1.886 GiB"
//],
//[
//"  └ Labels",
//"kernelversion=3.10.0-514.21.1.el7.x86_64, operatingsystem=CentOS Linux 7 (Core), storagedriver=overlay"
//],
//[
//"  └ UpdatedAt",
//"2017-07-06T11:20:15Z"
//],
//[
//"  └ ServerVersion",
//"1.12.6"
//],
//[
//" iZ8vbcm5gh7j8t4gwe8oh2Z",
//"172.16.200.34:2376"
//],
//[
//"  └ ID",
//"KRHJ:M4SZ:KP2X:SO2I:EJZ2:5F7E:LSVU:CO2P:ZEHI:CSAH:A4K7:3MHK"
//],
//[
//"  └ Status",
//"Healthy"
//],
//[
//"  └ Containers",
//"3 (3 Running, 0 Paused, 0 Stopped)"
//],
//[
//"  └ Reserved CPUs",
//"0 / 1"
//],
//[
//"  └ Reserved Memory",
//"0 B / 1.886 GiB"
//],
//[
//"  └ Labels",
//"kernelversion=3.10.0-514.21.1.el7.x86_64, operatingsystem=CentOS Linux 7 (Core), storagedriver=overlay"
//],
//[
//"  └ UpdatedAt",
//"2017-07-06T11:20:08Z"
//],
//[
//"  └ ServerVersion",
//"1.12.6"
//],
//[
//" iZ8vbcm5gh7j8t4gwe8oh3Z",
//"172.16.200.36:2376"
//],
//[
//"  └ ID",
//"7XLD:HJ35:QPY6:KMYS:RGGT:ABTM:BAPL:BC5V:5HTA:4RR5:QCKU:QQL7"
//],
//[
//"  └ Status",
//"Healthy"
//],
//[
//"  └ Containers",
//"3 (3 Running, 0 Paused, 0 Stopped)"
//],
//[
//"  └ Reserved CPUs",
//"0 / 1"
//],
//[
//"  └ Reserved Memory",
//"0 B / 1.886 GiB"
//],
//[
//"  └ Labels",
//"kernelversion=3.10.0-514.21.1.el7.x86_64, operatingsystem=CentOS Linux 7 (Core), storagedriver=overlay"
//],
//[
//"  └ UpdatedAt",
//"2017-07-06T11:20:02Z"
//],
//[
//"  └ ServerVersion",
//"1.12.6"
//]
//],
func ParseNodes(input [][]string, poolId string) (string, string, []types.Node, error) {

	strategy := input[1][1]
	filters := input[2][1]

	nodes, _ := strconv.Atoi(input[3][1])

	result := make([]types.Node, nodes)

	for i := 0; i < nodes; i++ {
		result[i] = types.Node{
			PoolId: poolId,
			//PoolName: poolName,
		}

		result[i].Hostname = input[4+i*9][0]
		result[i].Endpoint = input[4+i*9][1]
		result[i].NodeId = input[5+i*9][1]
		result[i].Status = input[6+i*9][1]
		result[i].Containers = input[7+i*9][1]

		result[i].ReservedCPUs = input[8+i*9][1]
		result[i].ReservedMemory = input[9+i*9][1]

		result[i].Labels = parseLabels(input[10+i*9][1])

		//ignore 11-UpdateAt
		result[i].ServerVersion = input[12+i*9][1]

	}

	return strategy, filters, result, nil
}
func parseLabels(labels string) map[string]string {

	result := make(map[string]string)

	for _, value := range strings.Split(labels, ",") {

		key2value := strings.Split(value, "=")

		if len(key2value) != 2 {
			logrus.Debugf("parseLabels:: err lable :%s", value)
			continue
		}

		result[key2value[0]] = key2value[1]

	}

	return result

}

//查询请求可以直接到后端集群，bypass proxy
func CreateDockerClient(poolInfo *types.PoolInfo) (*client.Client, error) {

	var httpClient *http.Client
	if poolInfo.DriverOpts.TlsConfig != nil {
		tlsc, err := tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
		if err != nil {
			return nil, err
		}
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsc,
			},
			CheckRedirect: client.CheckRedirect,
		}
	}

	cli, err := client.NewClient(poolInfo.DriverOpts.EndPoint, poolInfo.DriverOpts.APIVersion, httpClient, nil)
	if err != nil {
		return nil, err
	}

	return cli, nil

}
