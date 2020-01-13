package reflect

import (
	"reflect"
	"strconv"
)

/*
类型反转换规则：

如果是解析到一个指针，Unmarshal首先会检查是否是json字符null，如果是的话指针会设置为nil，其他情况Unmarshal会把值填充到指针所指的位置，如果是一个空指针则会分配一个新的位置。
如果是解析到struct，Unmarshal会用struct的字段名或者`json`标签指定的名字和json的key相匹配。
如果是解析到interface值，转化规则如下：
1.bool对应JSON的boolean
2.float64对应JSON的number，
3.string对应JSON的string，
4.[]interface{}对应JSON的array
5.map[string]interface{}对应JSON的object
6.nil对应JSON的null。
*/
func Map2Struct(inStructPtr interface{}, filedMap map[string]interface{}) {

	getType := reflect.TypeOf(inStructPtr)
	getValue := reflect.ValueOf(inStructPtr)
	if getType.Kind() == reflect.Ptr {
		// 传入的inStructPtr是指针，需要.Elem()取得指针指向的value
		getType = getType.Elem()
		getValue = getValue.Elem()
	} else {
		panic("inStructPtr must be ptr to struct")
	}

	for i := 0; i < getType.NumField(); i++ {
		fieldType := getType.Field(i)
		key := fieldType.Name
		fillValue := filedMap[key]
		fieldValue := getValue.FieldByName(key)

		if fillValue == nil || IsNil(reflect.ValueOf(fillValue)) {
			continue
		}

		if fieldValue.Kind() == reflect.ValueOf(fillValue).Kind() {
			if fieldValue.Kind() == reflect.Slice {
				fieldValue.Set(FillSlice(reflect.ValueOf(fillValue), fieldValue.Type().Elem().Kind()))
			} else if fieldValue.Kind() == reflect.Map {
				fieldValue.Set(FillMap(reflect.ValueOf(fillValue), fieldValue.Type().Elem().Kind()))
			} else {
				fieldValue.Set(reflect.ValueOf(fillValue))
			}
		} else if fieldValue.Kind() != reflect.ValueOf(fillValue).Kind() {
			if fieldValue.Kind() == reflect.Struct && reflect.ValueOf(fillValue).Kind() == reflect.Map {
				instance := reflect.New(fieldValue.Type())
				Map2Struct(instance.Interface(), fillValue.(map[string]interface{}))
				fieldValue.Set(instance.Elem())
			} else if fieldValue.Kind() == reflect.Ptr && reflect.ValueOf(fillValue).Kind() == reflect.Map {
				instance := reflect.New(fieldValue.Type().Elem())
				Map2Struct(instance.Interface(), fillValue.(map[string]interface{}))
				fieldValue.Set(instance)
			} else {
				fieldValue = TypeCompatibility(fieldValue, reflect.ValueOf(fillValue)) //类型兼容
			}
		}

	}
}

/**
  根据sliceType填充slice类型
*/
func FillSlice(value reflect.Value, sliceType reflect.Kind) reflect.Value {

	switch sliceType {
	case reflect.String:
		var valArray []string
		for i := 0; i < value.Len(); i++ {
			valArray = append(valArray, value.Index(i).Interface().(string))
		}
		return reflect.ValueOf(valArray)
	case reflect.Int:
		var valArray []int64
		for i := 0; i < value.Len(); i++ {
			valArray = append(valArray, value.Index(i).Interface().(int64))
		}
		return reflect.ValueOf(valArray)
	case reflect.Uint:
		var valArray []uint64
		for i := 0; i < value.Len(); i++ {
			valArray = append(valArray, value.Index(i).Interface().(uint64))
		}
		return reflect.ValueOf(valArray)
	case reflect.Float64:
		var valArray []float64
		for i := 0; i < value.Len(); i++ {
			valArray = append(valArray, value.Index(i).Interface().(float64))
		}
		return reflect.ValueOf(valArray)
	case reflect.Bool:
		var valArray []bool
		for i := 0; i < value.Len(); i++ {
			valArray = append(valArray, value.Index(i).Interface().(bool))
		}
		return reflect.ValueOf(valArray)
	case reflect.Interface:
		return value
	}
	return value
}

/**
  根据mapType填充map类型
*/
func FillMap(value reflect.Value, mapType reflect.Kind) reflect.Value {
	keys := value.MapKeys()
	fieldValue := reflect.Value{}
	for _, key := range keys {
		switch mapType {
		case reflect.String:
			fieldValue.SetMapIndex(key, reflect.ValueOf(value.MapIndex(key).Interface().(string)))
		case reflect.Int:
			fieldValue.SetMapIndex(key, reflect.ValueOf(value.MapIndex(key).Interface().(int)))
		case reflect.Uint:
			fieldValue.SetMapIndex(key, reflect.ValueOf(value.MapIndex(key).Interface().(uint)))
		case reflect.Float64:
			fieldValue.SetMapIndex(key, reflect.ValueOf(value.MapIndex(key).Interface().(float64)))
		case reflect.Bool:
			fieldValue.SetMapIndex(key, reflect.ValueOf(value.MapIndex(key).Interface().(bool)))
		case reflect.Interface:
			fieldValue.SetMapIndex(key, reflect.ValueOf(value.MapIndex(key).Interface()))
		case reflect.Ptr:
			instance := reflect.New(fieldValue.Type().Elem().Elem())
			Map2Struct(instance.Interface(), value.MapIndex(key).Interface().(map[string]interface{}))
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
			}
			fieldValue.SetMapIndex(key, instance)
		case reflect.Struct:
			instance := reflect.New(fieldValue.Type().Elem())
			Map2Struct(instance.Interface(), value.MapIndex(key).Interface().(map[string]interface{}))
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
			}
			fieldValue.SetMapIndex(key, instance.Elem())
		}
	}
	return fieldValue
}

func IsNil(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	case reflect.Slice, reflect.Map:
		return value.Len() == 0
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

/**
该方法为兼容json默认转换类型导致与结构体类型不匹配问题
类型兼容 int,int8,int32,int64 -> float64
        int,int8,int32,int64 -> int,int8,int32,int64
        int,int8,int32,int64 -> string
        string -> int,int8,int32,int64
        float32,float64 -> string
		string -> float64,float32
*/
func TypeCompatibility(fieldValue reflect.Value, fillValue reflect.Value) reflect.Value {

	fieldValueKind := fieldValue.Kind()
	fillValueKind := fillValue.Kind()
	if (fieldValue.Kind() == reflect.Float32 || fieldValue.Kind() == reflect.Float64) && (fillValueKind == reflect.Int || fillValueKind == reflect.Int8 || fillValueKind == reflect.Int16 || fillValueKind == reflect.Int32 || fillValueKind == reflect.Int64) {
		//int transfer to float
		intStr := strconv.FormatInt(int64(fillValueKind), 10)
		floatVal, err := strconv.ParseFloat(intStr, 32)
		if err != nil {
			panic(err)
		}
		fieldValue.SetFloat(floatVal)
	} else if (fieldValueKind == reflect.Int || fieldValueKind == reflect.Int8 || fieldValueKind == reflect.Int16 || fieldValueKind == reflect.Int32 || fieldValueKind == reflect.Int64) && (fillValueKind == reflect.Int || fillValueKind == reflect.Int8 || fillValueKind == reflect.Int16 || fillValueKind == reflect.Int32 || fillValueKind == reflect.Int64) {
		//int transfer to int
		fieldValue.SetInt(fillValue.Int())
	} else if (fieldValueKind == reflect.Int || fieldValueKind == reflect.Int8 || fieldValueKind == reflect.Int16 || fieldValueKind == reflect.Int32 || fieldValueKind == reflect.Int64) && fillValueKind == reflect.String {
		//string transfer to int
		intVal, err := strconv.Atoi(fillValue.String())
		if err != nil {
			panic(err)
		}
		fieldValue.SetInt(int64(intVal))
	} else if fieldValueKind == reflect.String && (fillValueKind == reflect.Int || fillValueKind == reflect.Int8 || fillValueKind == reflect.Int16 || fillValueKind == reflect.Int32 || fillValueKind == reflect.Int64) {
		//int transfer to string
		fieldValue.SetString(strconv.FormatInt(fillValue.Int(), 10))
	} else if fieldValueKind == reflect.String && (fillValueKind == reflect.Float32 || fillValueKind == reflect.Float64) {
		//float32/64 transfer to string
		fieldValue.SetString(strconv.FormatFloat(fillValue.Float(), 'E', -1, 64))
	} else if (fieldValueKind == reflect.Float32 || fieldValueKind == reflect.Float64) && fillValueKind == reflect.String {
		//string transfer to float32/64
		floatVar, err := strconv.ParseFloat(fillValue.String(), 64)
		if err != nil {
			panic(err)
		}
		fieldValue.SetFloat(floatVar)
	}
	return fieldValue

}
