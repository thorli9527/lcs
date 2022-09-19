package lcs_test

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/coming-chat/lcs"
)

func ExampleMarshal_struct() {
	type MyStruct struct {
		Boolean    bool
		Bytes      []byte
		Label      string
		unexported uint32
	}
	type Wrapper struct {
		Inner *MyStruct `lcs:"optional"`
		Name  string
	}

	bytes, err := lcs.Marshal(&Wrapper{
		Name: "test",
		Inner: &MyStruct{
			Bytes: []byte{0x01, 0x02, 0x03, 0x04},
			Label: "hello",
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x", bytes)
	// Output: 010004010203040568656c6c6f0474657374
}

func ExampleUnmarshal_struct() {
	type MyStruct struct {
		Boolean    bool
		Bytes      []byte
		Label      string
		unexported uint32
	}
	type Wrapper struct {
		Inner *MyStruct `lcs:"optional"`
		Name  string
	}

	bytes, _ := hex.DecodeString("010004010203040568656c6c6f0474657374")
	out := &Wrapper{}
	err := lcs.Unmarshal(bytes, out)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Name: %s, Label: %s\n", out.Name, out.Inner.Label)
	// Output: Name: test, Label: hello
}

const (
	Uint64Kind lcs.EnumT = iota
	BytesKind
	StringKind
)

type TransactionArgument struct {
	kind lcs.EnumT
}

func (t TransactionArgument) GetIdx() lcs.EnumT {
	return t.kind
}

func (t TransactionArgument) GetType(enumT lcs.EnumT, pointer unsafe.Pointer) (reflect.Value, error) {
	switch enumT {
	case Uint64Kind:
		return reflect.ValueOf((*Uint64)(pointer)), nil
	case BytesKind:
		return reflect.ValueOf((*Bytes)(pointer)), nil
	case StringKind:
		return reflect.ValueOf((*String)(pointer)), nil
	default:
		return reflect.Value{}, errors.New("unknown enum kind")
	}
}

type Uint64 struct {
	TransactionArgument `lcs:"-"`
	Value               uint64 `lcs:"value"`
}

type Bytes struct {
	TransactionArgument `lcs:"-"`
	Value               []byte `lcs:"value"`
}

type String struct {
	TransactionArgument `lcs:"-"`
	Value               string `lcs:"value"`
}

type Program struct {
	Code    []byte
	Args    []*TransactionArgument
	Modules [][]byte
}

func ExampleMarshal_libra_program() {
	d1 := String{TransactionArgument{StringKind}, "CAFE D00D"}
	d2 := Bytes{TransactionArgument{BytesKind}, []byte{0xaa, 0x3d, 0x22}}
	d3 := Uint64{TransactionArgument{Uint64Kind}, 12}
	prog := &Program{
		Code: []byte("move"),
		Args: []*TransactionArgument{
			&d1.TransactionArgument,
			&d2.TransactionArgument,
			&d3.TransactionArgument,
		},
		Modules: [][]byte{{0xca}, {0xfe, 0xd0}, {0x0d}},
	}
	bytes, err := lcs.Marshal(prog)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%X\n", bytes)
	// Output:
	// 046D6F76650302094341464520443030440103AA3D22000C000000000000000301CA02FED0010D
}

func ExampleUnmarshal_libra_program() {
	bytes, _ := hex.DecodeString("046D6F76650302094341464520443030440103AA3D22000C000000000000000301CA02FED0010D")
	out := &Program{}
	err := lcs.Unmarshal(bytes, out)
	if err != nil {
		panic(err)
	}
	d1 := (*String)(unsafe.Pointer(out.Args[0]))
	d2 := (*Bytes)(unsafe.Pointer(out.Args[1]))
	d3 := (*Uint64)(unsafe.Pointer(out.Args[2]))
	fmt.Printf("%+v\n", d1)
	fmt.Printf("%+v\n", d2)
	fmt.Printf("%+v\n", d3)
	fmt.Printf("%+v\n", out.Code)
	fmt.Printf("%+v\n", out.Modules)
	// Output:
	// &{TransactionArgument:{kind:2} Value:CAFE D00D}
	// &{TransactionArgument:{kind:1} Value:[170 61 34]}
	// &{TransactionArgument:{kind:0} Value:12}
	// [109 111 118 101]
	// [[202] [254 208] [13]]
}
