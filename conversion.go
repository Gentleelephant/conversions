package conversions

import (
	"conversions/template"
	"conversions/utils"
	"fmt"
	_ "k8s.io/api/core/v1"
	 "log"
	"net/http"
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

// 处理接收到的Alertmanager告警
func alertHandler(w http.ResponseWriter, r *http.Request) {

		// Parse alerts sent through Alertmanager webhook, more detail please refer to
	// https://github.com/prometheus/alertmanager/blob/master/template/template.go#L231
	data := template.Data{}
	if err := utils.JsonDecode(r.Body, &data); err != nil {
		log.Printf(err.Error())
		return
	}
	// 读取请求体
	 sendLightning := make([]Lightning, len(data.Alerts))
	 fmt.println(data)
     fmt.Println(data.tostring())
	for i, alert := range data.Alerts {
		sendLightning[i].alertName = alert.Labels["alertname"]
		alert.ID = utils.Hash(alert)

	}

	// 发送给萤火虫
	//sendToFirefly(sendLightning)

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Alert received and forwarded to Firefly")
}

// 发送数据给萤火虫
//func sendToFirefly(data map[string]interface{}) {
//	// 实现发送逻辑，将数据发送给萤火虫
//	// 例如使用HTTP POST请求发送数据给萤火虫的API
//}

func main() {
	// 注册处理程序
	http.HandleFunc("/alert", alertHandler)

	// 启动HTTP服务器
	log.Fatal(http.ListenAndServe(":8080", nil))
}