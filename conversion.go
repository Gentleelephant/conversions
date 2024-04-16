package conversions


import (
	"encoding/json"
	"fmt"
	_ "k8s.io/api/core/v1"
	_ "log"
	"net/http"
)

// 定义Alertmanager告警的结构
type AlertmanagerAlert struct {
	Status string `json:"status"`
	Labels struct {
		Alertname string `json:"alertname"`
		Severity  string `json:"severity"`
		// 其他告警标签
	} `json:"labels"`
	Annotations struct {
		Summary string `json:"summary"`
		// 其他注释
	} `json:"annotations"`
}

// 处理接收到的Alertmanager告警
func alertHandler(w http.ResponseWriter, r *http.Request) {

		// Parse alerts sent through Alertmanager webhook, more detail please refer to
	// https://github.com/prometheus/alertmanager/blob/master/template/template.go#L231
	data := template.Data{}
	if err := utils.JsonDecode(r.Body, &data); err != nil {
		h.handle(w, &response{http.StatusBadRequest, err.Error()})
		return
	}
	// 读取请求体
	var alert AlertmanagerAlert
	err := json.NewDecoder(r.Body).Decode(&alert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 转换成萤火虫的格式
	fireflyAlert := map[string]interface{}{
		"status":       alert.Status,
		"alertname":    alert.Labels.Alertname,
		"severity":     alert.Labels.Severity,
		"summary":      alert.Annotations.Summary,
		// 其他字段转换
	}

	// 发送给萤火虫
	sendToFirefly(fireflyAlert)

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Alert received and forwarded to Firefly")
}

// 发送数据给萤火虫
func sendToFirefly(data map[string]interface{}) {
	// 实现发送逻辑，将数据发送给萤火虫
	// 例如使用HTTP POST请求发送数据给萤火虫的API
}

func main() {
	// 注册处理程序
	http.HandleFunc("/alert", alertHandler)
 event.
	// 启动HTTP服务器
	log.Fatal(http.ListenAndServe(":8080", nil))
}