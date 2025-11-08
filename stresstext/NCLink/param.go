package NCLink

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	FileHeadSize      = 180
	ParamFileHeadSize = 32
	ParamValueSize    = 8
)
const (
	NCUNum   = 1
	MACNum   = 1
	CHANNum  = 4
	AXESNum  = 32
	ACMPNum  = AXESNum
	CFGNum   = 80
	TABLENum = 1
)
const (
	NCUParamNum   = 2000
	MACParamNum   = 2000
	CHANParamNum  = 1000
	AXESParamNum  = 1000
	ACMPParamNum  = 500
	CFGParamNum   = 32
	TABLEParamNum = 20000
)

type RawFileHead [FileHeadSize]byte
type RawParamFileHead [ParamFileHeadSize]byte
type RawParamValue [ParamValueSize]byte
type FileHead struct {
	fileFlag     [4]byte
	version      int32
	fileInfoAddr int32
	fileDataAddr int32
	verifyType   int32
	verifyCode   [8]int32
	reserve      [32]int32
}
type ParamFileHead struct {
	headInfoSize uint32
	majorVersion uint32
	minorVersion uint32
	dataSize     uint32
	chkSum       uint32
	userVer      int32
	reserved     [2]uint32
}
type ParamValue struct {
	id        int
	attribute ParamAttribute
	value     interface{}
	raw       [ParamValueSize]byte
}
type ParamSubLib struct {
	paramItems []ParamValue
}

var paramItems [][][]ParamValue
var paramDefine ParamDefine

const (
	NCU = iota
	MAC
	CHAN
	AXES
	ACMP
	CFG
	TABLE
)

type ParamManager struct {
	NCU   [NCUParamNum]ParamValue
	MAC   [MACParamNum]ParamValue
	CHAN  [4 * CHANParamNum]ParamValue
	AXES  [32 * AXESParamNum]ParamValue
	ACMP  [ACMPParamNum]ParamValue
	CFG   [80 * CFGParamNum]ParamValue
	TABLE [TABLEParamNum]ParamValue
}

type ParamDefine struct {
	XMLName xml.Name
	Version string `xml:"version,attr"`
	Lib     []Lib  `xml:"LIB"`
}
type Lib struct {
	XMLName xml.Name
	Type    string `xml:"type,attr"`
	Name    string `xml:"name,attr"`
	SubName string `xml:"subname,attr"`
	StartId string `xml:"startid,attr"`
	ParmNum string `xml:"parmnum,attr"`
	SubNum  string `xml:"subnum,attr"`
	Item    []Item `xml:"item"`
}
type Item struct {
	XMLName   xml.Name
	Id        string   `xml:"id,attr"`
	Name      string   `xml:"name,attr"`
	Dimension string   `xml:"dimension,attr"`
	Right     string   `xml:"right,attr"`
	ActType   string   `xml:"acttype,attr"`
	StoreType string   `xml:"storetype,attr"`
	Default   string   `xml:"def,attr"`
	Min       string   `xml:"min,attr"`
	Max       string   `xml:"max,attr"`
	SubLib    []SubLib `xml:"sublib"`
}
type SubLib struct {
	XMLName xml.Name
	Index   string `xml:"index,attr"`
	Name    string `xml:"name,attr"`
	Value   string `xml:"value,attr"`
	Item    []Item `xml:"item"`
}
type ParamAttribute struct {
	Name    string      `json:"name"`
	Max     interface{} `json:"max_value,omitempty"`
	Min     interface{} `json:"min_value,omitempty"`
	Default interface{} `json:"default_value,omitempty"`
	Act     int         `json:"action_type,omitempty"`
}

func ParamGetValue(id int) interface{} {
	switch {
	case id >= 0 && id < 10000:
		if id >= NCUParamNum {
			return nil
		} else {
			return paramItems[0][0][id].value
		}
	case id >= 10000 && id < 40000:
		id %= 10000
		if id >= MACParamNum {
			return nil
		} else {
			return paramItems[1][0][id].value
		}
	case id >= 40000 && id < 100000:
		id %= 40000
		ch := id / 1000
		if ch >= CHANNum {
			return nil
		}
		id %= 1000
		if id >= CHANParamNum {
			return nil
		} else {
			return paramItems[2][ch][id].value
		}
	case id >= 100000 && id < 300000:
		id %= 100000
		ax := id / 1000
		if ax >= AXESNum {
			return nil
		}
		id %= 1000
		if id >= AXESParamNum {
			return nil
		} else {
			return paramItems[2][ax][id].value
		}
	case id >= 300000 && id < 500000:
		id %= 300000
		acmp := id / 1000
		if acmp >= ACMPNum {
			return nil
		}
		id %= 1000
		if id >= ACMPParamNum {
			return nil
		} else {
			return paramItems[2][acmp][id].value
		}
	case id >= 500000 && id < 700000:
		id %= 500000
		cfg := id / 1000
		if cfg >= CFGNum {
			return nil
		}
		id %= 1000
		if id >= CFGParamNum {
			return nil
		} else {
			return paramItems[2][cfg][id].value
		}
	}
	return nil
}
func ParamGetAttribute(id int) interface{} {
	switch {
	case id >= 0 && id < 10000:
		if id >= NCUParamNum || paramItems[0][0][id].value == nil {
			return nil
		} else {
			return paramItems[0][0][id].attribute
		}
	case id >= 10000 && id < 40000:
		id %= 10000
		if id >= MACParamNum || paramItems[1][0][id].value == nil {
			return nil
		} else {
			return paramItems[1][0][id].attribute
		}
	case id >= 40000 && id < 100000:
		id %= 40000
		ch := id / 1000
		if ch >= CHANNum {
			return nil
		}
		id %= 1000
		if id >= CHANParamNum || paramItems[2][ch][id].value == nil {
			return nil
		} else {
			return paramItems[2][ch][id].attribute
		}
	case id >= 100000 && id < 300000:
		id %= 100000
		ax := id / 1000
		if ax >= AXESNum {
			return nil
		}
		id %= 1000
		if id >= AXESParamNum || paramItems[2][ax][id].value == nil {
			return nil
		} else {
			return paramItems[2][ax][id].attribute
		}
	case id >= 300000 && id < 500000:
		id %= 300000
		acmp := id / 1000
		if acmp >= ACMPNum {
			return nil
		}
		id %= 1000
		if id >= ACMPParamNum || paramItems[2][acmp][id].value == nil {
			return nil
		} else {
			return paramItems[2][acmp][id].attribute
		}
	case id >= 500000 && id < 700000:
		id %= 500000
		cfg := id / 1000
		if cfg >= CFGNum {
			return nil
		}
		id %= 1000
		if id >= CFGParamNum || paramItems[2][cfg][id].value == nil {
			return nil
		} else {
			return paramItems[2][cfg][id].attribute
		}
	}
	return nil
}
func ParamGetLength() int {
	return 700000
}
func ParamSetValue(id int, value interface{}) bool {
	if value == nil {
		return false
	}
	switch {
	case id >= 0 && id < 10000:
		if id >= NCUParamNum {
			return false
		} else {
			paramItems[0][0][id].value = value
			return true
		}
	case id >= 10000 && id < 40000:
		id %= 10000
		if id >= MACParamNum {
			return false
		} else {
			paramItems[0][0][id].value = value
			return true
		}
	case id >= 40000 && id < 100000:
		id %= 40000
		ch := id / 1000
		if ch >= CHANNum {
			return false
		}
		id %= 1000
		if id >= CHANParamNum {
			return false
		} else {
			paramItems[2][ch][id].value = value
			return true
		}
	case id >= 100000 && id < 300000:
		id %= 100000
		ax := id / 1000
		if ax >= AXESNum {
			return false
		}
		id %= 1000
		if id >= AXESParamNum {
			return false
		} else {
			paramItems[2][ax][id].value = value
			return true
		}
	case id >= 300000 && id < 500000:
		id %= 300000
		acmp := id / 1000
		if acmp >= ACMPNum {
			return false
		}
		id %= 1000
		if id >= ACMPParamNum {
			return false
		} else {
			paramItems[2][acmp][id].value = value
			return true
		}
	case id >= 500000 && id < 700000:
		id %= 500000
		cfg := id / 1000
		if cfg >= CFGNum {
			return false
		}
		id %= 1000
		if id >= CFGParamNum {
			return false
		} else {
			paramItems[2][cfg][id].value = false
			return true
		}
	}
	return false
}
func ParamInit() bool {
	paramDataFile := "param/818BM.DAT"
	paramDefineFile := "param/PARM-CN.XML"
	dataFile, err := os.Open(paramDataFile)
	if err != nil {
		fmt.Printf("Failed to Open File:%v", err.Error())
		return false
	}
	defer dataFile.Close()
	fileHeadBuffer := make([]byte, FileHeadSize)
	n, err := dataFile.Read(fileHeadBuffer)
	if err != nil {
		fmt.Printf("Failed to Read File Head:%v", err.Error())
		return false
	}
	if n != FileHeadSize {
		fmt.Printf("Failed to Read File Head")
		return false
	}
	paramFileHeadBuffer := make([]byte, ParamFileHeadSize)
	n, err = dataFile.Read(paramFileHeadBuffer)
	if err != nil {
		fmt.Printf("Failed to Read Param File Head:%v", err.Error())
		return false
	}
	if n != ParamFileHeadSize {
		fmt.Printf("Failed to Read File Head")
		return false
	}
	paramItems = make([][][]ParamValue, 7)
	paramItems[NCU] = make([][]ParamValue, NCUNum)
	for i := 0; i < NCUNum; i++ {
		paramItems[NCU][i] = make([]ParamValue, NCUParamNum)
	}
	paramItems[MAC] = make([][]ParamValue, MACNum)
	for i := 0; i < MACNum; i++ {
		paramItems[MAC][i] = make([]ParamValue, MACParamNum)
	}
	paramItems[CHAN] = make([][]ParamValue, CHANNum)
	for i := 0; i < CHANNum; i++ {
		paramItems[CHAN][i] = make([]ParamValue, CHANParamNum)
	}
	paramItems[AXES] = make([][]ParamValue, AXESNum)
	for i := 0; i < AXESNum; i++ {
		paramItems[AXES][i] = make([]ParamValue, AXESParamNum)
	}
	paramItems[ACMP] = make([][]ParamValue, ACMPNum)
	for i := 0; i < ACMPNum; i++ {
		paramItems[ACMP][i] = make([]ParamValue, ACMPParamNum)
	}
	paramItems[CFG] = make([][]ParamValue, CFGNum)
	for i := 0; i < CFGNum; i++ {
		paramItems[CFG][i] = make([]ParamValue, CFGParamNum)
	}
	paramItems[TABLE] = make([][]ParamValue, TABLENum)
	for i := 0; i < TABLENum; i++ {
		paramItems[TABLE][i] = make([]ParamValue, TABLEParamNum)
	}
	for i := 0; i < len(paramItems); i++ {
		for j := 0; j < len(paramItems[i]); j++ {
			for k := 0; k < len(paramItems[i][j]); k++ {
				n, err = dataFile.Read(paramItems[i][j][k].raw[:])
				if err != nil {
					fmt.Printf("Failed to Read Param Value:%v", err.Error())
					return false
				}
				if n != ParamValueSize {
					fmt.Printf("Failed to Read File")
					return false
				}
			}
		}
	}
	data, err := ioutil.ReadFile(paramDefineFile)
	if err != nil {
		fmt.Printf("Failed to Read Param Define File:%v", err.Error())
		return false
	}
	err = xml.Unmarshal(data, &paramDefine)
	if err != nil {
		fmt.Printf("Failed to Parse Param Define File:%v", err)
	}
	for i := 0; i < len(paramDefine.Lib); i++ {
		if i >= len(paramItems) {
			continue
		}
		lib := paramDefine.Lib[i]
		libParamItems := paramItems[i]
		if lib.Item == nil {
			continue
		}
		startId, err := strconv.Atoi(lib.StartId)
		if err != nil {
			continue
		}
		for j := 0; j < len(paramDefine.Lib[i].Item); j++ {
			item := paramDefine.Lib[i].Item[j]
			dimension, err := strconv.Atoi(item.Dimension)
			if dimension > 0 {
				continue
			}
			index, err := strconv.Atoi(item.Id)
			if err != nil {
				continue
			}
			index -= startId
			if index < 0 || index >= len(libParamItems[0]) {
				continue
			}
			var m, n int
			switch i {
			case NCU:
				m = NCUNum
				n = NCUParamNum
			case MAC:
				m = MACNum
				n = MACParamNum
			case CHAN:
				m = CHANNum
				n = CHANParamNum
			case AXES:
				m = AXESNum
				n = AXESParamNum
			case ACMP:
				m = ACMPNum
				n = ACMPParamNum
			case CFG:
				m = CFGNum
				n = CFGParamNum
			case TABLE:
				m = TABLENum
				n = TABLEParamNum
			}
			if index >= n {
				continue
			}
			for u := 0; u < m; u++ {
				paramItem := &libParamItems[u][index]
				paramItem.id = startId + index + u*1000
				paramItem.attribute.Name = item.Name
				switch item.ActType {
				case "ACT_SAVE":
					paramItem.attribute.Act = 0
				case "ACT_NOW":
					paramItem.attribute.Act = 1
				case "ACT_RST":
					paramItem.attribute.Act = 2
				case "ACT_PWR":
					paramItem.attribute.Act = 3
				case "ACT_HIDE":
					paramItem.attribute.Act = 4
				}
				byteReader := bytes.NewReader(paramItem.raw[:])
				if strings.HasPrefix(item.StoreType, "STRING") {
					w := 0
					for ; w < 8; w++ {
						if paramItem.raw[w] == 0 {
							break
						}
					}
					paramItem.value = string(paramItem.raw[0:w])
				} else {
					switch item.StoreType {
					case "BOOL", "UINT1":
						paramItem.value = new(uint8)
						if len(item.Default) != 0 {
							paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
						}
						if len(item.Max) != 0 {
							paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
						}
						if len(item.Min) != 0 {
							paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
						}
					case "INT1":
						paramItem.value = new(int8)
						if len(item.Default) != 0 {
							paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
						}
						if len(item.Max) != 0 {
							paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
						}
						if len(item.Min) != 0 {
							paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
						}
					case "UINT2":
						paramItem.value = new(uint16)
						if len(item.Default) != 0 {
							paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
						}
						if len(item.Max) != 0 {
							paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
						}
						if len(item.Min) != 0 {
							paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
						}
					case "INT2":
						paramItem.value = new(int16)
						if len(item.Default) != 0 {
							paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
						}
						if len(item.Max) != 0 {
							paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
						}
						if len(item.Min) != 0 {
							paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
						}
					case "UINT4":
						paramItem.value = new(uint32)
						if len(item.Default) != 0 {
							paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
						}
						if len(item.Max) != 0 {
							paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
						}
						if len(item.Min) != 0 {
							paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
						}
					case "INT4", "HEX4":
						paramItem.value = new(int32)
						if len(item.Default) != 0 {
							paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
						}
						if len(item.Max) != 0 {
							paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
						}
						if len(item.Min) != 0 {
							paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
						}
					case "REAL", "INTEXP":
						paramItem.value = new(float64)
						if len(item.Default) != 0 {
							paramItem.attribute.Default, _ = strconv.ParseFloat(item.Default, 64)
						}
						if len(item.Max) != 0 {
							paramItem.attribute.Default, _ = strconv.ParseFloat(item.Default, 64)
						}
						if len(item.Min) != 0 {
							paramItem.attribute.Default, _ = strconv.ParseFloat(item.Default, 64)
						}
					case "BYTE[8]", "BYTE[4]":
						paramItem.value = new([8]byte)
					}
					err = binary.Read(byteReader, binary.LittleEndian, paramItem.value)
					if err != nil {
						continue
					}
				}
			}
		}
	}

	for i := 0; i < len(paramDefine.Lib); i++ {
		if i >= len(paramItems) {
			break
		}
		lib := paramDefine.Lib[i]
		if lib.Item == nil {
			continue
		}
		startId, err := strconv.Atoi(lib.StartId)
		if err != nil {
			continue
		}
		if i != AXES && i != CFG {
			continue
		}
		for j := 0; j < len(paramDefine.Lib[i].Item); j++ {
			item := paramDefine.Lib[i].Item[j]
			dimension, _ := strconv.Atoi(item.Dimension)
			if dimension <= 0 || item.SubLib == nil || len(item.SubLib) == 0 {
				continue
			}

			for k := 0; k < len(item.SubLib); k++ {
				if len(item.SubLib[k].Value) == 0 {
					continue
				}
				subItems := item.SubLib[k].Item
				if subItems == nil {
					continue
				}
				subValue, err := strconv.Atoi(item.SubLib[k].Value)
				if err != nil {
					continue
				}
				var subParamItems []ParamValue
				var num, target int
				if i == AXES {
					num = AXESNum
					for l := 0; l < num; l++ {
						typeValue := paramItems[i][l][1].value
						if typeValue == nil {
							continue
						}
						typeNumber := *typeValue.(*uint8)
						if subValue == int(typeNumber) {
							subParamItems = paramItems[i][l]
							target = l
							for l := 0; l < len(subItems); l++ {
								subItem := subItems[l]
								dimension, _ := strconv.Atoi(subItem.Dimension)
								if dimension > 0 {
									continue
								}
								index, err := strconv.Atoi(subItem.Id)
								if err != nil {
									continue
								}
								index -= startId
								if index < 0 || index >= len(subParamItems) {
									continue
								}
								paramItem := &subParamItems[index]
								paramItem.id = startId + index + target*1000
								paramItem.attribute.Name = item.Name
								switch item.ActType {
								case "ACT_SAVE":
									paramItem.attribute.Act = 0
								case "ACT_NOW":
									paramItem.attribute.Act = 1
								case "ACT_RST":
									paramItem.attribute.Act = 2
								case "ACT_PWR":
									paramItem.attribute.Act = 3
								case "ACT_HIDE":
									paramItem.attribute.Act = 4
								}
								byteReader := bytes.NewReader(paramItem.raw[:])
								if strings.HasPrefix(item.StoreType, "STRING") {
									w := 0
									for ; w < 8; w++ {
										if paramItem.raw[w] == 0 {
											break
										}
									}
									paramItem.value = string(paramItem.raw[0:w])
								} else {
									switch item.StoreType {
									case "BOOL", "UINT1":
										paramItem.value = new(uint8)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}

									case "INT1":
										paramItem.value = new(int8)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "UINT2":
										paramItem.value = new(uint16)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "INT2":
										paramItem.value = new(int16)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "UINT4":
										paramItem.value = new(uint32)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "INT4", "HEX4":
										paramItem.value = new(int32)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "REAL", "INTEXP":
										paramItem.value = new(float64)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.ParseFloat(item.Default, 64)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Default, _ = strconv.ParseFloat(item.Default, 64)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Default, _ = strconv.ParseFloat(item.Default, 64)
										}
									case "BYTE":
										paramItem.value = new([8]byte)
									}
									err = binary.Read(byteReader, binary.LittleEndian, paramItem.value)
									if err != nil {
										continue
									}
								}
							}
						}
					}
				} else if i == CFG {
					num = CFGNum
					for l := 0; l < num; l++ {
						typeValue := paramItems[i][l][2].value
						if typeValue == nil {
							continue
						}
						typeNumber := *typeValue.(*int32)
						if subValue == int(typeNumber) {
							subParamItems = paramItems[i][l]
							target = l
							for l := 0; l < len(subItems); l++ {
								subItem := subItems[l]
								dimension, _ := strconv.Atoi(subItem.Dimension)
								if dimension > 0 {
									continue
								}
								index, err := strconv.Atoi(subItem.Id)
								if err != nil {
									continue
								}
								index -= startId
								if index < 0 || index >= len(subParamItems) {
									continue
								}
								paramItem := &subParamItems[index]
								paramItem.id = startId + index + target*1000
								paramItem.attribute.Name = item.Name
								switch item.ActType {
								case "ACT_SAVE":
									paramItem.attribute.Act = 0
								case "ACT_NOW":
									paramItem.attribute.Act = 1
								case "ACT_RST":
									paramItem.attribute.Act = 2
								case "ACT_PWR":
									paramItem.attribute.Act = 3
								case "ACT_HIDE":
									paramItem.attribute.Act = 4
								}
								byteReader := bytes.NewReader(paramItem.raw[:])
								if strings.HasPrefix(item.StoreType, "STRING") {
									w := 0
									for ; w < 8; w++ {
										if paramItem.raw[w] == 0 {
											break
										}
									}
									paramItem.value = string(paramItem.raw[0:w])
								} else {
									switch item.StoreType {
									case "BOOL", "UINT1":
										paramItem.value = new(uint8)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "INT1":
										paramItem.value = new(int8)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "UINT2":
										paramItem.value = new(uint16)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "INT2":
										paramItem.value = new(int16)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "UINT4":
										paramItem.value = new(uint32)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "INT4", "HEX4":
										paramItem.value = new(int32)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.Atoi(item.Default)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Max, _ = strconv.Atoi(item.Max)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Min, _ = strconv.Atoi(item.Min)
										}
									case "REAL", "INTEXP":
										paramItem.value = new(float64)
										if len(item.Default) != 0 {
											paramItem.attribute.Default, _ = strconv.ParseFloat(item.Default, 64)
										}
										if len(item.Max) != 0 {
											paramItem.attribute.Default, _ = strconv.ParseFloat(item.Default, 64)
										}
										if len(item.Min) != 0 {
											paramItem.attribute.Default, _ = strconv.ParseFloat(item.Default, 64)
										}
									case "BYTE":
										paramItem.value = new([8]byte)
									}
									err = binary.Read(byteReader, binary.LittleEndian, paramItem.value)
									if err != nil {
										continue
									}
								}
							}
						}
					}
				} else {
					continue
				}
			}
		}
	}
	return true
}
