package main

import "fmt"

// @Author dx
// @Date 2025-08-04 21:44:00
// @Desc

// Handler 处理器接口
type Handler interface {
	SetNext(handler Handler) Handler
	Handle(request interface{}) error
}

// BaseHandler 基础处理器，实现了链式调用的基本逻辑
type BaseHandler struct {
	next Handler
}

func (h *BaseHandler) SetNext(handler Handler) Handler {
	h.next = handler
	return handler
}

func (h *BaseHandler) Handle(request interface{}) error {
	if h.next != nil {
		return h.next.Handle(request)
	}
	return nil
}

// 具体的请求类型
type ApprovalRequest struct {
	Amount int
	Type   string
}

// TeamLeaderHandler 组长处理器（处理1000元以下）
type TeamLeaderHandler struct {
	BaseHandler
}

func (h *TeamLeaderHandler) Handle(request interface{}) error {
	req, ok := request.(*ApprovalRequest)
	if !ok {
		return fmt.Errorf("invalid request type")
	}

	if req.Amount <= 1000 {
		fmt.Printf("组长审批通过：金额 %d 元\n", req.Amount)
		return nil
	}

	fmt.Printf("组长：金额 %d 元超出权限，转交上级\n", req.Amount)
	return h.BaseHandler.Handle(request)
}

// ManagerHandler 经理处理器（处理5000元以下）
type ManagerHandler struct {
	BaseHandler
}

func (h *ManagerHandler) Handle(request interface{}) error {
	req, ok := request.(*ApprovalRequest)
	if !ok {
		return fmt.Errorf("invalid request type")
	}

	if req.Amount <= 5000 {
		fmt.Printf("经理审批通过：金额 %d 元\n", req.Amount)
		return nil
	}

	fmt.Printf("经理：金额 %d 元超出权限，转交上级\n", req.Amount)
	return h.BaseHandler.Handle(request)
}

// DirectorHandler 总监处理器（处理10000元以下）
type DirectorHandler struct {
	BaseHandler
}

func (h *DirectorHandler) Handle(request interface{}) error {
	req, ok := request.(*ApprovalRequest)
	if !ok {
		return fmt.Errorf("invalid request type")
	}

	if req.Amount <= 10000 {
		fmt.Printf("总监审批通过：金额 %d 元\n", req.Amount)
		return nil
	}

	fmt.Printf("总监：金额 %d 元超出权限，需要开会讨论\n", req.Amount)
	return fmt.Errorf("金额过大，需要开会讨论")
}

// 使用示例
func main() {
	// 创建处理器
	teamLeader := &TeamLeaderHandler{}
	manager := &ManagerHandler{}
	director := &DirectorHandler{}

	// 设置责任链
	teamLeader.SetNext(manager).SetNext(director)

	// 测试不同金额的审批请求
	requests := []*ApprovalRequest{
		{Amount: 500, Type: "日常开支"},
		{Amount: 3000, Type: "设备采购"},
		{Amount: 8000, Type: "项目经费"},
		{Amount: 15000, Type: "大额采购"},
	}

	for _, req := range requests {
		fmt.Printf("\n处理请求：%s，金额：%d 元\n", req.Type, req.Amount)
		if err := teamLeader.Handle(req); err != nil {
			fmt.Printf("错误：%v\n", err)
		}
	}
}
