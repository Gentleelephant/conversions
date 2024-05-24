### 部署
参数信息
- port : conversions服务端口，默认监听8080
- address : 萤火虫接收数据的地址
- alertId : 萤火虫告警id
- alertKey : 萤火虫告警key
- applyType : 是否是预置告警，custom:自定义告警，preset:预置告警。默认custom
```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: conversions
  namespace: kubesphere-monitoring-system
  labels:
    app: conversions
spec:
  replicas: 1
  selector:
    matchLabels:
      app: conversions
  template:
    metadata:
      labels:
        app: conversions
    spec:
      containers:
        - name: container-5kese9
          image: 'kubesphere/conversions:v0.5'
          args:
            - >-
              --address=http://10.27.0.33:1980/thirdalert/v1/firefly/thirdpartyAlert/process
            - '--alertId=0000336001'
            - '--alertKey=RKKSUVtxKXUPLJEbt02PFQ=='
            - '--port=8080'
          ports:
            - name: http-0
              containerPort: 8080
              protocol: TCP
          resources:
            limits:
              cpu: 100m
              memory: 50Mi
            requests:
              cpu: 100m
              memory: 50Mi
          imagePullPolicy: IfNotPresent
      restartPolicy: Always
      terminationGracePeriodSeconds: 30
      securityContext: {}
      schedulerName: default-scheduler

---
kind: Service
apiVersion: v1
metadata:
  name: conversions-svc
  namespace: kubesphere-monitoring-system
  labels:
    app: conversions
spec:
  ports:
    - name: http
      port: 8080
      targetPort: 8080
      protocol: TCP
  selector:
    app: conversions
  type: ClusterIP
```
修改alertmanager的配置，修改 kubesphere-monitoring-system namespace下的 alertmanager-main secret 配置
```text
"receivers":
- "name": "conversions"
  "webhook_configs":
  - "url": "http://conversions-svc.kubesphere-monitoring-system.svc:8080/alert"
  
"route":
  "routes":
  - "match_re":
      "alerttype": ".*"
    "receiver": "conversions"
  	"continue": true
```