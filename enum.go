package lcs

import (
	"reflect"
	"unsafe"
)

type EnumT uint64

type Enum interface {
	GetIdx() EnumT
	GetType(t EnumT, pointer unsafe.Pointer) (reflect.Value, error)
}

func enumGetTypeByIdx(enumType reflect.Value, idx uint64) (reflect.Value, bool) {
	reflectType, err := enumType.Interface().(Enum).GetType(EnumT(idx), enumType.UnsafePointer())
	if err != nil {
		return reflect.Value{}, false
	}
	return reflectType, true
}

func enumGetType(enumType reflect.Value) (reflect.Value, bool) {
	return enumGetTypeByIdx(enumType, uint64(enumType.Interface().(Enum).GetIdx()))
}
