package app

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestName(t *testing.T) {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			// 打印接收到的数据
			// 读取请求体中的数据
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			// 解析 JSON 数据到结构体
			//var myData Lightning
			//err = json.Unmarshal(body, &myData)
			//if err != nil {
			//	log.Println("Failed to parse JSON data:", err)
			//	http.Error(w, "Failed to parse JSON data", http.StatusBadRequest)
			//	return
			//}
			//
			//// 打印接收到的数据
			//fmt.Printf("Received data:%+v\n", myData)

			fmt.Println(string(body))

			// 响应请求
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Data received successfully"))
		})

		if err := http.ListenAndServe(":9999", nil); err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
	}()

	ch := make(chan struct{})

	<-ch
}
