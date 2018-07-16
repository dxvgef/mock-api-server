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

	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// 配置文件结构
type Config struct {
	Listen  string   `json:"listen"`
	Routers []Router `json:"routers"`
}

// 路由结构
type Router struct {
	Include string `json:"include,omitempty"`
	Route
}

// 路由结构详细
type Route struct {
	Desc   string                 `json:"desc,omitempty"`
	Path   string                 `json:"path"`
	Method string                 `json:"method"`
	Status int                    `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

const timeTpl = "2006-01-02 15:04:05"

var config *Config

var configFiles []string

var configPath string

func main() {
	log.SetFlags(log.Lshortfile)

	//从运行命令中获得入口配置文件名
	configFileName := flag.String("config", "./mock/api.json", "Specify the entry configuration file")
	flag.Parse()

	//获得配置文件的目录
	configPath = filepath.Dir(*configFileName)

	//读取json配置文件
	err := loadConfig(*configFileName, &config)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//更新路由
	err = updateRouters()
	if err != nil {
		log.Println(err.Error())
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

// 监视配置文件变更
func watcher(configFileName *string) {
	//创建监视器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return
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
					err = loadConfig(*configFileName, &config)
					if err != nil {
						log.Println(err.Error())
						break
					}
					//更新路由
					err = updateRouters()
					if err != nil {
						log.Println(err.Error())
						break
					}
				}
			case err := <-watcher.Errors:
				log.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add(configPath)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

// 加载配置文件
func loadConfig(filename string, v interface{}) error {
	//fmt.Println("Loading configuration file")
	//读取配置文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	//反序列化json
	err = json.Unmarshal(data, v)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	//遍历配置文件中的路由节点
	for k := range config.Routers {
		this := config.Routers[k]

		//如果是外部json文件
		if this.Include != "" {
			//读取配置文件
			data, err := ioutil.ReadFile(configPath + "/" + this.Include)
			if err != nil {
				log.Println(err.Error())
				return err
			}
			var route *Route
			//反序列化json
			err = json.Unmarshal(data, &route)
			if err != nil {
				log.Println(err.Error())
				return err
			}
			config.Routers[k].Desc = route.Desc
			config.Routers[k].Path = route.Path
			config.Routers[k].Method = route.Method
			config.Routers[k].Status = route.Status
			config.Routers[k].Data = route.Data
		}
	}

	fmt.Println("Load configuration file success")
	return nil
}

// 更新所有路由规则
func updateRouters() error {
	//重置路由
	http.DefaultServeMux = http.NewServeMux()

	//存在根路径的路由
	var hasRootPath = false

	//遍历配置文件中的路由节点
	for k := range config.Routers {
		this := config.Routers[k]

		//fmt.Println("Registered Routing: [" + this.Method + "] " + this.Path)

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
			fmt.Println("Request Time: " + time.Now().Format(timeTpl))
			fmt.Println("Request Resource: [" + r.Method + "] " + r.RequestURI)
			fmt.Println("Request Headers: ")
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
				fmt.Println("Response Status: " + strconv.Itoa(this.Status))
				fmt.Println("Response Data: ")
				fmt.Println(string(result))
				return
			}
			//响应405错误
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("[" + r.Method + "] " + http.StatusText(http.StatusMethodNotAllowed)))
			fmt.Println("Response Status: ", http.StatusMethodNotAllowed)
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

	fmt.Println("Update router success")
	return nil
}
