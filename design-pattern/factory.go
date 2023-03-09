package main

import (
	"fmt"
)

// 工厂模式 -- 建造型模式

// 1. 简单工厂
// 简单工厂：是简单工厂模式的核心，负责实现创建所有实例的内部逻辑。
// 			工厂类的创建产品类的方法可以被外界直接调用，创建所需的产品对象。
// 抽象产品：是简单工厂创建的所有对象的抽象父类/接口，负责描述所有实例的行为。
// 具体产品：是简单工厂模式的创建目标

// 简单工厂的优点是，简单，
// 缺点嘛，如果具体产品扩产，就必须修改工厂内部，增加Case，
// 一旦产品过多就会导致简单工厂过于臃肿，为了解决这个问题，才有了下一级别的工厂模式--工厂方法。

// 简单工厂 ------------------------------------- start
type Printer interface {
	Print(string) string
}

func NewPrinter(lang string) Printer {
	switch lang {
	case "cn":
		return new(CNPrinter)
	case "en":
		return new(ENPrinter)
	default:
		return new(CNPrinter)
	}
}

type CNPrinter struct{}

func (cn *CNPrinter) Print(name string) string {
	return fmt.Sprintf("你好，%s", name)
}

type ENPrinter struct{}

func (en *ENPrinter) Print(name string) string {
	return fmt.Sprintf("Hello, %s", name)
}

// 简单工厂 ------------------------------------- end

// 2. 工厂方法
// 工厂方法模式（Factory Method Pattern）又叫作多态性工厂模式，
// 指的是定义一个创建对象的接口，但由实现这个接口的工厂类来决定实例化哪个产品类，
// 工厂方法把类的实例化推迟到子类中进行。

// 在工厂方法模式中，不再由单一的工厂类生产产品，而是由工厂类的子类实现具体产品的创建。
// 因此，当增加一个产品时，只需增加一个相应的工厂类的子类, 以解决简单工厂生产太多产品时导致其内部代码臃肿（switch … case分支过多）的问题。

// 工厂方法 ------------------------------------- start
// 工厂接口由具体工厂类实现
type OperatorFactory interface {
	Create() MathOperator
}

// 实际产品实现的接口
type MathOperator interface {
	SetOperatorA(int)
	SetOperatorB(int)
	ComputeResult() int
}

// go 不支持继承，封装公共方法
type BaseOperator struct {
	operatorA, operatorB int
}

func (o *BaseOperator) SetOperatorA(num int) {
	o.operatorA = num
}

func (o *BaseOperator) SetOperatorB(num int) {
	o.operatorB = num
}

// plus 工厂类
type PlusOperatorFactory struct{}

// new plus 工厂类
func (pf *PlusOperatorFactory) Create() MathOperator {
	return &PlusOperator{
		BaseOperator: &BaseOperator{},
	}
}

// pluc 实际产品实现
type PlusOperator struct {
	*BaseOperator
}

func (p *PlusOperator) ComputeResult() int {
	return p.operatorA + p.operatorB
}

// multi 工厂类
type MultiOperatorFactory struct{}

// new multi 工厂类
func (mf *MultiOperatorFactory) Create() MathOperator {
	return &MultiOperator{
		BaseOperator: &BaseOperator{},
	}
}

// multi 实际产品实现
type MultiOperator struct {
	*BaseOperator
}

func (m *MultiOperator) ComputeResult() int {
	return m.operatorA * m.operatorB
}

// 工厂方法模式的优点
// 灵活性增强，对于新产品的创建，只需多写一个相应的工厂类。
// 典型的解耦框架。高层模块只需要知道产品的抽象类，无须关心其他实现类，满足迪米特法则、依赖倒置原则和里氏替换原则。

// 工厂方法模式的缺点
// 类的个数容易过多，增加复杂度。
// 增加了系统的抽象性和理解难度。
// 只能生产一种产品，此弊端可使用抽象工厂模式解决。

// 工厂方法 ------------------------------------- end

// 3. 抽象工厂
// 抽象工厂模式：用于创建一系列相关的或者相互依赖的对象

// 抽象工厂 ------------------------------------- start
type AbstracFactory interface {
	CreateTelevision() ITelevision
	CreateAirConditioner() IAirConditioner
}

type ITelevision interface {
	Watch()
}

type IAirConditioner interface {
	SetTemperature(int)
}

type HuaweiFactory struct{}

func (hf *HuaweiFactory) CreateTelevision() ITelevision {
	return &Huawei{}
}

func (hf *HuaweiFactory) CreateAirConditioner() IAirConditioner {
	return &HuaweiAirConditioner{}
}

type Huawei struct{}

func (h *Huawei) Watch() {
	fmt.Println("Watch Huawei TV")
}

type HuaweiAirConditioner struct{}

func (h *HuaweiAirConditioner) SetTemperature(temp int) {
	fmt.Printf("HuaWei AirConditioner set temperature to %d ℃\n", temp)
}

type MiFactory struct{}

func (m *MiFactory) CreateTelevision() ITelevision {
	return &Mi{}
}

func (m *MiFactory) CreateAirConditioner() IAirConditioner {
	return &MiAirConditioner{}
}

type Mi struct{}

func (m *Mi) Watch() {
	fmt.Println("Watch Mi TV")
}

type MiAirConditioner struct{}

func (m *MiAirConditioner) SetTemperature(temp int) {
	fmt.Printf("Mi AirConditioner set temperature to %d ℃\n", temp)
}

// 抽象工厂 ------------------------------------- end

func main() {
	// 抽象工厂
	var factory AbstracFactory
	var tv ITelevision
	var air IAirConditioner

	factory = &HuaweiFactory{}
	tv = factory.CreateTelevision()
	air = factory.CreateAirConditioner()
	tv.Watch()
	air.SetTemperature(25)

	factory = &MiFactory{}
	tv = factory.CreateTelevision()
	air = factory.CreateAirConditioner()
	tv.Watch()
	air.SetTemperature(26)

	// 工厂方法
	// var factory OperatorFactory
	// var mathOp MathOperator
	// factory = &PlusOperatorFactory{}
	// mathOp = factory.Create()
	// mathOp.SetOperatorB(3)
	// mathOp.SetOperatorA(2)
	// fmt.Printf("Plus operation reuslt: %d\n", mathOp.ComputeResult())

	// factory = &MultiOperatorFactory{}
	// mathOp = factory.Create()
	// mathOp.SetOperatorB(3)
	// mathOp.SetOperatorA(2)
	// fmt.Printf("Multiple operation reuslt: %d\n", mathOp.ComputeResult())

	// 简单工厂 ----------------
	// printer := NewPrinter("cn")
	// fmt.Println(printer.Print("test"))
}
