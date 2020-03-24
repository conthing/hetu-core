package http

import (
	"bytes"
	"fmt"
	"hetu-core/dto"
	"io/ioutil"
	gohttp "net/http"

	"github.com/conthing/utils/common"
)

// Publish 上报
func Publish(info *dto.PubHTTPInfo, mJSON []byte) {
	url := fmt.Sprintf("%s:%d%s", info.Address, info.Port, info.URL)
	buf := bytes.NewBuffer(mJSON)
	resp, err := gohttp.Post(url, "application/json", buf)
	if err != nil {
		common.Log.Error("HTTP 上报失败", err)
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		common.Log.Error("HTTP 上报失败", err)
		return
	}
	common.Log.Info("HTTP 上报成功")
}
