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
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
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

// 所有配置
var config *Config

// 配置文件所在的路径
var configPath string

var router *httprouter.Router

// 配置文件清单
// var configFiles []string

func main() {
	log.SetFlags(log.Lshortfile)

	router = httprouter.New()

	// 从运行命令中获得入口配置文件名
	configFileName := flag.String("config", "./mock/api.json", "Specify the entry configuration file")
	flag.Parse()

	// 获得配置文件的目录
	configPath = filepath.Dir(*configFileName)

	// 读取json配置文件
	err := loadConfig(*configFileName, &config)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// 启动配置文件监听
	// go watcher(configFileName)

	// 更新路由并重启服务
	err = updateRouters()
	if err != nil {
		log.Println(err.Error())
		return
	}

	// cors处理
	handler := cors.Default().Handler(router)

	// 启动服务
	fmt.Println("Listen on " + config.Listen)
	if err := http.ListenAndServe(config.Listen, handler); err != nil {
		log.Println(err.Error())
		return
	}
}

// 监视配置文件变更
func watcher(configFileName *string) {
	// 创建监视器
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

				// fmt.Println("Event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					// 读取json配置文件
					err = loadConfig(*configFileName, &config)
					if err != nil {
						log.Println(err.Error())
						break
					}
					// 更新路由
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
	// 清空配置文件清单
	// configFiles = []string{}

	// 将入口配置文件加入清单
	// configFiles = append(configFiles, filepath.Base(filename))

	// 读取配置文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	// 反序列化json
	err = json.Unmarshal(data, v)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// 遍历配置文件中的路由节点
	for k := range config.Routers {
		thisInclude := config.Routers[k].Include

		// 如果是外部json文件
		if thisInclude != "" {
			// 将包含的配置文件加入清单
			// configFiles = append(configFiles, this.Include)

			// 读取配置文件
			data, err := ioutil.ReadFile(configPath + "/" + thisInclude)
			if err != nil {
				log.Println(err.Error())
				return err
			}
			var route *Route
			// 反序列化json
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
	// 重置路由
	router = httprouter.New()

	// 404错误
	router.NotFound = func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(404)
		resp.Write([]byte(http.StatusText(404)))
	}
	// 405错误
	// router.HandleMethodNotAllowed = true
	// router.MethodNotAllowed = func(resp http.ResponseWriter, req *http.Request) {
	// 	resp.WriteHeader(405)
	// 	resp.Write([]byte(http.StatusText(405)))
	// }

	// 遍历配置文件中的路由节点
	for k := range config.Routers {
		thisMethod := config.Routers[k].Method
		thisPath := config.Routers[k].Path
		thisData := config.Routers[k].Data
		thisStatus := config.Routers[k].Status

		// // 预检测路由
		// _, _, exist := router.Lookup(thisMethod, thisPath)
		// if exist == true {
		// 	router.OPTIONS(thisPath, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		// 		// 允许来自所有域的请求
		// 		resp.Header().Set("Access-Control-Allow-Origin", "*")
		// 		resp.Header().Set("Access-Control-Allow-Credentials", "true")
		// 		resp.Header().Set("Access-Control-Allow-Methods", "*")
		// 		resp.Header().Set("Access-Control-Allow-Headers", "*")
		// 		resp.Header().Set("Access-Control-Expose-Headers", "*")
		// 		resp.WriteHeader(200)
		// 	})
		// }
		// 注册路由
		router.Handle(strings.ToUpper(thisMethod), thisPath, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
			// 允许来自所有域的请求
			// resp.Header().Set("Access-Control-Allow-Origin", "*")
			// resp.Header().Set("Access-Control-Allow-Credentials", "true")
			// resp.Header().Set("Access-Control-Allow-Methods", "*")
			// resp.Header().Set("Access-Control-Allow-Headers", "*")
			// resp.Header().Set("Access-Control-Expose-Headers", "*")

			// 打印请求
			fmt.Println()
			fmt.Println("Request Time: " + time.Now().Format(timeTpl))
			fmt.Println("Request Resource: [" + req.Method + "] " + req.RequestURI)
			fmt.Println("Request Headers: ")
			for k := range req.Header {
				fmt.Println(k, "=", req.Header.Get(k))
			}
			// 将data节点序列化成json
			result, err := json.Marshal(thisData)
			if err != nil {
				log.Println(err.Error())
				return
			}

			resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
			// 将数据响应给客户端
			resp.WriteHeader(thisStatus)
			resp.Write(result)
			fmt.Println("Response Status: " + strconv.Itoa(thisStatus))
			fmt.Println("Response Data: ")
			fmt.Println(string(result))
		})
		// 输出路由注册信息
		fmt.Println("Registered Routing: [" + strings.ToUpper(thisMethod) + "] " + thisPath)
	}
	fmt.Println("Update router success")
	return nil
}
