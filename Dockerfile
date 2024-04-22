# Build the manager binary
FROM golang:alpine

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
        GOPROXY="https://goproxy.cn,direct"

# 移动到工作目录：/home/www/goWebBlog 这个目录 是你项目代码 放在linux上
# 这是我的代码跟目录
# 你们得修改成自己的
WORKDIR /app
COPY . .
RUN go mod download

# 将代码复制到容器中

# 将我们的代码编译成二进制可执行文件  可执行文件名为 app
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o app .
RUN chmod +x app

# 启动容器时运行的命令
CMD ["./app"]
