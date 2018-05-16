#Mock API Server
用Go语言开发的模拟API Server，使Web前端开发过程中自行模拟后端服务响应的数据。

## 使用方法：
在项目中创建一个json配置文件，内容如下：
```JSON
{
    "listen": "127.0.0.1:3000",
    "routers": [
        {
            "path": "/",
            "method": "post",
            "data": {
                "error": "ok",
                "id": 123
            }
        }
    ]
}
```
`listen` 表示服务要监听的IP及端口

`routers` 表示要注册的路由

`path` 表示允许HTTP请求的路径

`method` 表示允许HTTP请求的方法

`data` 表示要响应给客户端的数据，该节点可以写成任意结构的JSON数据

将`mock-api-server`的可执行文件拷贝到系统目录中：
- Windows的路径为`C:\Windows`
- Linux、Mac OS的路径为`/usr/local/bin`

启动终端，进入配置文件所在的路径，运行`mock-api-server`命令启动服务。

Mock API Server默认会在命令所在的路径下查找`api.json`文件做为配置文件，也可以通过`-config=`参数指定配置文件，例如：

`mock-api-server -config=test.json`

`mock-api-server -config=~/test.json`