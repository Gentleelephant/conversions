/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"github.com/prometheus/alertmanager/template"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
	"sort"
	"strings"
	"time"
)

var (
	port      int
	address   string
	alertId   string
	alertKey  string
	applyType string
)

type AlertNotify struct {
	emails  []string
	mobiles []string
	wechats []string
}

type Lightning struct {
	AlertName   string       `json:"alertName"`
	AlertDesc   string       `json:"alertDesc"`
	AlertLevel  string       `json:"alertLevel,omitempty"`
	AlertStatus string       `json:"alertStatus"`
	ApplyType   string       `json:"applyType"`
	AlertNotify *AlertNotify `json:"alertNotify,omitempty"`
	AlertKey    string       `json:"alertKey"`
	AlertId     string       `json:"alertId"`
	AlertMsgId  string       `json:"alertMsgId"`
}

func AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&port, "port", 8080, "The port which the server listen, default 8080")
	fs.StringVar(&address, "address", "", "The address of send the data to")
	fs.StringVar(&alertId, "alertId", "", "The alert id")
	fs.StringVar(&alertKey, "alertKey", "", "The alert key")
	fs.StringVar(&applyType, "applyType", "custom", "The apply type")
}

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "webhook",
		Long: `The ks alert webhook to conversion event`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run()
		},
	}
	AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	return cmd
}

func Run() error {
	pflag.VisitAll(func(flag *pflag.Flag) {
		glog.Errorf("FLAG: --%s=%q", flag.Name, flag.Value)
	})

	log.Println("Starting conversions server...")

	return httpServer()
}

func httpServer() error {
	container := restful.NewContainer()
	ws := new(restful.WebService)
	ws.Path("").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/alert").To(alertHandler))

	container.Add(ws)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: container,
	}

	var err error

	err = server.ListenAndServe()

	return err
}

// 处理接收到的Alertmanager告警
func alertHandler(req *restful.Request, resp *restful.Response) {
	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		err := resp.WriteHeaderAndEntity(http.StatusBadRequest, "")
		if err != nil {
			glog.Errorf("response error %s", err)
		}
		return
	}
	defer req.Request.Body.Close()
	// Parse alerts sent through Alertmanager webhook, more detail please refer to
	// https://github.com/prometheus/alertmanager/blob/master/template/template.go#L231
	var alert template.Data
	var data []Lightning
	if err = json.Unmarshal(body, &alert); err != nil {
		log.Println(err)
		return
	}

	for _, item := range alert.Alerts {
		ligthning := Lightning{}
		ligthning.AlertName = item.Labels["alertname"]
		ligthning.AlertDesc = item.Annotations["message"]
		ligthning.AlertStatus = item.Status
		ligthning.AlertId = alertId
		ligthning.AlertKey = alertKey
		ligthning.ApplyType = applyType // preset or custom
		ligthning.AlertMsgId = generateUniqueMsgID(item.Labels)
		if ligthning.AlertLevel == "" {
			ligthning.AlertLevel = "INFO"
		}
		if item.Status == "firing" {
			ligthning.AlertLevel = strings.ToUpper(item.Labels["severity"])
			if ligthning.AlertLevel == "ERROR" {
				ligthning.AlertLevel = "CRITICAL"
			}
		} else {
			ligthning.AlertLevel = ""
		}

		data = append(data, ligthning)
	}

	sendToGlowworm(data)

	// 返回成功响应
	resp.WriteHeader(http.StatusOK)
	fmt.Fprint(resp, "Alert received and forwarded to Firefly")
}

// 发送数据给萤火虫
func sendToGlowworm(data []Lightning) {
	// 实现发送逻辑，将数据发送给萤火虫
	// 例如使用HTTP POST请求发送数据给萤火虫的API
	client := http.Client{
		Timeout: 5 * time.Second, // 设置连接超时时间为 5 秒
	}

	url, err := url2.Parse(address)
	if err != nil {
		log.Println(err)
		return
	}
	address = url.Scheme + "://" + url.Host + url.Path

	for _, item := range data {

		dataByte, err := json.Marshal(item)
		if err != nil {
			log.Fatal(err)
		}
		reader := bytes.NewReader(dataByte)
		request, err := http.NewRequest("POST", address, reader)
		if err != nil {
			log.Printf(err.Error())
			return
		}

		request.Header.Set("Content-Type", "application/json; charset=utf-8")

		// 发送请求
		resp, err := client.Do(request)
		if err != nil {
			log.Println(err)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}

		// 打印响应
		fmt.Printf("Response: %s\n", string(body))
		resp.Body.Close()

	}

}

func generateUniqueMsgID(alertInfo map[string]string) string {
	// 对 labels 键进行排序,确保生成的 msgid 是确定的
	keys := make([]string, 0, len(alertInfo))
	for k := range alertInfo {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接所有 label 值,用作 msgid 的基础
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(alertInfo[k])
	}
	baseStr := sb.String()

	// 使用 MD5 哈希生成 msgid
	hasher := md5.New()
	hasher.Write([]byte(baseStr))
	msgID := hex.EncodeToString(hasher.Sum(nil))

	// 如果长度超过 18 位,则截断
	if len(msgID) > 18 {
		msgID = msgID[:18]
	}

	return msgID
}
