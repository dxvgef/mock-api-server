# Mock API Server
Go语言开发的模拟API服务端，Web前端开发可以用它模拟后端服务响应的JSON数据。

## 下载
[进入Release](https://github.com/dxvgef/mock-api-server/releases) 下载编译好的可执行文件

## 使用方法：
#### 一、在项目中创建一个json配置文件，内容如下：
```JSON
{
    "listen": "127.0.0.1:3000",
    "routers": [
        {
            "desc": "这里可以写说明注释",
            "path": "/",
            "method": "post",
            "status": 200,
            "data": {
                "error": "",
                "id": 123
            }
        },
        {
            "desc": "用include指令加载了一个与入口配置文件同目录下的data.json文件",
            "include": "data.json"
        }
    ]
}
```
`listen` 必要，表示服务要监听的IP及端口

`routers` 必要，表示要注册的路由

`desc` 非必要，表示说明注释

`path` 必要，表示允许HTTP请求的路径

`method` 必要，表示允许HTTP请求的方法

`status` 必要，表示响应给客户端的HTTP状态码

`data` 非必要，表示要响应给客户端的数据，该节点可以写成任意结构的JSON数据

`include` 非必要，表示加载同目录下的另一个JSON配置文件

#### 二、将`mock-api-server`的可执行文件拷贝到系统目录中：
- Windows的路径为`C:\Windows`
- Linux、Mac OS的路径为`/usr/local/bin`

#### 三、启动终端，进入配置文件所在的路径，运行`mock-api-server`命令启动服务。

Mock API Server默认会在命令所在的路径下查找`/mock/api.json`文件做为入口配置文件，也可以通过`-config=`参数指定入口配置文件，例如：

`mock-api-server -config=entry.json`

`mock-api-server -config=./mock/entry.json`

入口配置文件所在的路径内，如果有文件内容被改写，将会触发路由规则的实时更新。