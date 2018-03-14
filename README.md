# webhook

## 概述

用来处理 coding.net 的 webhook
使用 yaml 作为配置文件

## 使用

#### 编译
```
go build github.com/MalsonQu/webhook
```
#### 配置
复制配置文件`conf.yaml.template`到`conf.yaml`并按照实际项目编辑中的内容

```yaml
port: # 服务开启的端口号 eg:8080
project:
  -
    path: # git 项目所在的路径 eg:~/go/src/github.com/MalsonQu/webhook
    user: # 使用哪个用户来执行git pull (如果指定此项则必须使用root账号来运行webhook) eg:www-data
    publickey: # 公钥 eg:123123
    full_name: # 项目名称 eg:MalsonQu/webhook
getteremail: # 邮件接收者
  - #eg:quqingyu@live.cn
senderemail: # 邮件发送者
  host: # 主机名 eg:127.0.0.1
  port: # 端口号 eg:25
  user: # 用户 abc@def.com
  password: # 密码 asdasd
```

#### 运行
```
./webhook
```
#### 持久化运行
- 自建docker，不过太麻烦了 =(
- 使用 screen 
    ```text
    // 安装 screen
    root@Malson:~# apt-get install -y screen
    // 新建一个名称为 webhook 的 窗口
    root@Malson:~# screen -S webhook
    // 执行 webhook
    root@Malson:~# ./webhook 
    2018/03/14 08:55:36 开始初始化项目
    2018/03/14 08:55:36 开始初始化配置文件
    2018/03/14 08:55:36 启动web服务中
    2018/03/14 08:55:36 开始监听 0.0.0.0:8080
    2018/03/14 08:55:36 服务运行中...
    // 后台运行 screen ctrl + a + d
  
    // 查看运行中的 screen
    root@Malson:~# screen -ls
    There is a screen on:
            18263.webhook   (03/13/2018 04:58:12 PM)        (Detached)
    1 Socket in /run/screen/S-root.
    // 重新连接 screen
    root@Malson:~# screen -r webhook
    2018/03/14 08:55:36 开始初始化项目                                                                                                                                  
    2018/03/14 08:55:36 开始初始化配置文件                                                                                                                              
    2018/03/14 08:55:36 启动web服务中                                                                                                                                   
    2018/03/14 08:55:36 开始监听 0.0.0.0:8080                                                                                                                           
    2018/03/14 08:55:36 服务运行中...  
    ```

## Thanks
- go-yaml/yaml by @rogpeppe






