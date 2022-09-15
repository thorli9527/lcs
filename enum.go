package lcs

import (
	"reflect"
)

type EnumT uint64

type Enum interface {
	GetIdx() EnumT
	GetType() (reflect.Type, error)
}

func enumGetTypeByIdx(enumType reflect.Value) (reflect.Type, bool) {
	enum, ok := enumType.Interface().(Enum)
	if !ok {
		return nil, false
	}
	reflectType, err := enum.GetType()
	if err != nil {
		return nil, false
	}
	return reflectType, true
}

func enumGetIdxByType(enumType reflect.Value) (uint64, bool) {
	enum, ok := enumType.Interface().(Enum)
	if !ok {
		return 0, false
	}
	return uint64(enum.GetIdx()), true
}
