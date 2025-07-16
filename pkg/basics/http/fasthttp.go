package main

import (
	"fmt"
	"github.com/valyala/fasthttp"

	"log"
)

func GetMQInfos() {
	getUrl := fmt.Sprintf("https://www.baidu.com")
	log.Println(fmt.Sprintf("GetMQInfos request url (%+v)", getUrl))
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(getUrl)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := fasthttp.Do(req, resp); err != nil {
		log.Fatalf(fmt.Sprintf("GetMQInfos err %+v", err))
	}
	content := resp.Body()
	log.Println(fmt.Sprintf("GetMQInfos response content %+v", string(content)))
}

func main() {
	GetMQInfos()
}
