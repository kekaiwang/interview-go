package main

import "fmt"

type Subject interface {
	Subsribe(Observer)
	Notify(string)
}

type Observer interface {
	Update(string)
}

type SubjectImpl struct {
	Obersers []Observer
}

func (sub *SubjectImpl) Subject(o Observer) {
	sub.Obersers = append(sub.Obersers, o)
}

func (sub *SubjectImpl) Notify(msg string) {
	for _, ob := range sub.Obersers {
		ob.Update(msg)
	}
}

type Observer1 struct{}

func (Observer1) Update(msg string) {
	fmt.Println("observer1: " + msg)
}

type Observer2 struct{}

func (Observer2) Update(msg string) {
	fmt.Println("observer2: " + msg)
}

func main() {
	sub := &SubjectImpl{}
	sub.Subject(Observer1{})
	sub.Subject(Observer2{})
	sub.Notify("notify .......")
}
