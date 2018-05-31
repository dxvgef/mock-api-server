package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Config 配置文件结构
type Config struct {
	Listen  string   `json:"listen"`
	Routers []Router `json:"routers"`
}

// Router 配置文件中的路由结构
type Router struct {
	Path   string                 `json:"path"`
	Method string                 `json:"method"`
	Status int                    `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

const timeTpl = "2006-01-02 15:04:05"

func main() {
	//log.SetFlags(log.Lshortfile)

	//从运行命令中获得配置文件名
	configFileName := flag.String("config", "api.json", "Specify the configuration file")
	flag.Parse()

	//读取json配置文件
	var config *Config
	err := config.Load(*configFileName, &config)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//遍历配置文件中的路由节点
	for k := range config.Routers {
		this := config.Routers[k]
		//fmt.Println("注册路由 [" + this.Method + "] " + this.Path)
		//注册路由
		http.HandleFunc(this.Path, func(w http.ResponseWriter, r *http.Request) {
			//允许来自所有域的请求
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Expose-Headers", "*")
			//如果是预检查的options请求
			if r.Method == "OPTIONS" {
				//响应202请求
				w.WriteHeader(http.StatusAccepted)
				return
			}
			// 打印请求
			fmt.Println()
			fmt.Println("* Request Time: " + time.Now().Format(timeTpl))
			fmt.Println("* Request Resource: [" + r.Method + "] " + r.RequestURI)
			fmt.Println("* Request Headers: ")
			for k := range r.Header {
				fmt.Println(k, "=", r.Header.Get(k))
			}
			// 判断HTTP方法是否匹配
			if r.Method == strings.ToUpper(this.Method) {
				//将data节点序列化成json
				result, err := json.Marshal(this.Data)
				if err != nil {
					log.Println(err.Error())
					return
				}
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				//将数据响应给客户端
				w.WriteHeader(this.Status)
				w.Write(result)
				fmt.Println("* Response Status: " + strconv.Itoa(this.Status))
				fmt.Println("* Response Data: ")
				fmt.Println(string(result))
				return
			}
			//响应405错误
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("[" + r.Method + "] " + http.StatusText(http.StatusMethodNotAllowed)))
			fmt.Println("* Response Status: ", http.StatusMethodNotAllowed)
		})
	}

	//启动服务
	fmt.Println("Listen on " + config.Listen)
	if err := http.ListenAndServe(config.Listen, nil); err != nil {
		log.Println(err.Error())
		return
	}
}

// 加载json配置文件
func (obj *Config) Load(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}
