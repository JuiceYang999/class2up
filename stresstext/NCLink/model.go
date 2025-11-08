package NCLink

import (
	"encoding/json"
	"fmt"
	"time"
)

const RootType string = "NC_LINK_ROOT"
const DeviceTypes string = "MACHINE"
const SampleChannelType string = "SAMPLE_CHANNEL"

type ModelNode interface {
	NodeId() string
	NodeType() string
	NodeName() string
	NodeDescription() string
	NodePath() string
	SetNodePath(path string)
}

type Base struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	Number      string `json:"number,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Path        string `json:"-"`
}

func (self *Base) NodeId() string {
	return self.Id
}
func (self *Base) NodeType() string {
	return self.Type
}
func (self *Base) NodeNumber() string {
	return self.Number
}
func (self *Base) NodeName() string {
	return self.Name
}
func (self *Base) NodeDescription() string {
	return self.Description
}
func (self *Base) NodePath() string {
	return self.Path
}
func (self *Base) SetNodePath(path string) {
	self.Path = path
}

type SampleItem struct {
	Id   string    `json:"id"`
	node ModelNode `json:"-"`
}
type Config struct {
	Base
	Driver         *DriverItem  `json:"-"`
	SampleInterval uint16       `json:"sampleInterval,omitempty"`
	UploadInterval uint16       `json:"uploadInterval,omitempty"`
	SampleItems    []SampleItem `json:"ids,omitempty"`
}
type DataItem struct {
	Base
	Driver *DriverItem `json:"-"`
}
type Component struct {
	Base
	Configs    []Config    `json:"configs,omitempty"`
	DataItems  []DataItem  `json:"dataItems,omitempty"`
	Components []Component `json:"components,omitempty"`
}
type Device struct {
	Base
	Guid       string      `json:"guid,omitempty"`
	Version    string      `json:"version,omitempty"`
	Configs    []Config    `json:"configs,omitempty"`
	DataItems  []DataItem  `json:"dataItems,omitempty"`
	Components []Component `json:"components,omitempty"`
}
type Model struct {
	Base
	Devices []Device `json:"devices,omitempty"`
}

type QueryParameters struct {
	Operation string   `json:"operation,omitempty"`
	Indexes   []string `json:"indexes,omitempty"`
	Keys      []string `json:"keys,omitempty"`
	Offset    *int      `json:"offset,omitempty"`
	Length    *int      `json:"length,omitempty"`
}
type QueryRequestItem struct {
	Id     string          `json:"id"`
	Params *QueryParameters `json:"params,omitempty"`
}
type QueryResponse struct {
	Values []QueryResponseItem `json:"values"`
	ID	string `json:"@id,omitempty"`
}

type QueryResponseItem struct {
	Id     string          `json:"id"`
	Code   string          `json:"code"`
	Params *QueryParameters `json:"params,omitempty"`
	Value  []interface{}   `json:"values,omitempty"`
}
type QueryRequest struct {
	Ids []QueryRequestItem `json:"ids"`
	ID	string `json:"@id,omitempty"`
}

type SetParameters struct {
	Operation string      `json:"operation,omitempty""`
	Index     *int    `json:"index,omitempty"`
	Key       string    `json:"key,omitempty"`
	Offset    *int         `json:"offset,omitempty"`
	Length    *int         `json:"length,omitempty"`
	Value     interface{} `json:"value,omitempty"`
}
type SetRequestItem struct {
	Id     string        `json:"id"`
	Params *SetParameters `json:"params,omitempty"`
}
type SetRequest struct {
	Values []SetRequestItem `json:"values"`
	ID	string `json:"@id,omitempty"`
}

type SetResponseItem struct {
	Id     string        `json:"id"`
	Code   string        `json:"code"`
	Params *SetParameters `json:"params,omitempty"`
}
type SetResponse struct {
	Results []SetResponseItem `json:"results"`
	ID	string `json:"@id,omitempty"`
}

func CheckModel(model *Model) bool {
	if model.Type != RootType || model.Devices == nil || len(model.Devices) != 1 {
		return false
	}
	return true
}

type SampleTask interface {
	GetSampleData(device* VirtualDevice) (*SampleData,uint64)
	Id() string
	SampleInterval() uint16
	UploadInterval() uint16
}

type sampleTask struct {
	id             string
	sampleInterval uint16
	uploadInterval uint16
	SampleItems    []SampleItem
	t0             int64
}

type NCLinkObject struct {
	Model       Model
	IdNodeMap   map[string]ModelNode
	SampleTasks []SampleTask
}
type SampleDataRow struct {
	Data interface{} `json:"data"`
}
type SampleData struct {
	Id        string          `json:"id"`
	BeginTime string          `json:"beginTime"`
	Data      []SampleDataRow `json:"data"`
}

func (self *sampleTask) GetSampleData(device *VirtualDevice) (*SampleData,uint64) {
	var n uint16 = 0
	if self.t0 == 0 {
		self.t0 = time.Now().UnixNano()
		n = self.uploadInterval / self.sampleInterval
	} else {
		t1 := time.Now().UnixNano()
		n = uint16((t1 - self.t0) / (int64(self.sampleInterval) * 1000000))
		if n > 1000 {
			n = 1000
		}
		self.t0 = t1
	}
	sampleData := new(SampleData)
	sampleData.Id = self.id
	sampleData.BeginTime = fmt.Sprintf("%d", self.t0/1000000)
	if n < 1 {
		n = 1
	}
	for i := 0; i < len(self.SampleItems); i++ {
		dataRow := new(SampleDataRow)
		driverItem, ok := Driver[self.SampleItems[i].node.NodePath()]
		if !ok || driverItem == nil || driverItem.Sample == nil {
			dataRow.Data = nil
			continue
		}
		dataRow.Data, _ = driverItem.Sample(uint64(n),device)
		sampleData.Data = append(sampleData.Data, *dataRow)
	}
	return sampleData,uint64(n)
}
func (self *sampleTask) Id() string {
	return self.id
}
func (self *sampleTask) SampleInterval() uint16 {
	return self.sampleInterval
}
func (self *sampleTask) UploadInterval() uint16 {
	return self.uploadInterval
}
func buildComponentIdMap(component *Component, parent ModelNode, idMap map[string]ModelNode) bool {
	if len(component.Id) == 0 {
		return false
	}
	_, ok := idMap[component.Id]
	if ok {
		return false
	}
	//idMap[component.Id] = component
	if component.Number != "" {
		component.Path = fmt.Sprintf("%s:%s#%s", parent.NodePath(), component.Type, component.Number)
	} else {
		component.Path = fmt.Sprintf("%s:%s", parent.NodePath(), component.Type)
	}
	for i := 0; i < len(component.Configs); i++ {
		if len(component.Configs[i].Id) == 0 {
			return false
		}
		_, ok := idMap[component.Configs[i].Id]
		if ok {
			return false
		}
		idMap[component.Configs[i].Id] = &component.Configs[i]
		if component.Configs[i].Number != "" {
			component.Configs[i].Path = fmt.Sprintf("%s:%s#%s", component.NodePath(), component.Configs[i].Type, component.Configs[i].Number)
		} else {
			component.Configs[i].Path = fmt.Sprintf("%s:%s", component.NodePath(), component.Configs[i].Type)
		}
		driverItem, ok := Driver[component.Configs[i].Path]
		if ok {
			component.Configs[i].Driver = driverItem
		}
	}

	for i := 0; i < len(component.DataItems); i++ {
		if len(component.DataItems[i].Id) == 0 {
			return false
		}
		_, ok := idMap[component.DataItems[i].Id]
		if ok {
			return false
		}
		idMap[component.DataItems[i].Id] = &component.DataItems[i]
		if component.DataItems[i].Number != "" {
			component.DataItems[i].Path = fmt.Sprintf("%s:%s#%s", component.NodePath(), component.DataItems[i].Type, component.DataItems[i].Number)
		} else {
			component.DataItems[i].Path = fmt.Sprintf("%s:%s", component.NodePath(), component.DataItems[i].Type)
		}
		driverItem, ok := Driver[component.DataItems[i].Path]
		if ok {
			component.DataItems[i].Driver = driverItem
		}
	}
	for i := 0; i < len(component.Components); i++ {
		if component.Components[i].Number != "" {
			component.Components[i].Path = fmt.Sprintf("%s:%s#%s", component.NodePath(), component.Components[i].Type, component.Components[i].Number)
		} else {
			component.Components[i].Path = fmt.Sprintf("%s:%s", component.NodePath(), component.Components[i].Type)
		}
		if !buildComponentIdMap(&component.Components[i], component, idMap) {
			return false
		}
	}
	return true
}

func buildDeviceIdMap(device *Device, parent ModelNode, idMap map[string]ModelNode) bool {
	if len(device.Id) == 0 {
		return false
	}
	_, ok := idMap[device.Id]
	if ok {
		return false
	}
	//idMap[device.Id] = device
	if device.Number != "" {
		device.Path = fmt.Sprintf("%s:%s#%s", parent.NodePath(), device.Type, device.Number)
	} else {
		device.Path = fmt.Sprintf("%s:%s", parent.NodePath(), device.Type)
	}
	for i := 0; i < len(device.Configs); i++ {
		if len(device.Configs[i].Id) == 0 {
			return false
		}
		_, ok := idMap[device.Configs[i].Id]
		if ok {
			return false
		}
		idMap[device.Configs[i].Id] = &device.Configs[i]
		if device.Configs[i].Number != "" {
			device.Configs[i].Path = fmt.Sprintf("%s:%s#%s", device.NodePath(), device.Configs[i].Type, device.Configs[i].Number)
		} else {
			device.Configs[i].Path = fmt.Sprintf("%s:%s", device.NodePath(), device.Configs[i].Type)
		}
		driverItem, ok := Driver[device.Configs[i].Path]
		if ok {
			device.Configs[i].Driver = driverItem
		}
	}

	for i := 0; i < len(device.DataItems); i++ {
		if len(device.DataItems[i].Id) == 0 {
			return false
		}
		_, ok := idMap[device.DataItems[i].Id]
		if ok {
			return false
		}
		idMap[device.DataItems[i].Id] = &device.DataItems[i]
		if device.DataItems[i].Number != "" {
			device.DataItems[i].Path = fmt.Sprintf("%s:%s#%s", device.NodePath(), device.DataItems[i].Type, device.DataItems[i].Number)
		} else {
			device.DataItems[i].Path = fmt.Sprintf("%s:%s", device.NodePath(), device.DataItems[i].Type)
		}
		driverItem, ok := Driver[device.DataItems[i].Path]
		if ok {
			device.DataItems[i].Driver = driverItem
		}
	}
	for i := 0; i < len(device.Components); i++ {
		if !buildComponentIdMap(&device.Components[i], device, idMap) {
			return false
		}
	}
	return true
}

func buildRootIdMap(model *Model, idMap map[string]ModelNode) bool {
	if len(model.Id) == 0 || idMap[model.Id] != nil {
		return false
	}
	idMap[model.Id] = model
	model.Path = model.Type
	for i := 0; i < len(model.Devices); i++ {
		if !buildDeviceIdMap(&model.Devices[i], model, idMap) {
			return false
		}
	}
	return true
}

func buildModelIdMap(model *Model) map[string]ModelNode {
	idMap := make(map[string]ModelNode)
	if !buildRootIdMap(model, idMap) {
		return nil
	}
	return idMap
}

func NewNCLinkObject(modelText []byte) *NCLinkObject {
	nclinkObject := &NCLinkObject{}
	err := json.Unmarshal(modelText, &nclinkObject.Model)
	if err != nil {
		fmt.Printf("json.Unmarshal:%v\n", err.Error())
		return nil
	}

	if !CheckModel(&nclinkObject.Model) {
		fmt.Printf("CheckModel Failed\n")
		return nil
	}
	nclinkObject.IdNodeMap = buildModelIdMap(&nclinkObject.Model)

	if nclinkObject.IdNodeMap == nil {
		return nil
	}

	if len(nclinkObject.Model.Devices) == 1 {
	ConfigLoop:
		for i := 0; i < len(nclinkObject.Model.Devices[0].Configs); i++ {
			if nclinkObject.Model.Devices[0].Configs[i].Type == SampleChannelType {
				for j := 0; j < len(nclinkObject.Model.Devices[0].Configs[i].SampleItems); j++ {
					node, ok := nclinkObject.IdNodeMap[nclinkObject.Model.Devices[0].Configs[i].SampleItems[j].Id]
					if !ok {
						continue ConfigLoop
					}
					nclinkObject.Model.Devices[0].Configs[i].SampleItems[j].node = node
				}
				sampleTask := &sampleTask{
					id:             nclinkObject.Model.Devices[0].Configs[i].Id,
					sampleInterval: nclinkObject.Model.Devices[0].Configs[i].SampleInterval,
					uploadInterval: nclinkObject.Model.Devices[0].Configs[i].UploadInterval,
					SampleItems:    nclinkObject.Model.Devices[0].Configs[i].SampleItems,
					t0:             0,
				}
				nclinkObject.SampleTasks = append(nclinkObject.SampleTasks, sampleTask)
			}
		}
	}

	return nclinkObject
}
