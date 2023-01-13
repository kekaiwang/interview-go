package main

import "fmt"

// design interface
type PayBehavior interface {
	OrderPay(*PayCtx)
}

type Wxpay struct{}

func (w *Wxpay) OrderPay(payCtx *PayCtx) {
	fmt.Println("Wxpay.....")
}

type ThirdPay struct{}

func (t *ThirdPay) OrderPay(payCtx *PayCtx) {
	fmt.Println("ThirdPay.....")
}

type PayCtx struct {
	payBehavior PayBehavior
	payParams   map[string]interface{}
}

func NewPayCtx(p PayBehavior) *PayCtx {
	params := map[string]interface{}{
		"appID":   "WX123123",
		"OrderID": 123789,
	}

	return &PayCtx{
		payBehavior: p,
		payParams:   params,
	}
}

func (px *PayCtx) setBehavior(p PayBehavior) {
	px.payBehavior = p
}

func (px *PayCtx) Pay() {
	px.payBehavior.OrderPay(px)
}

func main() {
	wx := &Wxpay{}
	px := NewPayCtx(wx)
	px.Pay()

	third := &ThirdPay{}
	px.setBehavior(third)
	px.payBehavior.OrderPay(px)
}
