### 命令行参数

- port : conversions服务端口，默认监听8080
- address : 萤火虫接收数据的地址
- alertId : 萤火虫告警id
- alertKey : 萤火虫告警key

将 alertmanager 的数据以 POST 方式发送到 conversions 服务的接口即 http://localhost:8080/alert ,conversions 服务会将数据转换成萤火虫所规定的数据格式并发送。