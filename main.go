package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Config 配置文件结构
type Config struct {
	Listen  string   `json:"listen"`
	Routers []Router `json:"routers"`
}

// Router 配置文件中的路由结构
type Router struct {
	Path   string
	Method string
	Data   map[string]interface{}
}

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
	for _, v := range config.Routers {
		//注册路由
		http.HandleFunc(v.Path, func(w http.ResponseWriter, r *http.Request) {
			//允许来自所有域的请求
			w.Header().Set("Access-Control-Allow-Origin", "*")
			// 打印请求
			log.Println(r.Method, r.RequestURI, r.Header)
			// 判断HTTP方法是否匹配
			if r.Method == strings.ToUpper(v.Method) {
				//将data节点序列化成json
				result, err := json.Marshal(v.Data)
				if err != nil {
					log.Println(err.Error())
					return
				}
				//将数据响应给客户端
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write(result)
				return
			}
			//响应405错误
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("[" + r.Method + "] " + http.StatusText(http.StatusMethodNotAllowed)))
		})
	}

	//启动服务
	log.Println("Listen on " + config.Listen)
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
