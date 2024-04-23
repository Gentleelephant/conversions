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
	"conversions/template"
	"conversions/utils"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	port          int

)

type AlertNotify struct {
	emails []string
	mobiles []string
	wechats []string
}

type Lightning struct {
	alertName string
	alertDesc string
	alertLevel string
	alertStatus string
	applyType string
	alertNotify AlertNotify
	alertKey string
	alertId string
	alertMsgId string
}


func AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&port, "port", 8080, "The port which the server listen, default 8080")
}

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "webhook",
		Long: `The ks alert webhook to conversion event`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run()
		},
	}
	 fmt.Printf("2")
	AddFlags(cmd.Flags())
	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	return cmd
}

func Run() error {
 fmt.Printf("3")
	pflag.VisitAll(func(flag *pflag.Flag) {
		glog.Errorf("FLAG: --%s=%q", flag.Name, flag.Value)
	})
	return httpServer()
}

func httpServer() error {
 fmt.Printf("4")
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

		// Parse alerts sent through Alertmanager webhook, more detail please refer to
	// https://github.com/prometheus/alertmanager/blob/master/template/template.go#L231
	data := template.Data{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf(err.Error())
		return
	}
	// 读取请求体
	 sendLightning := make([]Lightning, len(data.Alerts))
	 fmt.Print(data)
	 fmt.Sprintf("%v", data)

	for i, alert := range data.Alerts {
		sendLightning[i].alertName = alert.Labels["alertname"]
		alert.ID = utils.Hash(alert)

	}

	// 发送给萤火虫
	//sendToFirefly(sendLightning)

	// 返回成功响应
	resp.WriteHeader(http.StatusOK)
	fmt.Fprint(resp, "Alert received and forwarded to Firefly")
}

// 发送数据给萤火虫
//func sendToFirefly(data map[string]interface{}) {
//	// 实现发送逻辑，将数据发送给萤火虫
//	// 例如使用HTTP POST请求发送数据给萤火虫的API
//}

