# tfjl-h5-fight
 塔防精灵H5对战服务器（本仓库仅供学习交流使用，不得用于非法用途，否则后果自负）
 
## 项目介绍
塔防精灵H5版本的对战服务器，后端使用golang，数据库使用MongoDB，使用Websocket通信。

## 本地搭建
1. 安装MongoDB，创建数据库`tfjl`，配置好数据库的账号密码，然后在`db文件夹`下的`dbconnection.go`文件中的InitDatabase方法中配置好数据库账号密码
2. 将本仓库下载到本地，打开仓库所在目录的命令行，执行`go mod tidy`下载相关依赖，执行`go run main.go`，启动对战服务器的后端服务
3. 启动本对战服务前，需要先启动主逻辑服务器`tfjl-h5`，参考[tfjl-h5](https://github.com/Xiaeer/tfjl-h5)
