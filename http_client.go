package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	// 代理服务器地址
	proxyURL, _ := url.Parse("http://192.168.3.221:8780")

	//设置Proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// 使用Transport创建Client
	client := &http.Client{Transport: transport}

	// 构建请求 （配置一个web服务监听127.0.0.1:8080，且与代理服务器在同一台机器上测试）
	req, _ := http.NewRequest("GET", "https://www.baidu.com", nil)
	resp, _ := client.Do(req)
	// 读取响应
	body, _ := ioutil.ReadAll(resp.Body)
	// 打印结果
	fmt.Println(string(body))
}
