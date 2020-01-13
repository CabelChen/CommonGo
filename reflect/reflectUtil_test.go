package reflect

import (
	. "github.com/smartystreets/goconvey/convey"
	"strconv"
	"testing"
)

type OuterStruct struct {
	Field1 string
	Field2 int
	Field3 float64
	Field4 EmbeddedStruct
	Field5 *EmbeddedStruct
}

type EmbeddedStruct struct {
	EmbeddedField1 bool
}

func TestReflectUtil(t *testing.T) {
	Convey("测试map映射结构体，字段类型相同", t, func() {
		outerStructMap := make(map[string]interface{})
		outerStructMap["Field1"] = "Test String!!"
		outerStructMap["Field2"] = 1
		outerStructMap["Field3"] = 1.0

		embeddedStructMap := make(map[string]interface{})
		embeddedStructMap["EmbeddedField1"] = true
		outerStructMap["Field4"] = embeddedStructMap
		outerStructMap["Field5"] = embeddedStructMap

		outerStruct := &OuterStruct{}
		Map2Struct(outerStruct, outerStructMap)

		So(outerStruct.Field1, ShouldEqual, outerStructMap["Field1"])
		So(outerStruct.Field2, ShouldEqual, outerStructMap["Field2"])
		So(outerStruct.Field3, ShouldEqual, outerStructMap["Field3"])
		So(outerStruct.Field4.EmbeddedField1, ShouldEqual, outerStructMap["Field4"].(map[string]interface{})["EmbeddedField1"])
		So(outerStruct.Field5.EmbeddedField1, ShouldEqual, outerStructMap["Field5"].(map[string]interface{})["EmbeddedField1"])
	})
	Convey("测试map映射结构体,字段类型不同,int -> string", t, func() {
		outerStructMap := make(map[string]interface{})
		outerStructMap["Field1"] = 123

		outerStruct := &OuterStruct{}
		Map2Struct(outerStruct, outerStructMap)

		So(outerStruct.Field1, ShouldEqual, strconv.Itoa(outerStructMap["Field1"].(int)))
	})

	Convey("测试map映射结构体,字段类型不同,int64 -> int", t, func() {
		outerStructMap := make(map[string]interface{})
		outerStructMap["Field2"] = int64(123)

		outerStruct := &OuterStruct{}
		Map2Struct(outerStruct, outerStructMap)

		So(outerStruct.Field2, ShouldEqual, outerStructMap["Field2"].(int64))
	})

	Convey("测试map映射结构体,字段类型不同,string -> int", t, func() {
		outerStructMap := make(map[string]interface{})
		outerStructMap["Field2"] = "123"

		outerStruct := &OuterStruct{}
		Map2Struct(outerStruct, outerStructMap)
		intVal, err := strconv.Atoi(outerStructMap["Field2"].(string))
		if err != nil {
			panic(err)
		}
		So(outerStruct.Field2, ShouldEqual, intVal)
	})

	Convey("测试map映射结构体,字段类型不同,string -> float", t, func() {
		outerStructMap := make(map[string]interface{})
		outerStructMap["Field3"] = "123"

		outerStruct := &OuterStruct{}
		Map2Struct(outerStruct, outerStructMap)
		floatVar, err := strconv.ParseFloat(outerStructMap["Field3"].(string), 64)
		if err != nil {
			panic(err)
		}
		So(outerStruct.Field3, ShouldEqual, floatVar)
	})

	Convey("测试map映射结构体,字段类型不同,float -> string", t, func() {
		outerStructMap := make(map[string]interface{})
		outerStructMap["Field1"] = 25.01

		outerStruct := &OuterStruct{}
		Map2Struct(outerStruct, outerStructMap)

		So(outerStruct.Field1, ShouldEqual, strconv.FormatFloat(outerStructMap["Field1"].(float64), 'E', -1, 64))
	})
}
