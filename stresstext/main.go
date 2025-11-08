package main

import (
	"fmt"
)

func main() {
	var v interface{}
	v = 3000
	aray := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		aray[i] = v
	}
	v = aray
	fmt.Printf("%+v", v)
	// 	runtime.GOMAXPROCS(runtime.NumCPU()*30 )
	// 	appConfig := Config.GetAppConfig()
	// 	var instanceNum,baseNum int
	// 	flag.IntVar(&instanceNum,"n",10,"模拟设备台数")
	// 	flag.IntVar(&baseNum,"b",0,"设备起始编号")
	// 	flag.Parse()
	// 	if appConfig == nil {
	// 		fmt.Println("appConfig==nil")
	// 	}
	// 	fmt.Println("设备台数：",instanceNum,"	起始编号：",baseNum)
	// 	if instanceNum <= 0 {
	// 		fmt.Println("instanceNum<=0")
	// 	}
	// 	waitGroup := new(sync.WaitGroup)
	// 	instances := make([]*Instance.Instance, instanceNum)
	// 	Monitor.Start()
	// 	for i := 0; i < instanceNum; i++ {
	// 		waitGroup.Add(1)
	// 		go func(i int) {
	// 			defer waitGroup.Done()
	// 			Monitor.RunNumAdd(1)
	// 			instances[i] = Instance.NewInstance(&appConfig.Mqtt, fmt.Sprintf("STRESS_TEST_%05d", i+baseNum), "./model.json")
	// 			if instances[i] != nil {
	// 				instances[i].Start()
	// 			} else {
	// 				fmt.Printf("instances[%05d]==nil\n", i)
	// 			}
	// 		}(i)
	// 	}
	// 	waitGroup.Add(1)
	// 	running:=true
	// 	go func() {
	// 		for running{
	// 			time.Sleep(time.Duration(1)*time.Second)
	// 			//cmd:=exec.Command("cmd","/c","cls")
	// 			//cmd.Stdout=os.Stdout
	// 			//cmd.Run()
	// 			//fmt.Println(Monitor.GetInfo())
	// 		}
	// 		waitGroup.Done()
	// 	}()
	// 	var cmd string
	// CmdLoop:
	// 	for {
	// 		fmt.Scanln(&cmd)
	// 		switch cmd {
	// 		case "q", "quit", "exit":
	// 			for i := 0; i < instanceNum; i++ {
	// 				waitGroup.Add(1)
	// 				go func(i int) {
	// 					defer waitGroup.Done()
	// 					Monitor.RunNumDelete(1)
	// 					if instances[i] != nil {
	// 						instances[i].Stop()
	// 					}
	// 				}(i)
	// 			}
	// 			running=false
	// 			waitGroup.Wait()
	// 			Monitor.Stop()
	// 			return
	// 		default:
	// 			goto CmdLoop
	// 		}
	// 	}
}
