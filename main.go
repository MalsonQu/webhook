package main

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"net/http"
	"crypto/sha1"
	"crypto/hmac"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"net/smtp"
)

//type WebHook interface {
//	readConfig() ([]byte, error)
//}

type WebHook struct {
	//Conf map[interface{}]interface{}
	Conf Conf
}

type Conf struct {
	//GetterEmail []string `ymal:",flow"`
	Project []struct {
		Path      string
		User      string
		PublicKey string
		Full_name string
	}
	GetterEmail []string
	Port        string
	SenderEmail struct {
		Host     string
		Port     string
		User     string
		Password string
	}
}

type Json struct {
	Head_commit struct {
		Committer struct {
			Name     string
			Email    string
			Username string
		}
		Message string
	}
	Repository struct {
		Full_name string
	}
	Ref string
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

func (th *WebHook) Pull(path, user string) ([]byte, error) {

	var cmd *exec.Cmd

	// 如果没有定义用户类型
	if user == "" {
		cmd = exec.Command("git", "pull")
	} else {
		cmd = exec.Command("sudo", "-u", user, "git", "pull")
	}

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
		log.Printf("请求类型为 %s ,忽略请求\n" , r.Header.Get("X-Coding-Event"))
		return
	}

	//获取 请求body 信息 json 数据
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("读取请求信息出错 %v\n", err)
		return
	}

	// 解析json
	response, err := th.UnmarshalJson(body)

	if err != nil {
		log.Printf("读取JSON数据出错 %v\n", err)
		return
	}

	// 设置项目索引
	projectIndex := -1

	for index, value := range th.Conf.Project {
		if response.Repository.Full_name == value.Full_name {
			projectIndex = index
			break
		}
	}

	if projectIndex == -1 {
		log.Println("未找到匹配的项目配置")
		return
	}

	// 验证请求是否合法
	if !th.CheckMAC(&body, r.Header.Get(`X-Coding-Signature`)[5:], th.Conf.Project[projectIndex].PublicKey) {
		log.Println("服务端请求验证失败")
		return
	}
	// 拉取项目
	output, err := th.Pull(th.Conf.Project[projectIndex].Path, th.Conf.Project[projectIndex].User)

	if err != nil {
		log.Printf("项目<%v>拉取出错,错误信息将通过邮件发送", th.Conf.Project[projectIndex].Full_name)
		err = th.SendMail(th.Conf.SenderEmail.Host, th.Conf.SenderEmail.Port, th.Conf.SenderEmail.User, th.Conf.SenderEmail.Password, th.Conf.GetterEmail, "项目<"+th.Conf.Project[projectIndex].Full_name+">拉取出错!", "错误信息:\n---------------------------------------------------------------------------------\n"+string(output))
		if err != nil {
			log.Printf("邮件发送失败 \n%v\n", err)
			return
		} else {
			log.Println("错误信息已发送到邮箱中!")
			return
		}
	}

	log.Printf("项目 %v 更新成功", th.Conf.Project[projectIndex].Full_name)
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
	}

	// 启动
	err = w.StartServer()
	if err != nil {
		log.Fatalf("服务器启动失败 %v\n", err)
	}

}
