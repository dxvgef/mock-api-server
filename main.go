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

	"github.com/fsnotify/fsnotify"
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

var config *Config

func main() {
	log.SetFlags(log.Lshortfile)

	//从运行命令中获得配置文件名
	configFileName := flag.String("config", "api.json", "Specify the configuration file")
	flag.Parse()

	//读取json配置文件
	err := config.Load(*configFileName, &config)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//启动配置文件监听
	go watcher(configFileName)

	//启动服务
	fmt.Println("Listen on " + config.Listen)
	if err := http.ListenAndServe(config.Listen, nil); err != nil {
		log.Println(err.Error())
		return
	}
}

func watcher(configFileName *string) {
	//创建监视器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				//fmt.Println("Event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					//读取json配置文件
					err = config.Load(*configFileName, &config)
					if err != nil {
						fmt.Println(err.Error())
						return
					}
				}
			case err := <-watcher.Errors:
				fmt.Println("* Error:", err)
			}
		}
	}()

	err = watcher.Add(*configFileName)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func Handle404(handler http.Handler, handle404 func(w http.ResponseWriter, r *http.Request) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(http.StatusText(404)))
	})
}

// 加载json配置文件
func (obj *Config) Load(filename string, v interface{}) error {
	fmt.Println("* Loading configuration file: " + filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	//重置路由
	http.DefaultServeMux = http.NewServeMux()

	//存在根路径的路由
	var hasRootPath = false

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

			if this.Path == "/" {
				hasRootPath = true
				if r.RequestURI != "/" {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(http.StatusText(http.StatusNotFound)))
					return
				}
			}

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

	//如果不存在根路由
	if hasRootPath == false {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			//允许来自所有域的请求
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Expose-Headers", "*")

			if r.RequestURI != "/" {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(http.StatusText(http.StatusNotFound)))
				return
			}
		})
	}

	fmt.Println("* The configuration file was successfully loaded")
	return nil
}
