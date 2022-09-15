package lcs

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"
)

const (
	Enum10pt0Kind EnumT = iota
	Enum10pt1Kind
	Enum10pt2Kind
	Enum10pt3Kind
)

type Enum1 struct {
	kind EnumT
}

func (e Enum1) GetIdx() EnumT {
	return e.kind
}

func (e Enum1) GetType() (reflect.Type, error) {
	switch e.kind {
	case Enum10pt0Kind:
		return reflect.TypeOf(Enum1Opt0{}), nil
	case Enum10pt1Kind:
		return reflect.TypeOf(Enum1Opt1{}), nil
	case Enum10pt2Kind:
		return reflect.TypeOf(Enum1Opt2{}), nil
	case Enum10pt3Kind:
		return reflect.TypeOf(Enum1Opt3{}), nil
	default:
		return nil, errors.New("unknown enum kind")
	}
}

type Enum1Opt0 struct {
	Enum1
	Data uint32
}
type Enum1Opt1 struct {
	Enum1
	Data bool
}

type Enum1Opt2 struct {
	Enum1
	Data []byte
}
type Enum1Opt3 struct {
	Enum1
	Data []Enum1
}

func TestInterfaceAsEnum(t *testing.T) {
	enum10pt0 := Enum1Opt0{Enum1{Enum10pt0Kind}, 3}
	enum10pt1 := Enum1Opt1{Enum1{Enum10pt1Kind}, true}
	enum10pt2 := Enum1Opt2{Enum1{Enum10pt2Kind}, []byte{1, 2, 3}}
	enum10pt3 := Enum1Opt3{Enum1{Enum10pt3Kind}, []Enum1{enum10pt0.Enum1, enum10pt1.Enum1, enum10pt2.Enum1}}
	//	e0 := &enum10pt0.Enum1
	//	e1 := &enum10pt1.Enum1
	//	e2 := &enum10pt2.Enum1
	e3 := &enum10pt3.Enum1
	data := (*Enum1Opt3)(unsafe.Pointer(e3))
	t.Log(data)
	t.Log(enum10pt3)

	//runTest(t, []*testCase{
	//	{
	//		v:    &e0,
	//		b:    hexMustDecode("00 03000000"),
	//		name: "struct pointer",
	//	},
	//	{
	//		v:    &e1,
	//		b:    hexMustDecode("01 01"),
	//		name: "bool",
	//	},
	//	{
	//		v:    &e2,
	//		b:    hexMustDecode("02 02 1122"),
	//		name: "[]byte",
	//	},
	//	{
	//		v:    &e3,
	//		b:    hexMustDecode("03 02 00 03000000 01 01"),
	//		name: "enum slice of self",
	//	},
	//})
}
