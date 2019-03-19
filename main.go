package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os/exec"
	"strings"
)

//type WebHook interface {
//	readConfig() ([]byte, error)
//}

type WebHook struct {
	//Conf map[interface{}]interface{}
	Conf Conf
}

type project struct {
	Path      string `yaml:"path"`
	PublicKey string `yaml:"public_key"`
	FullName  string `yaml:"full_name"`
	Ref       string `yaml:"ref"`
}

type senderEmail struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Conf struct {
	Project         []project   `yaml:"project"`
	Port            string      `yaml:"port"`
	TurnOnEmailSend bool        `yaml:"turn_on_email_send"`
	GetterEmail     []string    `yaml:"getter_email"`
	SenderEmail     senderEmail `yaml:"sender_email"`
}

type committer struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type headCommit struct {
	Committer committer `json:"committer"`
	Message   string    `json:"message"`
}

type repository struct {
	FullName string `json:"full_name"`
}

type Json struct {
	Ref        string     `json:"ref"`
	HeadCommit headCommit `json:"head_commit"`
	Repository repository `json:"repository"`
}

func (th *WebHook) init() error {
	log.Println("开始初始化项目")
	return th.InitConf()

}

func (th *WebHook) InitConf() error {
	log.Println("开始初始化配置文件")

	fileContent, err := ioutil.ReadFile("./conf.yaml")

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(fileContent, &th.Conf)

	if err != nil {
		return err
	}

	return nil
}

func (th *WebHook) Pull(path string) ([]byte, error) {

	var cmd *exec.Cmd

	// 如果没有定义用户类型
	cmd = exec.Command("git", "pull")

	// 设置运行路径
	cmd.Dir = path
	// 返回运行结果
	return cmd.CombinedOutput()
}

func (th *WebHook) SendMail(host, port, user, password string, to []string, subject, content string) error {
	// 用户验证数据
	auth := smtp.PlainAuth("", user, password, host)
	// 内容格式
	contentType := `Content-Type: text/plain; charset=utf-8`
	// 构造接受者邮箱
	sendTo := strings.Join(to, ";")
	// 编写消息
	msg := []string{
		"To: " + sendTo,
		"Form: " + user + "<" + user + ">",
		"Subject: " + subject,
		contentType,
		"",
		content,
	}
	// 发送邮件
	return smtp.SendMail(host+":"+port, auth, user, to, []byte(strings.Join(msg, "\r\n")))

}

// 处理请求
func (th *WebHook) Process(w http.ResponseWriter, r *http.Request) {
	log.Println("收到服务器请求")

	if r.Header.Get("X-Coding-Event") == "" || r.Header.Get("X-Coding-Signature") == "" || r.Header.Get("X-Coding-Delivery") == "" {
		log.Println("非法请求")
		return
	}

	if r.Header.Get("X-Coding-Event") != "push" {
		log.Printf("请求类型为 %s ,忽略请求\n", r.Header.Get("X-Coding-Event"))
		return
	}

	//获取 请求body 信息 json 数据
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("读取请求信息出错 %v\n", err)
		return
	}

	defer r.Body.Close()

	// 解析json
	response, err := th.UnmarshalJson(body)

	if err != nil {
		log.Printf("读取JSON数据出错 %v\n", err)
		return
	}

	// 设置项目索引
	projectIndex := -1

	for index, value := range th.Conf.Project {
		if response.Repository.FullName == value.FullName && response.Ref == value.Ref {
			projectIndex = index
			break
		}
	}

	if projectIndex == -1 {
		log.Println("未找到匹配的项目配置")
		return
	}

	// 项目配置
	projectConf := th.Conf.Project[projectIndex]

	// 验证请求是否合法
	if !th.CheckMAC(&body, r.Header.Get(`X-Coding-Signature`)[5:], projectConf.PublicKey) {
		log.Println("服务端请求验证失败")
		return
	}
	// 拉取项目
	output, err := th.Pull(projectConf.Path)

	if err != nil {
		log.Printf("项目<%v>拉取出错,错误信息:\n%s", projectConf.FullName+"-"+projectConf.Ref, string(output))

		// 判断是否发送邮件
		if th.Conf.TurnOnEmailSend {
			log.Printf("项目<%v>拉取出错,错误信息将通过邮件发送", projectConf.FullName+"-"+projectConf.Ref)
			err = th.SendMail(th.Conf.SenderEmail.Host, th.Conf.SenderEmail.Port, th.Conf.SenderEmail.User, th.Conf.SenderEmail.Password, th.Conf.GetterEmail, "项目<"+projectConf.FullName+"-"+projectConf.Ref+">拉取出错!", "错误信息:\n---------------------------------------------------------------------------------\n"+string(output))
			if err != nil {
				log.Printf("邮件发送失败 \n%v\n", err)
				return
			} else {
				log.Println("错误信息已发送到邮箱中!")
				return
			}
		}
		return
	}

	log.Printf("项目 %v 更新成功", projectConf.FullName)
	return
}

func (th *WebHook) UnmarshalJson(jsonByte []byte) (Json, error) {
	var jsonContent Json

	err := json.Unmarshal(jsonByte, &jsonContent)

	if err != nil {
		return jsonContent, err
	}
	return jsonContent, nil
}

func (th *WebHook) CheckMAC(body *[]byte, bodyMAC, key string) bool {

	mac := hmac.New(sha1.New, []byte(key))
	mac.Write(*body)
	expectedMAC := mac.Sum(nil)

	return hmac.Equal([]byte(bodyMAC), []byte(fmt.Sprintf("%x", expectedMAC)))
}

func (th *WebHook) StartServer() error {
	log.Println("启动web服务中")
	http.HandleFunc("/", th.Process)
	log.Printf("开始监听 0.0.0.0:%s\n", th.Conf.Port)
	log.Println("服务运行中...")
	return http.ListenAndServe(":"+th.Conf.Port, nil)
}

func main() {

	var err error
	w := WebHook{}

	// 初始化
	err = w.init()

	if err != nil {
		log.Printf("系统初始化失败 %v\n", err)
		return
	}

	// 启动
	err = w.StartServer()
	if err != nil {
		log.Fatalf("服务器启动失败 %v\n", err)
		return
	}

}
