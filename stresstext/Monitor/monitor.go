package Monitor

import (
	"fmt"
	"net/http"
)

type Monitor struct {
}

var runNumCh chan int
var runningCh chan bool
var exitCh chan bool
var runNum = 0
var connectedNumCh chan int
var connectedNum =0


func RunNumAdd(n int) {
	runNumCh <- n
}
func RunNumDelete(n int)  {
	runNumCh <- -n
}
func ConnectedNumAdd(n int){
	//fmt.Printf("%s:%8d\n","ConnectedNumAdd",connectedNum)
	connectedNumCh<-n
	//fmt.Printf("%s:%8d\n","ConnectedNumAdd---",connectedNum)
}
func ConnectedNumDelete(n int)  {
	connectedNumCh<- -n
}

func (self *Monitor) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte(fmt.Sprintf("<div>Running Instance Number:%d</div></br>",runNum)))
	writer.Write([]byte(fmt.Sprintf("<div>Connected Instance Number:%d<div></br>",connectedNum)))
}

var monitor = &Monitor{}

func Start() {
	go func() {
		server := http.Server{
			Addr:    "localhost:12222",
			Handler: monitor,
		}
		err := server.ListenAndServe()
		if err != nil {
			fmt.Printf("server.ListenAndServe Failed:%v\n", err)
		} else {
			fmt.Println("Monitor Started")
		}
	}()
	runNumCh = make(chan int, 1)
	runningCh = make(chan bool, 1)
	exitCh = make(chan bool, 1)
	go func() {
	Loop:
		for {
			select {
			case i := <-runNumCh:
				runNum += i
			case i := <-connectedNumCh:
				connectedNum += i
			case running := <-runningCh:
				if !running{
					break Loop
				}
			}
		}
		exitCh<-true
	}()
}
func Stop() {
	runningCh<-false
	<-exitCh
}

func GetInfo() string {
	return fmt.Sprintf("%s:%8d","Connected",connectedNum)
}
