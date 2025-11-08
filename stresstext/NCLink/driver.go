package NCLink

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

//type QuerySetter interface {
//	Get(...interface{})(interface{},bool)
//	Set(...interface{})(interface{},bool)
//}
//type Getter func(count uint64, no uint32, num uint16, a ...interface{}) (interface{}, bool)
//type Setter func(count uint64, no uint32, num uint16, a ...interface{}) (interface{}, bool)

var AxisChanSystemData = [][]interface{}{
	{"X", 1, 0, 0, -1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 35.7, 0, 0, 0, 1, 0, "0.0", "180UD-35A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{"Y", 1, 0, 1, -1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 35.7, 0, 0, 0, 1, 0, "0.0", "180UD-35A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{"Z", 1, 0, 2, -1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 35.7, 0, 0, 0, 1, 0, "0.0", "180UD-35A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{},
	{},
	{"C", 10, 0, -1, 0, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 35.7, 0, 0, 0, 1, 0, "0.0", "180US-35A", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},

	{1, 1, 39, 0, "CH0", 0, 0, 0, 0, 100, 25, -1, 0, 0, 0, 0, 2, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 58, 58, 0, 0, nil, 810211375270854656, nil, 0, "X", "S", -1, 5, 5, 0, 0, 100, 1.62245919E-254, 0, 0, 0, 0, 0, 0, 0, 10000, 0, 0, nil, 0, 0, 0, 0},
	{},
	{},
	{},
	{1, 100000, 100000, 1, 1, 1, 0, 1.62191936E-254, 4, 2, 0, "274.40992", "BAA08", "12.40992", 0, 1, 0, 0, 0, 0, nil, nil, nil, 0, "2.41.00_a", "122FFD6BEF36C6E", "", "", "", "", "", "", 1, 810210962953994241, 0, nil, 8, 8, 8, 0, "0.0", "2.41.00_a", "0.0", ".2020-04-16 23:53:20", ".2020-04-16 23:53:20", "V1.0", "V1.0", "V1.0", "0.0", "V1.0"},
}

type SampleBuffer struct {
	Status [1000]int
	Axis   [6][5][1000]float32
	Prog   [1000]int
	Tool   [1000]int
	Line   [1000]int
}
type ProgInfoItem struct {
	name       string
	lineNumber uint32
}
type VirtualDevice struct {
	ProgDir         string
	ToolParam       [20]ToolParam
	Crds            [180]Crd
	Axes            [6]Axis
	Regs            [6][]uint32
	RunState        *RunState
	SampleHeartBeat uint64
	SampleBuffer    SampleBuffer
	ProgList        []ProgInfoItem
}

type Getter func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool)
type Setter func(item SetRequestItem, virtualDevice *VirtualDevice) (*SetResponseItem, bool)
type Sampler func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool)
type DriverItem struct {
	Get    Getter
	Set    Setter
	Sample Sampler
}

type FileAttrib struct {
	Type       string `json:"type"`
	Size       int64  `json:"size"`
	ChangeTime string `json:"changeTime"`
}
type ToolParam struct {
	Length         *float32 `json:"length,omitempty"`
	Radius         *float32 `json:"radius"`
	LengthAbrasion *float32 `json:"length_abrasion,omitempty"`
	RadiusAbrasion *float32 `json:"radius_abrasion,omitempty"`
}
type Crd struct {
	X *float32 `json:"x,omitempty"`
	Y *float32 `json:"y,omitempty"`
	Z *float32 `json:"z,omitempty"`
	C *float32 `json:"c,omitempty"`
}
type Axis struct {
	Name        string
	Number      int
	Type        string
	ActSpeed    float32
	CmdSpeed    float32
	ActPosition float32
	CmdPosition float32
	Current     float32
}
type RunState struct {
	Line       uint64
	State      int64
	PartCount  int64
	ProgNumber int64
	ToolNumber int64
}

var statusS = []string{"free", "running", "holding"}
var statusN = []int{0, 1, 2}

var Driver = make(map[string]*DriverItem)

func NewVirtualDevice(deviceId string) *VirtualDevice {
	virtualDevice := &VirtualDevice{}
	virtualDevice.ProgDir = deviceId + "/prog/"
	for i := 0; i < len(virtualDevice.ToolParam); i++ {
		virtualDevice.ToolParam[i].Length = new(float32)
		virtualDevice.ToolParam[i].Radius = new(float32)
		virtualDevice.ToolParam[i].LengthAbrasion = new(float32)
		virtualDevice.ToolParam[i].RadiusAbrasion = new(float32)
	}
	for i := 0; i < len(virtualDevice.Crds); i++ {
		virtualDevice.Crds[i].X = new(float32)
		virtualDevice.Crds[i].Y = new(float32)
		virtualDevice.Crds[i].Z = new(float32)
		virtualDevice.Crds[i].C = new(float32)
	}
	virtualDevice.Axes[0].Name = "X"
	virtualDevice.Axes[0].Type = "linear"
	virtualDevice.Axes[0].Number = 0
	virtualDevice.Axes[1].Name = "Y"
	virtualDevice.Axes[1].Type = "linear"
	virtualDevice.Axes[1].Number = 1
	virtualDevice.Axes[2].Name = "Z"
	virtualDevice.Axes[2].Type = "linear"
	virtualDevice.Axes[2].Number = 2
	virtualDevice.Axes[3].Name = "C"
	virtualDevice.Axes[3].Type = "rotary"
	virtualDevice.Axes[3].Number = 5

	virtualDevice.Regs[0] = make([]uint32, 512)  //x
	virtualDevice.Regs[1] = make([]uint32, 512)  //y
	virtualDevice.Regs[2] = make([]uint32, 3728) //f
	virtualDevice.Regs[3] = make([]uint32, 3728) //g
	virtualDevice.Regs[4] = make([]uint32, 2048) //r
	virtualDevice.Regs[5] = make([]uint32, 1722) //b

	virtualDevice.RunState = &RunState{State: 1}

	prefixLen := len(virtualDevice.ProgDir)
	LineNumberOf := func(fileName string) uint32 {
		file, err := os.Open(fileName)
		if err != nil {
			return 0
		}
		defer file.Close()
		fd := bufio.NewReader(file)
		count := uint32(0)
		for {
			_, err := fd.ReadString('\n')
			if err != nil {
				break
			}
			count++
		}
		return count
	}
	var IterateDir func(dir string) []ProgInfoItem
	IterateDir = func(dir string) []ProgInfoItem {
		files := make([]ProgInfoItem, 0)
		fileInfos, err := ioutil.ReadDir(dir)
		if err != nil {
			return files
		}
		for _, fileInfo := range fileInfos {
			files = append(files, ProgInfoItem{(dir + fileInfo.Name())[prefixLen:], LineNumberOf(dir + fileInfo.Name())})
			if fileInfo.IsDir() {
				files = append(files, IterateDir(dir+fileInfo.Name())...)
			}
		}
		return files
	}
	virtualDevice.ProgList = IterateDir(virtualDevice.ProgDir)
	return virtualDevice
}

func init() {
	paramInitState := ParamInit()
	if paramInitState != true {
		fmt.Println("Failed to Init Parameter Simulator")
	}
	dataPool := new([100][1000]float32) //预先生成100组，每组10000个点的数据，采样数据直接从这里面取
	for i := 0; i < 100; i++ {
		for j := 0; j < 1000; j++ {
			dataPool[i][j] = float32(int32(float64(i+1)*100000.0*math.Sin(3.14*2.0*float64(j)/200.0))) / 10000.0
		}
	}

	Driver["NC_LINK_ROOT:MACHINE:STATUS"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					var status string
					switch virtualDevice.RunState.State {
					case 0:
						status = "running"
					case 1:
						status = "free"
					case 2:
						status = "holding"
					default:
						response.Code = "NG"
						return response, false
					}
					response.Value = append(response.Value, status)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			for i := uint64(0); i < n; i++ {
				virtualDevice.SampleBuffer.Status[i] = status
			}
			virtualDevice.RunState.State = int64(status)
			return virtualDevice.SampleBuffer.Status[0:n], true
		},
	}

	Driver["NC_LINK_ROOT:MACHINE:FEED_SPEED"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					var v uint32
					switch virtualDevice.RunState.State {
					case 0:
						v = 3000
					case 1, 2:
						v = 0
					default:
						response.Code = "NG"
						return response, false
					}
					response.Value = append(response.Value, v)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			data := make([]uint32, n)
			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			var v uint32
			switch status {
			case 0:
				v = 3000
			default:
				v = 0
			}
			for i := uint64(0); i < n; i++ {
				data[i] = v
			}
			return data, true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:FEED_OVERRIDE"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					response.Value = append(response.Value, 100)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			data := make([]uint32, n)
			for i := uint64(0); i < n; i++ {
				data[i] = 100
			}
			return data, true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:SPDL_OVERRIDE"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					response.Value = append(response.Value, 100)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			data := make([]uint32, n)
			for i := uint64(0); i < n; i++ {
				data[i] = 100
			}
			return data, true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:PART_COUNT"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					response.Value = append(response.Value, virtualDevice.SampleHeartBeat/300000)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			data := make([]uint32, n)
			for i := uint64(0); i < n; i++ {
				data[i] = uint32((virtualDevice.SampleHeartBeat + i) / 300000)
			}
			return data, true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:WARNING"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					response.Value = append(response.Value, []int{})
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			data := make([][]uint32, n)
			for i := uint64(0); i < n; i++ {
				data[i] = []uint32{}
			}
			return data, true
		},
	}

	funcGetAxisPosition := func(axis int, item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
		response := &QueryResponseItem{}
		response.Id = item.Id
		if virtualDevice == nil {
			response.Code = "NG"
			return response, false
		}
		var operation string
		if item.Params != nil && item.Params.Operation != "" {
			operation = item.Params.Operation
			response.Params = &QueryParameters{}
			response.Params.Operation = item.Params.Operation
		} else {
			operation = "get_value"
		}
		switch operation {
		case "get_value":
			return func() (*QueryResponseItem, bool) {
				virtualDevice.Axes[axis].ActSpeed = dataPool[axis*5][virtualDevice.SampleHeartBeat%1000]
				response.Value = append(response.Value, virtualDevice.Axes[axis].ActPosition)
				response.Code = "OK"
				return response, true
			}()
		default:
			response.Code = "NG"
			return response, false
		}
	}
	funcGetAxisSpeed := func(axis int, item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
		response := &QueryResponseItem{}
		response.Id = item.Id
		if virtualDevice == nil {
			response.Code = "NG"
			return response, false
		}
		var operation string
		if item.Params != nil && item.Params.Operation != "" {
			operation = item.Params.Operation
			response.Params = &QueryParameters{}
			response.Params.Operation = item.Params.Operation
		} else {
			operation = "get_value"
		}
		switch operation {
		case "get_value":
			return func() (*QueryResponseItem, bool) {
				virtualDevice.Axes[axis].ActSpeed = dataPool[axis*5+1][virtualDevice.SampleHeartBeat%1000]
				response.Value = append(response.Value, virtualDevice.Axes[axis].ActSpeed)
				response.Code = "OK"
				return response, true
			}()
		default:
			response.Code = "NG"
			return response, false
		}
	}
	funcGetAxisCurrent := func(axis int, item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
		response := &QueryResponseItem{}
		response.Id = item.Id
		if virtualDevice == nil {
			response.Code = "NG"
			return response, false
		}
		var operation string
		if item.Params != nil && item.Params.Operation != "" {
			operation = item.Params.Operation
			response.Params = &QueryParameters{}
			response.Params.Operation = item.Params.Operation
		} else {
			operation = "get_value"
		}
		switch operation {
		case "get_value":
			return func() (*QueryResponseItem, bool) {
				virtualDevice.Axes[axis].ActSpeed = dataPool[axis*5+4][virtualDevice.SampleHeartBeat%1000]
				response.Value = append(response.Value, virtualDevice.Axes[axis].Current)
				response.Code = "OK"
				return response, true
			}()
		default:
			response.Code = "NG"
			return response, false
		}
	}
	funcGetAxisName := func(axis int, item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
		response := &QueryResponseItem{}
		response.Id = item.Id
		if virtualDevice == nil {
			response.Code = "NG"
			return response, false
		}
		var operation string
		if item.Params != nil && item.Params.Operation != "" {
			operation = item.Params.Operation
			response.Params = &QueryParameters{}
			response.Params.Operation = item.Params.Operation
		} else {
			operation = "get_value"
		}
		switch operation {
		case "get_value":
			return func() (*QueryResponseItem, bool) {
				response.Value = append(response.Value, virtualDevice.Axes[axis].Name)
				response.Code = "OK"
				return response, true
			}()
		default:
			response.Code = "NG"
			return response, false
		}
	}
	funcGetAxisType := func(axis int, item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
		response := &QueryResponseItem{}
		response.Id = item.Id
		if virtualDevice == nil {
			response.Code = "NG"
			return response, false
		}
		var operation string
		if item.Params != nil && item.Params.Operation != "" {
			operation = item.Params.Operation
			response.Params = &QueryParameters{}
			response.Params.Operation = item.Params.Operation
		} else {
			operation = "get_value"
		}
		switch operation {
		case "get_value":
			return func() (*QueryResponseItem, bool) {
				response.Value = append(response.Value, virtualDevice.Axes[axis].Type)
				response.Code = "OK"
				return response, true
			}()
		default:
			response.Code = "NG"
			return response, false
		}
	}
	funcGetAxisNumber := func(axis int, item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
		response := &QueryResponseItem{}
		response.Id = item.Id
		if virtualDevice == nil {
			response.Code = "NG"
			return response, false
		}
		var operation string
		if item.Params != nil && item.Params.Operation != "" {
			operation = item.Params.Operation
			response.Params = &QueryParameters{}
			response.Params.Operation = item.Params.Operation
		} else {
			operation = "get_value"
		}
		switch operation {
		case "get_value":
			return func() (*QueryResponseItem, bool) {
				response.Value = append(response.Value, virtualDevice.Axes[axis].Number)
				response.Code = "OK"
				return response, true
			}()
		default:
			response.Code = "NG"
			return response, false
		}
	}

	Driver["NC_LINK_ROOT:MACHINE:AXIS#0:NAME"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisName(0, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#0:TYPE"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisType(0, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#0:NUMBER"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisNumber(0, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#0:SCREW:POSITION"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisPosition(0, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[0][0][i-s] = dataPool[0][i%1000]
				}
				virtualDevice.Axes[0].ActPosition = dataPool[0][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[0].ActPosition
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[0][0][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[0][0][0:n], true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#0:SCREW:SPEED"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisSpeed(0, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[0][1][i-s] = dataPool[1][i%1000]
				}
				virtualDevice.Axes[0].ActSpeed = dataPool[1][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[0].ActSpeed
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[0][1][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[0][1][0:n], true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#0:SERVO_DRIVER:POSITION"] = Driver["NC_LINK_ROOT:MACHINE:AXIS#0:SCREW:POSITION"]
	Driver["NC_LINK_ROOT:MACHINE:AXIS#0:SERVO_DRIVER:SPEED"] = Driver["NC_LINK_ROOT:MACHINE:AXIS#0:SCREW:SPEED"]
	Driver["NC_LINK_ROOT:MACHINE:AXIS#0:MOTOR:CURRENT"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisCurrent(0, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[0][4][i-s] = dataPool[4][i%1000]
				}
				virtualDevice.Axes[0].Current = dataPool[1][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[0].Current
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[0][4][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[0][4][0:n], true
		},
	}

	Driver["NC_LINK_ROOT:MACHINE:AXIS#1:NAME"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisName(1, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#1:TYPE"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisType(1, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#1:NUMBER"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisNumber(1, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#1:SCREW:POSITION"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisPosition(1, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[1][0][i-s] = dataPool[5][i%1000]
				}
				virtualDevice.Axes[1].ActPosition = dataPool[5][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[1].ActPosition
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[1][0][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[1][0][0:n], true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#1:SCREW:SPEED"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisSpeed(1, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[1][1][i-s] = dataPool[6][i%1000]
				}
				virtualDevice.Axes[1].ActSpeed = dataPool[6][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[1].ActSpeed
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[1][1][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[1][1][0:n], true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#1:SERVO_DRIVER:POSITION"] = Driver["NC_LINK_ROOT:MACHINE:AXIS#1:SCREW:POSITION"]
	Driver["NC_LINK_ROOT:MACHINE:AXIS#1:SERVO_DRIVER:SPEED"] = Driver["NC_LINK_ROOT:MACHINE:AXIS#1:SCREW:SPEED"]
	Driver["NC_LINK_ROOT:MACHINE:AXIS#1:MOTOR:CURRENT"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisCurrent(1, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[1][4][i-s] = dataPool[9][i%1000]
				}
				virtualDevice.Axes[1].Current = dataPool[9][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[1].Current
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[1][4][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[1][4][0:n], true
		},
	}

	Driver["NC_LINK_ROOT:MACHINE:AXIS#2:NAME"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisName(2, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#2:TYPE"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisType(2, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#2:NUMBER"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisNumber(2, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#2:SCREW:POSITION"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisPosition(2, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[2][0][i-s] = dataPool[10][i%1000]
				}
				virtualDevice.Axes[5].ActPosition = dataPool[10][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[5].ActPosition
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[2][0][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[2][0][0:n], true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#2:SCREW:SPEED"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisSpeed(2, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[2][1][i-s] = dataPool[11][i%1000]
				}
				virtualDevice.Axes[2].ActSpeed = dataPool[11][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[2].ActSpeed
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[2][1][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[2][1][0:n], true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#2:SERVO_DRIVER:POSITION"] = Driver["NC_LINK_ROOT:MACHINE:AXIS#2:SCREW:POSITION"]
	Driver["NC_LINK_ROOT:MACHINE:AXIS#2:SERVO_DRIVER:SPEED"] = Driver["NC_LINK_ROOT:MACHINE:AXIS#2:SCREW:SPEED"]
	Driver["NC_LINK_ROOT:MACHINE:AXIS#2:MOTOR:CURRENT"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisCurrent(2, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[2][4][i-s] = dataPool[14][i%1000]
				}
				virtualDevice.Axes[2].Current = dataPool[14][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[2].Current
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[2][4][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[2][4][0:n], true
		},
	}

	Driver["NC_LINK_ROOT:MACHINE:AXIS#5:NAME"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisName(5, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#5:TYPE"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisType(5, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#5:NUMBER"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisNumber(5, item, virtualDevice)
		},
		nil,
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#5:MOTOR:SPEED"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisSpeed(5, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[5][0][i-s] = dataPool[25][i%1000]
				}
				virtualDevice.Axes[5].ActSpeed = dataPool[25][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[5].ActSpeed
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[5][0][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[5][0][0:n], true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#5:MOTOR:POSITION"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisPosition(5, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[5][1][i-s] = dataPool[26][i%1000]
				}
				virtualDevice.Axes[5].ActPosition = dataPool[26][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[5].ActPosition
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[5][1][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[5][1][0:n], true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:AXIS#5:SERVO_DRIVER:POSITION"] = Driver["NC_LINK_ROOT:MACHINE:AXIS#5:SCREW:POSITION"]
	Driver["NC_LINK_ROOT:MACHINE:AXIS#5:SERVO_DRIVER:SPEED"] = Driver["NC_LINK_ROOT:MACHINE:AXIS#5:MOTOR:SPEED"]
	Driver["NC_LINK_ROOT:MACHINE:AXIS#5:MOTOR:CURRENT"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			return funcGetAxisCurrent(5, item, virtualDevice)
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {

			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			switch status {
			case 0:
				s := virtualDevice.SampleHeartBeat % 1000
				for i := s; i < s+n; i++ {
					virtualDevice.SampleBuffer.Axis[5][4][i-s] = dataPool[29][i%1000]
				}
				virtualDevice.Axes[5].Current = dataPool[29][(s+n-1)%1000]
			case 1, 2:
				v := virtualDevice.Axes[5].Current
				for i := uint64(0); i < n; i++ {
					virtualDevice.SampleBuffer.Axis[5][4][i] = float32(v)
				}
			default:
			}
			return virtualDevice.SampleBuffer.Axis[5][4][0:n], true
		},
	}

	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:LINE_NUMBER"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					response.Value = append(response.Value, virtualDevice.RunState.Line)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			data := make([]uint32, n)
			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			var v uint64
			switch status {
			case 0:
				progNumber := (virtualDevice.SampleHeartBeat / 300000) % uint64(len(virtualDevice.ProgList))
				if virtualDevice.ProgList[progNumber].lineNumber == 0 {
					v = 0
				} else {
					v = (virtualDevice.SampleHeartBeat / 300) % uint64(virtualDevice.ProgList[progNumber].lineNumber)
				}
				virtualDevice.RunState.Line = v
			case 1:
				v = virtualDevice.RunState.Line
			default:
				v = 0
			}
			for i := uint64(0); i < n; i++ {
				data[i] = uint32(v)
			}
			return data, true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:PROGRAM_NUMBER"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					response.Value = append(response.Value, virtualDevice.RunState.ProgNumber)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			data := make([]uint32, n)
			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			var v uint64
			switch status {
			case 0:
				v = (virtualDevice.SampleHeartBeat / 300000) % uint64(len(virtualDevice.ProgList))
				virtualDevice.RunState.ProgNumber = int64(v)
			default:
				v = uint64(virtualDevice.RunState.ProgNumber)
			}
			for i := uint64(0); i < n; i++ {
				data[i] = uint32(v)
			}
			return data, true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:TOOL_NUMBER"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					response.Value = append(response.Value, virtualDevice.RunState.ToolNumber)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			data := make([]uint32, n)
			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			var v uint64
			switch status {
			case 0:
				v = virtualDevice.SampleHeartBeat / 30000
				virtualDevice.RunState.ToolNumber = int64(v)
			default:
				v = uint64(virtualDevice.RunState.ToolNumber)
			}
			for i := uint64(0); i < n; i++ {
				data[i] = uint32(v)
			}
			return data, true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:TOOL_PARAM"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					if item.Params == nil {
						response.Code = "NG"
						return response, false
					}
					response.Params = &QueryParameters{}
					response.Params.Operation = item.Params.Operation
					indexes := item.Params.Indexes
					if indexes == nil || len(indexes) == 0 {
						response.Code = "NG"
						return response, false
					}
					response.Params.Indexes = indexes
					for _, indexStr := range indexes {
						indexN := strings.Split(indexStr, "-")
						switch len(indexN) {
						case 1:
							index, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							if index < 0 || index >= len(virtualDevice.ToolParam) {
								response.Value = append(response.Value, nil)
							} else {
								response.Value = append(response.Value, virtualDevice.ToolParam[index])
							}
						case 2:
							index0, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							index1, err := strconv.Atoi(indexN[1])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							for index := index0; index <= index1; index++ {
								if index < 0 || index >= len(virtualDevice.ToolParam) {
									response.Value = append(response.Value, nil)
								} else {
									response.Value = append(response.Value, virtualDevice.ToolParam[index])
								}
							}
						default:
							response.Value = response.Value[0:0]
							response.Code = "NG"
							return response, false
						}
					}
					response.Code = "OK"
					return response, true
				}()
			case "get_length":
				return func() (*QueryResponseItem, bool) {
					response.Params = &QueryParameters{}
					response.Params.Operation = operation
					response.Value = append(response.Value, len(virtualDevice.ToolParam))
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		func(item SetRequestItem, virtualDevice *VirtualDevice) (item2 *SetResponseItem, b bool) {
			response := &SetResponseItem{}
			response.Id = item.Id
			if item.Params == nil {
				response.Code = "NG"
				return response, false
			}
			response.Params = &SetParameters{}
			response.Params.Operation = item.Params.Operation
			response.Params.Index = item.Params.Index
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "set_value"
			}
			switch operation {
			case "set_value":
				return func() (*SetResponseItem, bool) {
					index := item.Params.Index
					if index == nil || *index < 0 || *index >= len(virtualDevice.ToolParam) {
						response.Code = "NG"
						return response, false
					}
					if item.Params.Value == nil {
						response.Code = "NG"
						return response, false
					}
					value, ok := item.Params.Value.(map[string]interface{})
					if !ok {
						response.Code = "NG"
						return response, false
					}
					for k, v := range value {
						switch k {
						case "length":
							*virtualDevice.ToolParam[*index].Length = float32(v.(float64))
						case "radius":
							*virtualDevice.ToolParam[*index].Radius = float32(v.(float64))
						case "length_abrasion":
							*virtualDevice.ToolParam[*index].LengthAbrasion = float32(v.(float64))
						case "radius_abrasion":
							*virtualDevice.ToolParam[*index].RadiusAbrasion = float32(v.(float64))
						}
					}
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
	}
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:PROGRAM"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
				response.Params = &QueryParameters{}
				response.Params.Operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					if int64(len(virtualDevice.ProgList)) <= virtualDevice.RunState.ProgNumber {
						response.Value = append(response.Value, "")
					} else {
						response.Value = append(response.Value, virtualDevice.ProgList[virtualDevice.RunState.ProgNumber].name)
					}
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
		func(n uint64, virtualDevice *VirtualDevice) (interface{}, bool) {
			data := make([]string, n)
			var status int
			if len(virtualDevice.ProgList) == 0 {
				status = 1
			} else {
				status = statusN[int(virtualDevice.SampleHeartBeat/300000)%len(statusS)]
			}
			var v string
			if int64(len(virtualDevice.ProgList)) <= virtualDevice.RunState.ProgNumber {
				v = ""
			} else {
				v = virtualDevice.ProgList[virtualDevice.RunState.ProgNumber].name
			}
			switch status {
			case 0:
			default:
			}
			for i := uint64(0); i < n; i++ {
				data[i] = v
			}
			return data, true
		},
	}
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:SUBPROGRAM"] = Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:PROGRAM"]
	Driver["NC_LINK_ROOT:MACHINE:FILE"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					if item.Params == nil {
						response.Code = "NG"
						return response, false
					}
					response.Params = &QueryParameters{}
					response.Params.Operation = item.Params.Operation
					response.Params.Keys = item.Params.Keys
					response.Params.Offset = item.Params.Offset
					response.Params.Length = item.Params.Length
					keys := item.Params.Keys
					if len(keys) == 0 {
						response.Code = "NG"
						return response, false
					}
					offset := item.Params.Offset
					length := item.Params.Length
					if len(keys) == 0 || offset == nil || length == nil || *offset < 0 || *length <= 0 {
						response.Code = "NG"
						return response, false
					}
					for _, key := range keys {
						file, err := os.Open(virtualDevice.ProgDir + key)
						if err != nil {
							response.Value = append(response.Value, nil)
							fmt.Printf("Failed to Open File \"%s\"\n", virtualDevice.ProgDir+key)
							continue
						}
						fileinfo, err := file.Stat()
						if err != nil || fileinfo.IsDir() || fileinfo.Size() < int64(*offset) {
							response.Value = append(response.Value, nil)
							file.Close()
							continue
						}
						_offet, err := file.Seek(int64(*offset), 0)
						if int64(*offset) != _offet || err != nil {
							response.Value = append(response.Value, nil)
							file.Close()
							continue
						}
						data := make([]byte, *length, *length)
						n, err := file.Read(data)
						if err != nil || n <= 0 {
							response.Value = append(response.Value, nil)
							file.Close()
							continue
						}
						encodedData := make([]byte, n*2)
						for i := 0; i < n; i++ {
							c0 := (data[i] >> 4) & 0x0f
							c1 := data[i] & 0x0f
							if c0 > 9 {
								encodedData[2*i] = c0 - 10 + 'A'
							} else {
								encodedData[2*i] = c0 + '0'
							}
							if c1 > 9 {
								encodedData[2*i+1] = c1 - 10 + 'A'
							} else {
								encodedData[2*i+1] = c1 + '0'
							}
						}
						response.Value = append(response.Value, string(encodedData))
						file.Close()
					}
					response.Code = "OK"
					return response, true
				}()
			case "get_keys":
				return func() (*QueryResponseItem, bool) {
					prefixLen := len(virtualDevice.ProgDir)
					response.Params = &QueryParameters{}
					response.Params.Operation = "get_keys"
					var IterateDir func(dir string) []interface{}

					IterateDir = func(dir string) []interface{} {
						files := make([]interface{}, 0)
						fileInfos, err := ioutil.ReadDir(dir)
						if err != nil {
							return files
						}
						for _, fileInfo := range fileInfos {
							files = append(files, (dir + fileInfo.Name())[prefixLen:])
							if fileInfo.IsDir() {
								files = append(files, IterateDir(dir+fileInfo.Name())...)
							}
						}
						return files
					}
					files := IterateDir(virtualDevice.ProgDir)
					response.Value = files
					response.Code = "OK"
					return response, true
				}()
			case "get_attributes":
				return func() (*QueryResponseItem, bool) {
					response.Params = &QueryParameters{}
					response.Params.Keys = item.Params.Keys
					response.Params.Operation = "get_attributes"
					keys := item.Params.Keys
					if len(keys) == 0 {
						response.Code = "NG"
						return response, false
					}
					attributes := make([]interface{}, 0)
					for _, key := range keys {
						fileInfo, err := os.Stat(virtualDevice.ProgDir + key)
						if err != nil {
							attributes = append(attributes, nil)
							continue
						}
						attr := FileAttrib{}
						if fileInfo.IsDir() {
							attr.Type = "directory"
						} else {
							attr.Type = "file"
						}
						attr.Size = fileInfo.Size()
						t := fileInfo.ModTime()
						attr.ChangeTime = fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
						attributes = append(attributes, attr)
					}
					response.Code = "OK"
					response.Value = attributes
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		func(item SetRequestItem, virtualDevice *VirtualDevice) (*SetResponseItem, bool) {
			response := &SetResponseItem{}
			response.Id = item.Id
			response.Params = &SetParameters{}
			response.Params.Operation = item.Params.Operation
			response.Params.Key = item.Params.Key
			if virtualDevice == nil || item.Params == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "set_value"
			}
			switch operation {
			case "set_value":
				return func() (*SetResponseItem, bool) {
					key := item.Params.Key
					offset := item.Params.Offset
					length := item.Params.Length
					if item.Params.Value == nil {
						response.Code = "NG"
						return response, false
					}
					if len(key) == 0 || offset == nil || length == nil || *offset < 0 || *length <= 0 {
						response.Code = "NG"
						return response, false
					}
					value, ok := item.Params.Value.(string)
					if !ok {
						response.Code = "NG"
						return response, false
					}

					if len(value) != 2*(*length) {
						return nil, false
					}
					byteValue := []byte(value)

					for i := 0; i < len(byteValue); i++ {
						if byteValue[i] < '0' || byteValue[i] > 'F' || (byteValue[i] > '9' && byteValue[i] < 'A') {
							response.Code = "NG"
							return response, false
						}
					}
					file, err := os.OpenFile(virtualDevice.ProgDir+key, os.O_CREATE|os.O_RDWR, 0666)
					if err != nil {
						response.Code = "NG"
						return response, false
					}
					fileinfo, err := file.Stat()
					if err != nil || fileinfo.IsDir() {
						file.Close()
						response.Code = "NG"
						return response, false
					}
					if fileinfo.Size() < int64(*offset) {
						_, err := file.Seek(0, io.SeekEnd)
						if err != nil {
							file.Close()
							response.Code = "NG"
							return response, false
						}
						zeros := make([]byte, 1024)
						for j := int64(0); j < (int64(*offset)-fileinfo.Size())/1024; j++ {
							n, err := file.Write(zeros)
							if n != len(zeros) || err != nil {
								file.Close()
								response.Code = "NG"
								return response, false
							}
						}
						n := (int64(*offset) - fileinfo.Size()) % 1024
						if n != 0 {
							k, err := file.Write(zeros[0:n])
							if int64(k) != n || err != nil {
								file.Close()
								response.Code = "NG"
								return response, false
							}
						}
					}

					_, err = file.Seek(int64(*offset), io.SeekStart)
					if err != nil {
						file.Close()
						response.Code = "NG"
						return response, false
					}

					bytes := make([]byte, 1024)
					var j, k int
					for j = 0; j < *length/1024+1; j++ {
						for k = 0; k < 1024 && k < *length-j*1024; k++ {
							c0 := byteValue[2*(k+1024*j)]
							c1 := byteValue[2*(k+1024*j)+1]
							if c0 > '9' {
								c0 -= 'A'
								c0 += 10
							}
							if c1 > '9' {
								c1 -= 'A'
								c1 += 10
							}
							bytes[k] = ((c0 << 4) & 0xf0) | (c1 & 0x0f)
						}
						if k > 0 {
							n, err := file.Write(bytes[0:k])
							if err != nil || n != k {
								fmt.Printf("Faild to Write File \"%s\":%v\n", virtualDevice.ProgDir+key, err)
							}
						}
					}
					file.Close()
					response.Code = "OK"
					return response, true
				}()
			case "add":
				return func() (*SetResponseItem, bool) {
					response.Params.Key = item.Params.Key
					key := item.Params.Key
					err := os.Mkdir(virtualDevice.ProgDir+key, os.ModePerm)
					if err != nil {
						response.Code = "NG"
						return response, false
					}
					response.Code = "OK"
					return response, true
				}()
			case "delete":
				return func() (*SetResponseItem, bool) {
					key := item.Params.Key
					err := os.Remove(virtualDevice.ProgDir + key)
					if err != nil {
						response.Code = "NG"
						return response, false
					}
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
	}

	regDriver := func(reg int) *DriverItem {
		driver := &DriverItem{}
		driver.Get = func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					if item.Params == nil {
						response.Code = "NG"
						return response, false
					}
					response.Params = &QueryParameters{}
					response.Params.Operation = item.Params.Operation
					indexes := item.Params.Indexes
					if indexes == nil || len(indexes) == 0 {
						response.Code = "NG"
						return response, false
					}
					response.Params.Indexes = indexes
					for _, indexStr := range indexes {
						indexN := strings.Split(indexStr, "-")
						switch len(indexN) {
						case 1:
							index, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							if index < 0 || index >= len(virtualDevice.Regs[reg]) {
								response.Value = append(response.Value, nil)
							} else {
								response.Value = append(response.Value, virtualDevice.Regs[reg][index])
							}
						case 2:
							index0, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							index1, err := strconv.Atoi(indexN[1])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							for index := index0; index <= index1; index++ {
								if index < 0 || index >= len(virtualDevice.Regs[reg]) {
									response.Value = append(response.Value, nil)
								} else {
									response.Value = append(response.Value, virtualDevice.Regs[reg][index])
								}
							}
						default:
							response.Value = response.Value[0:0]
							response.Code = "NG"
							return response, false
						}
					}
					response.Code = "OK"
					return response, true
				}()
			case "get_length":
				return func() (*QueryResponseItem, bool) {
					response.Params = &QueryParameters{}
					response.Params.Operation = operation
					response.Value = append(response.Value, len(virtualDevice.Regs[reg]))
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		}
		driver.Set = func(item SetRequestItem, virtualDevice *VirtualDevice) (item2 *SetResponseItem, b bool) {
			response := &SetResponseItem{}
			response.Id = item.Id
			if item.Params == nil {
				response.Code = "NG"
				return response, false
			}
			response.Params = &SetParameters{}
			response.Params.Operation = item.Params.Operation
			response.Params.Index = item.Params.Index
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "set_value"
			}
			switch operation {
			case "set_value":
				return func() (*SetResponseItem, bool) {
					index := item.Params.Index
					if item.Params.Value == nil {
						response.Code = "NG"
						return response, false
					}
					if index == nil || *index < 0 || *index >= len(virtualDevice.Regs[reg]) {
						response.Code = "NG"
						return response, false
					}
					value, ok := item.Params.Value.(float64)
					if !ok {
						response.Code = "NG"
						return response, false
					}
					virtualDevice.Regs[reg][*index] = uint32(value)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		}
		driver.Sample = nil
		return driver
	}

	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#REG_X"] = regDriver(0)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#REG_Y"] = regDriver(1)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#REG_F"] = regDriver(2)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#REG_G"] = regDriver(3)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#REG_R"] = regDriver(4)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#REG_B"] = regDriver(5)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:PARAMETER"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					if item.Params == nil {
						response.Code = "NG"
						return response, false
					}
					response.Params = &QueryParameters{}
					response.Params.Operation = item.Params.Operation
					indexes := item.Params.Indexes
					if indexes == nil || len(indexes) == 0 {
						response.Code = "NG"
						return response, false
					}
					response.Params.Indexes = indexes
					for _, indexStr := range indexes {
						indexN := strings.Split(indexStr, "-")
						switch len(indexN) {
						case 1:
							index, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							response.Value = append(response.Value, ParamGetValue(index))
						case 2:
							index0, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							index1, err := strconv.Atoi(indexN[1])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							for index := index0; index <= index1; index++ {
								response.Value = append(response.Value, ParamGetValue(index))
							}
						default:
							response.Value = response.Value[0:0]
							response.Code = "NG"
							return response, false
						}
					}
					response.Code = "OK"
					return response, true
				}()
			case "get_length":
				return func() (*QueryResponseItem, bool) {
					response.Params = &QueryParameters{}
					response.Params.Operation = operation
					response.Value = append(response.Value, ParamGetLength())
					response.Code = "OK"
					return response, true
				}()
			case "get_attributes":
				return func() (*QueryResponseItem, bool) {
					if item.Params == nil {
						response.Code = "NG"
						return response, false
					}
					response.Params = &QueryParameters{}
					response.Params.Operation = item.Params.Operation
					indexes := item.Params.Indexes
					if indexes == nil || len(indexes) == 0 {
						response.Code = "NG"
						return response, false
					}
					response.Params.Indexes = indexes
					for _, indexStr := range indexes {
						indexN := strings.Split(indexStr, "-")
						switch len(indexN) {
						case 1:
							index, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							response.Value = append(response.Value, ParamGetAttribute(index))
						case 2:
							index0, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							index1, err := strconv.Atoi(indexN[1])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							for index := index0; index <= index1; index++ {
								response.Value = append(response.Value, ParamGetAttribute(index))
							}
						default:
							response.Value = response.Value[0:0]
							response.Code = "NG"
							return response, false
						}
					}
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		func(item SetRequestItem, virtualDevice *VirtualDevice) (item2 *SetResponseItem, b bool) {
			response := &SetResponseItem{}
			response.Id = item.Id
			if item.Params == nil {
				response.Code = "NG"
				return response, false
			}
			response.Params = &SetParameters{}
			response.Params.Operation = item.Params.Operation
			response.Params.Index = item.Params.Index
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "set_value"
			}
			switch operation {
			case "set_value":
				return func() (*SetResponseItem, bool) {
					index := item.Params.Index
					if index == nil || *index < 0 {
						response.Code = "NG"
						return response, false
					}
					if item.Params.Value == nil {
						response.Code = "NG"
						return response, false
					}
					value := item.Params.Value
					state := ParamSetValue(*index, value)
					if state {
						response.Code = "OK"
						return response, true
					} else {
						response.Code = "NG"
						return response, true
					}
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
	}
	dataDriver := func(data int) *DriverItem {
		driver := &DriverItem{}
		driver.Get = func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					if item.Params == nil {
						response.Code = "NG"
						return response, false
					}
					response.Params = &QueryParameters{}
					response.Params.Operation = item.Params.Operation
					indexes := item.Params.Indexes
					if indexes == nil || len(indexes) == 0 {
						response.Code = "NG"
						return response, false
					}
					response.Params.Indexes = indexes
					for _, indexStr := range indexes {
						indexN := strings.Split(indexStr, "-")
						switch len(indexN) {
						case 1:
							index, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							if index < 0 || index >= len(AxisChanSystemData[data]) {
								response.Value = append(response.Value, nil)
							} else {
								response.Value = append(response.Value, AxisChanSystemData[data][index])
							}
						case 2:
							index0, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							index1, err := strconv.Atoi(indexN[1])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							for index := index0; index <= index1; index++ {
								if index < 0 || index >= len(AxisChanSystemData[data]) {
									response.Value = append(response.Value, nil)
								} else {
									response.Value = append(response.Value, AxisChanSystemData[data][index])
								}
							}
						default:
							response.Value = response.Value[0:0]
							response.Code = "NG"
							return response, false
						}
					}
					response.Code = "OK"
					return response, true
				}()
			case "get_length":
				return func() (*QueryResponseItem, bool) {
					response.Params = &QueryParameters{}
					response.Params.Operation = operation
					response.Value = append(response.Value, len(AxisChanSystemData[data]))
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		}
		driver.Set = func(item SetRequestItem, virtualDevice *VirtualDevice) (item2 *SetResponseItem, b bool) {
			response := &SetResponseItem{}
			response.Id = item.Id
			if item.Params == nil {
				response.Code = "NG"
				return response, false
			}
			response.Params = &SetParameters{}
			response.Params.Operation = item.Params.Operation
			response.Params.Index = item.Params.Index
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "set_value"
			}
			switch operation {
			case "set_value":
				return func() (*SetResponseItem, bool) {
					index := item.Params.Index
					if item.Params.Value == nil {
						response.Code = "NG"
						return response, false
					}
					if index == nil || *index < 0 || *index >= len(AxisChanSystemData[data]) {
						response.Code = "NG"
						return response, false
					}
					value, ok := item.Params.Value.(float64)
					if !ok {
						response.Code = "NG"
						return response, false
					}
					AxisChanSystemData[data][*index] = uint32(value)
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		}
		driver.Sample = nil
		return driver
	}

	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#AXIS_0"] = dataDriver(0)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#AXIS_1"] = dataDriver(1)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#AXIS_2"] = dataDriver(2)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#AXIS_5"] = dataDriver(5)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#CHAN_0"] = dataDriver(6)
	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:VARIABLE#SYS"] = dataDriver(10)

	Driver["NC_LINK_ROOT:MACHINE:CONTROLLER:COORDINATE"] = &DriverItem{
		func(item QueryRequestItem, virtualDevice *VirtualDevice) (*QueryResponseItem, bool) {
			response := &QueryResponseItem{}
			response.Id = item.Id
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params != nil && item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "get_value"
			}
			switch operation {
			case "get_value":
				return func() (*QueryResponseItem, bool) {
					if item.Params == nil {
						response.Code = "NG"
						return response, false
					}
					response.Params = &QueryParameters{}
					response.Params.Operation = item.Params.Operation
					indexes := item.Params.Indexes
					if indexes == nil || len(indexes) == 0 {
						response.Code = "NG"
						return response, false
					}
					response.Params.Indexes = indexes
					for _, indexStr := range indexes {
						indexN := strings.Split(indexStr, "-")
						switch len(indexN) {
						case 1:
							index, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							if index < 0 || index >= len(virtualDevice.Crds) {
								response.Value = append(response.Value, nil)
							} else {
								response.Value = append(response.Value, virtualDevice.Crds[index])
							}
						case 2:
							index0, err := strconv.Atoi(indexN[0])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							index1, err := strconv.Atoi(indexN[1])
							if err != nil {
								response.Value = response.Value[0:0]
								response.Code = "NG"
								return response, false
							}
							for index := index0; index <= index1; index++ {
								if index < 0 || index >= len(virtualDevice.Crds) {
									response.Value = append(response.Value, nil)
								} else {
									response.Value = append(response.Value, virtualDevice.Crds[index])
								}
							}
						default:
							response.Value = response.Value[0:0]
							response.Code = "NG"
							return response, false
						}
					}
					response.Code = "OK"
					return response, true
				}()
			case "get_length":
				return func() (*QueryResponseItem, bool) {
					response.Params = &QueryParameters{}
					response.Params.Operation = operation
					response.Value = append(response.Value, len(virtualDevice.Crds))
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		func(item SetRequestItem, virtualDevice *VirtualDevice) (item2 *SetResponseItem, b bool) {
			response := &SetResponseItem{}
			response.Id = item.Id
			if item.Params == nil {
				response.Code = "NG"
				return response, false
			}
			response.Params = &SetParameters{}
			response.Params.Operation = item.Params.Operation
			response.Params.Index = item.Params.Index
			if virtualDevice == nil {
				response.Code = "NG"
				return response, false
			}
			var operation string
			if item.Params.Operation != "" {
				operation = item.Params.Operation
			} else {
				operation = "set_value"
			}
			switch operation {
			case "set_value":
				return func() (*SetResponseItem, bool) {
					index := item.Params.Index
					if index == nil || *index < 0 || *index >= len(virtualDevice.Crds) {
						response.Code = "NG"
						return response, false
					}
					if item.Params.Value == nil {
						response.Code = "NG"
						return response, false
					}
					value, ok := item.Params.Value.(map[string]interface{})
					if !ok {
						response.Code = "NG"
						return response, false
					}
					for k, v := range value {
						switch k {
						case "x":
							*virtualDevice.Crds[*index].X = float32(v.(float64))
						case "y":
							*virtualDevice.Crds[*index].Y = float32(v.(float64))
						case "z":
							*virtualDevice.Crds[*index].Z = float32(v.(float64))
						case "c":
							*virtualDevice.Crds[*index].C = float32(v.(float64))
						}
					}
					response.Code = "OK"
					return response, true
				}()
			default:
				response.Code = "NG"
				return response, false
			}
		},
		nil,
	}
}
