package server

import (
	"encoding/json"
	"github.com/MalsonQu/webhook/utils/base"
	"github.com/MalsonQu/webhook/utils/helper"
	"github.com/MalsonQu/webhook/utils/request"
	"io/ioutil"
	"log"
	"net/http"
)

type Server struct{}

func (s *Server) Start() {
	log.Println("Web starting")
	// 处理 解析方法
	http.HandleFunc("/", s.process)
	// 启动服务器
	err := http.ListenAndServe(base.Conf.Server.Address+":"+base.Conf.Server.Port, nil)

	if err != nil {
		// 启动失败
		log.Fatalf("Field start Web server. error: %s\n", err.Error())
	}
	// 启动成功
	log.Printf("Success start web server")
}

// 请求处理函数
func (s *Server) process(_ http.ResponseWriter, r *http.Request) {
	log.Println("--------------------------------------")
	log.Println("New request")
	log.Println("Check header")

	// 判断 header 必要参数是否存在
	err := request.CheckHeaderRequestEmpty(r.Header)

	if err != nil {
		log.Printf("Not illegal Request. error: %s\n", err.Error())
		return
	}

	// 仅处理 Merge 请求
	if r.Header.Get("X-Coding-Event") != "push" {
		log.Printf("Not illegal Event %s\n", r.Header.Get("X-Coding-Event"))
		return
	}

	log.Println("Resolution body")
	// 解析 请求 body

	// 读取 body
	_body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("Field resolution Request body. error: %s\n", err.Error())
		return
	}

	defer func() {
		_ = r.Body.Close()
	}()

	// 解析 body 到json

	type repository struct {
		FullName string `json:"full_name"`
	}

	type jsonStruct struct {
		Ref        string     `json:"ref"`
		Repository repository `json:"repository"`
	}

	var _json jsonStruct

	// 解析 json
	err = json.Unmarshal(_body, &_json)

	if err != nil {
		log.Printf("Field unmarshal Request body to json. error: %s\n", err.Error())
		return
	}

	_projectIndex := -1

	log.Println("Match project")

	for index, _project := range base.Conf.Project {
		if _project.RefName == _json.Ref && _project.FullName == _json.Repository.FullName {
			_projectIndex = index
			break
		}
	}

	// 判断是否匹配到 客服
	if _projectIndex == -1 {
		log.Println("Field Match project")
		return
	}

	_project := base.Conf.Project[_projectIndex]

	log.Printf("Matched project %s:%s\n", _project.FullName, _project.Ref)

	log.Println("Check sign")
	// 验证 签名
	_pass := request.CheckSign(&_body, r.Header.Get(`X-Coding-Signature`)[5:], base.Conf.Server.PublicKey)

	if !_pass {
		log.Printf("Not illegal sign")
		return
	}

	log.Println("Start Pull")
	// 此处签名通过了可以进行拉取的操作了
	err = helper.HandlePull(_project.Path, _project.RemoteName, _project.Ref)
	if err != nil {
		log.Printf("Field pull project error: %s\n", err.Error())
		return
	}

	log.Println("Success pull project")
}
