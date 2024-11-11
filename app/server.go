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
	"time"
)

var (
	port          int
	mergeKey      string
	identifyKey   string
	networkDomain string
	notifyAddress []string
)

type Alert struct {
	Name          string                   `json:"name"`
	Severity      uint                     `json:"severity"`
	Description   string                   `json:"description"`
	OccurTime     int64                    `json:"occur_time"`
	EntityName    string                   `json:"entity_name"`
	EntityAddr    string                   `json:"entity_addr,omitempty"`
	MergedKey     string                   `json:"merged_key,omitempty"`
	IdentifyKey   string                   `json:"identify_key,omitempty"`
	Type          string                   `json:"type"`
	NetworkDomain string                   `json:"networkDomain"`
	Properties    []map[string]interface{} `json:"properties,omitempty"`
}

func AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&port, "port", 8080, "The port which the server listen, default 8080")
	fs.StringVar(&mergeKey, "mergeKey", "entity_name,name", "The merge key")
	fs.StringVar(&identifyKey, "identifyKey", "", "The identify key")
	fs.StringVar(&networkDomain, "networkDomain", "defaultZone", "The network domain")
	fs.StringSliceVar(&notifyAddress, "notifyAddress", []string{}, "The addresses of send the data to")
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

	fmt.Println("notifyAddress:", notifyAddress)
	fmt.Println("mergeKey:", mergeKey)
	fmt.Println("identifyKey:", identifyKey)
	fmt.Println("networkDomain:", networkDomain)

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
	var data []Alert
	if err = json.Unmarshal(body, &alert); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("alert:", alert)

	for _, item := range alert.Alerts {
		currentAlert := Alert{}
		currentAlert.Name = item.Labels["alertname"]
		currentAlert.Description = item.Annotations["message"]

		if currentAlert.EntityName == "" {
			if item.Labels["node"] != "" {
				currentAlert.EntityName = item.Labels["node"]
				currentAlert.EntityAddr = item.Labels["host_ip"]
			} else if item.Labels["pod"] != "" {
				currentAlert.EntityName = fmt.Sprintf("%s/%s", item.Labels["namespace"], item.Labels["pod"])
				currentAlert.EntityAddr = item.Labels["instance"]
			}

		}
		currentAlert.NetworkDomain = networkDomain
		currentAlert.Type = item.Labels["alerttype"]
		if item.Labels["severity"] == "critical" {
			currentAlert.Severity = 3
		} else if item.Labels["severity"] == "error" {
			currentAlert.Severity = 2
		} else if item.Labels["severity"] == "warning" {
			currentAlert.Severity = 1
		} else if item.Labels["severity"] == "info" {
			currentAlert.Severity = 0
		}
		currentAlert.OccurTime = item.StartsAt.UnixMilli()
		currentAlert.MergedKey = mergeKey
		currentAlert.IdentifyKey = identifyKey

		data = append(data, currentAlert)
	}

	sendToGlowworm(data)

	// 返回成功响应
	resp.WriteHeader(http.StatusOK)
	fmt.Fprint(resp, "SUCCESS")
}

// 发送数据
func sendToGlowworm(data []Alert) {

	for _, address := range notifyAddress {
		go send(address, data)
	}

}

func send(address string, data []Alert) {

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

		glog.V(3).Infof("Response: %s\n", string(body))
		resp.Body.Close()
	}
}
