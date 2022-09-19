package lcs_test

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"testing"
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

//type TransactionArgument interface{}

// Register TransactionArgument with LCS. Will be available globaly.
//var _ = lcs.RegisterEnum(
//	// pointer to enum interface:
//	(*TransactionArgument)(nil),
//	// zero-value of variants:
//	uint64(0), [32]byte{}, "",
//)

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
	Value               [32]byte `lcs:"value"`
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
	a1 := String{TransactionArgument{StringKind}, "CAFE D00D"}
	a2 := String{TransactionArgument{StringKind}, "cafe d00d"}
	prog := &Program{
		Code: []byte("move"),
		Args: []*TransactionArgument{
			&a1.TransactionArgument,
			&a2.TransactionArgument,
		},
		Modules: [][]byte{{0xca}, {0xfe, 0xd0}, {0x0d}},
	}
	bytes, err := lcs.Marshal(prog)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%X\n", bytes)
	// Output:
	// 046D6F766502020943414645204430304402096361666520643030640301CA02FED0010D
}

func ExampleUnmarshal_libra_program() {
	bytes, _ := hex.DecodeString("046D6F766502020943414645204430304402096361666520643030640301CA02FED0010D")
	out := &Program{}
	err := lcs.Unmarshal(bytes, out)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", out)
	// Output:
	// &{Code:[109 111 118 101] Args:[CAFE D00D cafe d00d] Modules:[[202] [254 208] [13]]}
}

func TestEEE(t *testing.T) {
	a1 := String{TransactionArgument{StringKind}, "CAFE D00D"}
	a2 := Bytes{TransactionArgument{StringKind}, [32]byte{}}
	prog := &Program{
		Code: []byte("move"),
		Args: []*TransactionArgument{
			&a1.TransactionArgument,
			&a2.TransactionArgument,
		},
		Modules: [][]byte{{0xca}, {0xfe, 0xd0}, {0x0d}},
	}
	ty := reflect.TypeOf((*String)(nil))
	newT := reflect.NewAt(ty, unsafe.Pointer(prog.Args[0]))
	fmt.Println(newT)
}
