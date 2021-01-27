package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	sdk "code.xxxxx.cn/platform/auth/sdk/golang"
	log "github.com/sirupsen/logrus"
)

var srcClientID = flag.Int64("src", 0, "client id")
var dstClientID = flag.Int64("dst", 0, "client id")
var srcSecret = flag.String("src_secret", "", "client secret")
var dstSecret = flag.String("dst_secret", "", "client secret")
var authHost = flag.String("host", "", "auth server host")

func main() {
	flag.Parse()

	apiConfig := sdk.APIConfig{
		ClientID:     *srcClientID,
		ClientSecret: *srcSecret,
		APIHost:      *authHost,
	}

	authClient := sdk.NewApiAuth(&apiConfig)
	_, err := authClient.GetClient()
	if err != nil {
		log.WithError(err).Error("sync failed!")
		return
	}

	apiConfig = sdk.APIConfig{
		ClientID:     *dstClientID,
		ClientSecret: *dstSecret,
		APIHost:      *authHost,
	}

	authClient = sdk.NewApiAuth(&apiConfig)
	_, err = authClient.GetClient()
	if err != nil {
		log.WithError(err).Error("sync failed!")
		return
	}

	params := url.Values{
		"src_id": []string{strconv.Itoa(int(*srcClientID))},
		"dst_id": []string{strconv.Itoa(int(*dstClientID))},
	}

	_, err = doRequest(*dstSecret, *authHost, "POST", "syncClients?"+params.Encode(), *dstClientID)

	if err != nil {
		log.WithError(err).Error("sync failed")
		return
	}
	log.Info("Sync Succeed!")
}

// 统一对接口返回结果进行处理，将有效数据部分序列化后返回
func processResp(response *http.Response) (data []byte, err error) {
	var statusErr error
	if response.StatusCode != http.StatusOK {
		statusErr = errors.New("unexpected status code of " + strconv.Itoa(response.StatusCode))
	}
	rawBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body error when process response: %v, statusErr: %v", err, statusErr)
	}

	type RespBody struct {
		ResCode int         `json:"res_code"` // auth接口返回状态码，SUCC-0-成功  FAILED-1-失败  UNKNOWN-2-其他
		ResMsg  string      `json:"res_msg"`  // auth接口返回状态描述
		Data    interface{} `json:"data"`     // auth接口返回数据
	}

	var body RespBody

	if err = json.Unmarshal(rawBody, &body); err != nil {
		return nil, fmt.Errorf("Unmarshal body error when process response: %v, statusErr: %v", err, statusErr)
	}
	if body.ResCode != 0 && body.Data == nil {
		return nil, fmt.Errorf("Unexpected response body, msg: %v, statusErr: %v", body.ResMsg, statusErr)
	}
	if statusErr != nil {
		return nil, statusErr
	}
	if data, err = json.Marshal(body.Data); err != nil {
		return nil, err
	}
	return
}

func doRequest(clientSecret, host, method, resource string, clientID int64) ([]byte, error) {
	req, err := http.NewRequest(method, host+"/"+resource, nil)
	if err != nil {
		return nil, fmt.Errorf("Create New Request error: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req.Header.Add(
		"Authorization",
		"Client "+sdk.GenerateJWTToken(clientID, clientSecret))

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Do request error: %v", err)
	}
	defer resp.Body.Close()
	return processResp(resp)
}
