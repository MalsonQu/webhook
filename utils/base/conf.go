package base

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type base struct {
	Server  server    `yaml:"server"`  // 服务器配置
	Project []project `yaml:"project"` // 项目配置
}

// 服务器配置
type server struct {
	Address   string `yaml:"address"`    // 监听地址
	Port      string `yaml:"port"`       // 监听端口
	PublicKey string `yaml:"public_key"` // 公钥
}

type project struct {
	Path       string `yaml:"path"`        // 项目所在服务器路径
	RemoteName string `yaml:"remote_name"` // 远端仓库名称		如 origin
	FullName   string `yaml:"full_name"`   // 项目全称			如 MalsonQu/CodingWebHook
	Ref        string `yaml:"ref"`         // 项目分支名称 		如 dev
	RefName    string `yaml:"ref_name"`    // coding 分支名称 	如 refs/heads/master
	User       string `yaml:"user"`        // 执行命令的用户	如 www
}

var Conf base

func init() {
	log.Println("init config")
	initConf()
	log.Println("success init config")
}

func initConf() {
	// 读取 配置文件
	fileContent, err := ioutil.ReadFile("./conf.yaml")

	if err != nil {
		log.Fatalf("Field open config file. error: %s\n", err.Error())
	}

	// 解析配置文件
	err = yaml.Unmarshal(fileContent, &Conf)

	if err != nil {
		log.Fatalf("Filed unmarshal config filr error: %s\n", err.Error())
	}
}
