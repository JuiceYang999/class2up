package Instance

import (
	"StressTest-INC-Cloud/Config"
	"StressTest-INC-Cloud/Monitor"
	"StressTest-INC-Cloud/NCLink"
	"container/list"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type Instance struct {
	Config.MqttConfig
	ClientID                  string
	modelFile                 string
	nclinkObject              *NCLink.NCLinkObject
	mqttClient                mqtt.Client
	run                       bool
	lock                      sync.Mutex
	queryRequestTopic         string
	queryResponseTopic        string
	setRequestTopic           string
	setResponseTopic          string
	probeQueryRequestTopic    string
	probeQueryResponseTopic   string
	probeSetRequestTopic      string
	probeSetResponseTopic     string
	probeVersionTopic         string
	probeVersionResponseTopic string
	registerRequestTopic      string
	registerResponseTopic     string
	waitGroup                 sync.WaitGroup
	topics                    map[string]byte
	sendQueue                 list.List
	sendQueueLock             *sync.Mutex
	sendQueueCond             *sync.Cond
	device                    *NCLink.VirtualDevice
}
type Message struct {
	topic   string
	message []byte
}

func NewInstance(mqttConfig *Config.MqttConfig, clientID, modelFile string) *Instance {

	instance := &Instance{MqttConfig: *mqttConfig, ClientID: clientID}

	fileInfo, err := os.Stat(clientID)
	if err != nil || !fileInfo.IsDir() {
		err = os.MkdirAll(clientID, os.ModePerm)
		if err != nil {
			fmt.Printf("[%s]Failed to Create Dir \"%s\":%v\n", clientID, clientID, err)
		}
		err = os.MkdirAll(clientID+"/prog", os.ModePerm)
		if err != nil {
			fmt.Printf("[%s]Failed to Create Dir \"%s\":%v\n", clientID, clientID+"/prog", err)
		}
	}

	fileInfo, err = os.Stat(clientID + "/model.json")
	if err != nil || fileInfo.IsDir() {
		instance.modelFile = modelFile
	} else {
		instance.modelFile = clientID + "/model.json"
	}

	modelText, err := ioutil.ReadFile(instance.modelFile)
	if err != nil {
		fmt.Printf("[%s]Failed to Open Model File:%s:%v\n", clientID, modelFile, err)
		return nil
	}

	instance.nclinkObject = NCLink.NewNCLinkObject(modelText)
	if instance.nclinkObject == nil {
		fmt.Printf("[%s]Failed to Create NC-Link Object", clientID)
		return nil
	}
	instance.queryRequestTopic = NCLink.QueryRequestTopic(instance.ClientID)
	instance.queryResponseTopic = NCLink.QueryResponseTopic(instance.ClientID)
	instance.setRequestTopic = NCLink.SetRequestTopic(instance.ClientID)
	instance.setResponseTopic = NCLink.SetResponseTopic(instance.ClientID)
	instance.probeQueryRequestTopic = NCLink.ProbeQueryRequestTopic(instance.ClientID)
	instance.probeQueryResponseTopic = NCLink.ProbeQueryResponseTopic(instance.ClientID)
	instance.probeSetRequestTopic = NCLink.ProbeSetRequestTopic(instance.ClientID)
	instance.probeSetResponseTopic = NCLink.ProbeSetResponseTopic(instance.ClientID)
	instance.probeVersionTopic = NCLink.ProbeVersionTopic(instance.ClientID)
	instance.probeVersionResponseTopic = NCLink.ProbeVersionResponseTopic(instance.ClientID)
	instance.registerRequestTopic = NCLink.RegisterRequestTopic(instance.ClientID)
	instance.registerResponseTopic = NCLink.RegisterResponseTopic(instance.ClientID)

	instance.topics = make(map[string]byte)
	instance.topics[instance.probeSetRequestTopic] = 0
	instance.topics[instance.probeVersionResponseTopic] = 0
	instance.topics[instance.registerResponseTopic] = 0
	instance.topics[instance.probeQueryRequestTopic] = 0
	instance.topics[instance.setRequestTopic] = 0
	instance.topics[instance.queryRequestTopic] = 0

	instance.sendQueueLock = new(sync.Mutex)
	instance.sendQueueCond = sync.NewCond(instance.sendQueueLock)

	instance.device = NCLink.NewVirtualDevice(clientID)
	clientOptions := mqtt.NewClientOptions()
	clientOptions.AddBroker(fmt.Sprintf("tcp://%s:%d", instance.IP, instance.Port))
	clientOptions.SetUsername(instance.UserName)
	clientOptions.SetPassword(instance.Password)
	clientOptions.SetConnectionLostHandler(
		func(client mqtt.Client, err error) {
			fmt.Printf("[%s]Connection Lost %s\n", instance.ClientID, err.Error())
			Monitor.ConnectedNumDelete(1)
		})
	clientOptions.SetOnConnectHandler(
		func(client mqtt.Client) {
			fmt.Printf("[%s]Connected\n", instance.ClientID)
			for {
				token := client.SubscribeMultiple(instance.topics,
					func(client mqtt.Client, message mqtt.Message) {
						switch message.Topic() {
						case instance.probeQueryRequestTopic:
							model, err := json.Marshal(instance.nclinkObject.Model)
							if err == nil {
								instance.EnqueueMessage(instance.probeQueryResponseTopic, []byte(fmt.Sprintf("{\"code\":\"ok\",\"probe\":%s}", string(model))))
							} else {
								fmt.Printf("[%s]Failed To Serialize Model Object to String\n", instance.ClientID)
							}
						case instance.queryRequestTopic:
							jsonReq := NCLink.QueryRequest{}
							err := json.Unmarshal(message.Payload(), &jsonReq)
							if err != nil {
								fmt.Println("[%s]Failed To Parse Query Request %s\n", instance.ClientID, string(message.Payload()))
							} else {
								response := NCLink.QueryResponse{}
								response.ID = jsonReq.ID
								var respItem *NCLink.QueryResponseItem
								for _, item := range jsonReq.Ids {
									if item.Id == "" {
										return
									}
									node, ok := instance.nclinkObject.IdNodeMap[item.Id]
									if !ok || node == nil {
										respItem = &NCLink.QueryResponseItem{}
										respItem.Id = item.Id
										respItem.Params = &NCLink.QueryParameters{}
										*(respItem.Params) = *item.Params
										respItem.Code = "NG"
										continue
									}
									executor, ok := NCLink.Driver[node.NodePath()]
									if !ok || executor == nil || executor.Get == nil {
										respItem = &NCLink.QueryResponseItem{}
										respItem.Id = item.Id
										if item.Params != nil {
											respItem.Params = &NCLink.QueryParameters{}
											*(respItem.Params) = *(item.Params)
										}
										respItem.Code = "NG"
										continue
									}
									respItem, _ = executor.Get(item, instance.device)
									response.Values = append(response.Values, *respItem)
								}
								jsonResp, err := json.Marshal(response)
								if err == nil {
									instance.EnqueueMessage(instance.queryResponseTopic, jsonResp)
								}
							}
						case instance.setRequestTopic:
							jsonReq := NCLink.SetRequest{}
							err := json.Unmarshal(message.Payload(), &jsonReq)
							if err != nil {
								fmt.Println("[%s]Failed To Parse Query Request %s\n", instance.ClientID, string(message.Payload()))
							} else {
								response := NCLink.SetResponse{}
								response.ID = jsonReq.ID
								var respItem *NCLink.SetResponseItem
								for _, item := range jsonReq.Values {
									if item.Id == "" {
										return
									}
									node, ok := instance.nclinkObject.IdNodeMap[item.Id]
									if !ok || node == nil {
										respItem = &NCLink.SetResponseItem{}
										respItem.Id = item.Id
										respItem.Params = &NCLink.SetParameters{}
										*(respItem.Params) = *item.Params
										respItem.Code = "NG"
										continue
									}
									executor, ok := NCLink.Driver[node.NodePath()]
									if !ok || executor == nil || executor.Get == nil {
										respItem = &NCLink.SetResponseItem{}
										respItem.Id = item.Id
										respItem.Params = &NCLink.SetParameters{}
										*(respItem.Params) = *(item.Params)
										respItem.Code = "NG"
										continue
									}
									respItem, _ = executor.Set(item, instance.device)
									response.Results = append(response.Results, *respItem)
								}
								jsonResp, err := json.Marshal(response)
								if err == nil {
									instance.EnqueueMessage(instance.setResponseTopic, jsonResp)
								}
							}
						default:
						}
					})
				if token.WaitTimeout(time.Duration(instance.ConnectTimeout*1000000)) && token.Error() == nil {
					break
				} else {
					fmt.Printf("[%s]Subscribe Topics Failed\n", instance.ClientID)
				}
			}
			instance.EnqueueMessage(instance.registerRequestTopic, []byte(fmt.Sprintf("{\"deviceid\":\"%s\"}", instance.ClientID)))
			instance.EnqueueMessage(instance.probeVersionTopic, []byte(fmt.Sprintf("{\"version\":\"%s\"}", instance.nclinkObject.Model.Devices[0].Version)))
			Monitor.ConnectedNumAdd(1)
		})
	clientOptions.SetAutoReconnect(instance.AutoReconnect)
	clientOptions.SetClientID(instance.ClientID)
	instance.mqttClient = mqtt.NewClient(clientOptions)
	return instance
}

func (self *Instance) isRunning() bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.run
}
func (self *Instance) setRunning(run bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.run = run
}
func (self *Instance) EnqueueMessage(topic string, message []byte) {
	self.sendQueueLock.Lock()
	defer self.sendQueueLock.Unlock()
	self.sendQueue.PushBack(&Message{topic, message})
	if self.sendQueue.Len() > 100 {
		self.sendQueue.Remove(self.sendQueue.Front())
		fmt.Printf("[%s]Queue is Full.Remove the Last Message.\n", self.ClientID)
	}
	if self.sendQueue.Len() == 1 {
		self.sendQueueCond.Signal()
	}
}

func (self *Instance) DequeueMessage() (*Message, bool) {
	self.sendQueueLock.Lock()
	defer self.sendQueueLock.Unlock()
	for self.sendQueue.Len() == 0 && self.isRunning() {
		self.sendQueueCond.Wait()
	}
	if self.sendQueue.Len() > 0 {
		return self.sendQueue.Remove(self.sendQueue.Front()).(*Message), true
	} else {
		return nil, false
	}
}

func (self *Instance) SendTask() {
	defer self.waitGroup.Done()
	for self.isRunning() {
		message, ok := self.DequeueMessage()
		if ok {
			token := self.mqttClient.Publish(message.topic, 2, false, message.message)
			ok := token.WaitTimeout(1000000000 * 3)
			if !ok || token.Error() != nil {
				fmt.Printf("[%s]Failed to Publish Sample Message.topic:%s %v\n", self.ClientID, message.topic, token.Error())
			}
		}
	}
}

func (self *Instance) SampleTask(task NCLink.SampleTask) {
	defer self.waitGroup.Done()
	//fmt.Printf("[%s](%s) is Running \n", self.ClientID, task.Id())
	topic := fmt.Sprintf("Sample/%s/%s", self.ClientID, task.Id())
	no := uint32(0)
	for self.isRunning() {
		t0 := time.Now().UnixNano()
		sampleData, n := task.GetSampleData(self.device)
		message, err := json.Marshal(sampleData)
		if err == nil {
			self.EnqueueMessage(topic, message)
		}
		if task.SampleInterval() == 1 {
			self.device.SampleHeartBeat += n
		}
		t1 := time.Now().UnixNano()
		time.Sleep(time.Duration(int64(task.UploadInterval())*1000000 - (t1 - t0)))
		no++
	}
	//fmt.Printf("[%s](%s) Stopped \n", self.ClientID, task.Id())
}
func (self *Instance) Start() {
	if self.isRunning() {
		return
	}
	self.setRunning(true)
	for !self.mqttClient.IsConnected() && self.isRunning() {
		fmt.Printf("[%s]is Connecting\n", self.ClientID)
		token := self.mqttClient.Connect()
		if token.WaitTimeout(time.Duration(self.ConnectTimeout)*time.Millisecond) && token.Error() == nil {
			break
		} else {
			fmt.Printf("[%s]Connect Failed %v\n", self.ClientID, token.Error())
		}
	}

	self.waitGroup.Add(1)
	go self.SendTask()
	for i := 0; i < len(self.nclinkObject.SampleTasks); i++ {
		self.waitGroup.Add(1)
		go self.SampleTask(self.nclinkObject.SampleTasks[i])
	}

}

func (self *Instance) Stop() {
	self.setRunning(false)
	self.sendQueueLock.Lock()
	self.sendQueueCond.Signal()
	self.sendQueueLock.Unlock()
	self.waitGroup.Wait()
	self.mqttClient.Disconnect(3000)
	fmt.Printf("[%s]Stopped\n", self.ClientID)
}
