package main

import "sync"

// 建造型模式

// 单例模式是用来控制类型实例的数量
// 当需要确保一个类型只有一个实例时，就需要使用单例模式。

// 单例模式还会提供一个访问该实例的全局端口，
// 一般都会命名个 `GetInstance` 之类的函数用作实例访问的端口

// 又因为在什么时间创建出实例，单例模式又可以分裂出 `饿汉模式` 和 `懒汉模式`
// `饿汉模式` 适用于在程序早期初始化时创建已经确定需要加载的类型实例，比如项目的数据库实例
// `懒汉模式` 其实就是延迟加载的模式，适合程序执行过程中条件成立才创建加载的类型实例

// 懒汉模式 - 饿汉模式可以使用 Go 的 init 方法实现
type singleton struct{}

var instance *singleton
var once sync.Once

func GetInstance() *singleton {
	once.Do(func() {
		instance = &singleton{}
	})

	return instance
}

func main() {

}
