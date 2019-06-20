# webhook

## 概述

用来处理 coding.net 的 webhook
使用 yaml 作为配置文件

## 使用

#### 下载依赖
```text
go get gopkg.in/yaml.v2
```

#### 编译
```
go build github.com/MalsonQu/webhook
```
#### 配置
复制配置文件`conf.template.yaml`到`conf.yaml`并按照实际项目编辑中的内容

#### 运行
```
./webhook
```

## Thanks
- go-yaml/yaml by @rogpeppe






