package lcs

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"sort"
	"strconv"
)

// Marshaler is the interface implemented by types that
// can marshal themselves into valid LCS.
type Marshaler interface {
	MarshalLCS(e *Encoder) error
}

type Encoder struct {
	w     *bufio.Writer
	enums map[reflect.Type]map[string]map[reflect.Type]EnumKeyType
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:     bufio.NewWriter(w),
		enums: make(map[reflect.Type]map[string]map[reflect.Type]EnumKeyType),
	}
}

func (e *Encoder) Encode(v interface{}) error {
	if err := e.encode(reflect.Indirect(reflect.ValueOf(v)), 0); err != nil {
		return err
	}
	e.w.Flush()
	return nil
}

func (e *Encoder) EncodeBytes(b []byte) error {
	return e.encodeSlice(reflect.Indirect(reflect.ValueOf(b)), 0)
}

// @params b must be a byte array of limited length. [N]byte
func (e *Encoder) EncodeFixedBytes(b []byte) (err error) {
	rv := reflect.Indirect(reflect.ValueOf(b))
	l := rv.Len()
	print(l)
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		if err = e.encode(item, 0); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) EncodeUleb128(u uint64) error {
	_, err := writeVarUint(e.w, u)
	return err
}

func (e *Encoder) encode(rv reflect.Value, fixedLen int) (err error) {

	if m, ok := rv.Interface().(Marshaler); ok {
		return m.MarshalLCS(e)
	}

	// rv = indirect(rv)
	switch rv.Kind() {
	case reflect.Bool,
		/*reflect.Int,*/ reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		/*reflect.Uint,*/ reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = binary.Write(e.w, binary.LittleEndian, rv.Interface())
	case reflect.Slice, reflect.Array, reflect.String:
		err = e.encodeSlice(rv, fixedLen)
	case reflect.Struct:
		err = e.encodeStruct(rv)
	case reflect.Map:
		err = e.encodeMap(rv)
	case reflect.Ptr:
		err = e.encode(rv.Elem(), 0)
	case reflect.Interface:
		//err = e.encodeInterface(rv)
		fallthrough
	default:
		err = errors.New("not supported kind: " + rv.Kind().String())
	}
	if err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encodeSlice(rv reflect.Value, fixedLen int) (err error) {
	if rv.Kind() == reflect.Array {
		// ignore fixedLen
	} else if fixedLen == 0 {
		if _, err := writeVarUint(e.w, uint64(rv.Len())); err != nil {
			return err
		}
	} else if fixedLen != rv.Len() {
		return errors.New("actual len not equal to fixed len")
	}
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		if err = e.encode(item, 0); err != nil {
			return err
		}
	}
	return nil
}

//func (e *Encoder) encodeInterface(rv reflect.Value) (err error) {
//	if rv.IsNil() {
//		return errors.New("non-optional enum value is nil")
//	}
//
//	enum, ok := rv.Interface().(Enum)
//	if !ok {
//		return errors.New("enum " + rv.Type().String() + " does not have variant of type ")
//	}
//	ev := uint64(enum.GetIdx())
//	rvReal, ok := enumGetTypeByIdx(rv.Addr(), ev)
//	if !ok {
//		return errors.New("enum " + rv.Type().String() + " does not have variant of type ")
//	}
//
//	if _, err = writeVarUint(e.w, ev); err != nil {
//		return
//	}
//	if err = e.encode(rvReal, 0); err != nil {
//		return err
//	}
//	return nil
//}

func (e *Encoder) encodeStruct(rv reflect.Value) (err error) {
	rt := rv.Type()

	if rt.Implements(reflect.TypeOf((*Enum)(nil)).Elem()) {
		rvReal, ok := enumGetType(rv.Addr())
		if !ok {
			return errors.New("enum " + rv.Type().String() + " does not have variant of type ")
		}
		if _, err = writeVarUint(e.w, uint64(rv.Interface().(Enum).GetIdx())); err != nil {
			return
		}

		if rvReal.Kind() == reflect.Pointer {
			rvReal = rvReal.Elem()
		}
		rv = rvReal
		rt = rv.Type()
	}

	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i)
		if !fv.CanInterface() {
			continue
		}
		if rt.Field(i).Tag.Get(lcsTagName) == "-" {
			continue
		}
		tag := parseTag(rt.Field(i).Tag.Get(lcsTagName))

		if _, ok := tag["optional"]; ok &&
			(fv.Kind() == reflect.Ptr ||
				fv.Kind() == reflect.Slice ||
				fv.Kind() == reflect.Map ||
				fv.Kind() == reflect.Interface) {
			if err = e.encode(reflect.ValueOf(!fv.IsNil()), 0); err != nil {
				return err
			}
			if fv.IsNil() {
				continue
			}
		}
		fixedLen := 0
		if fixedLenStr, ok := tag["len"]; ok && (fv.Kind() == reflect.Slice || fv.Kind() == reflect.String) {
			fixedLen, err = strconv.Atoi(fixedLenStr)
			if err != nil {
				return errors.New("tag len parse error: " + err.Error())
			}
		}
		if err = e.encode(fv, fixedLen); err != nil {
			return
		}
	}
	return nil
}

func (e *Encoder) encodeMap(rv reflect.Value) (err error) {
	_, err = writeVarUint(e.w, uint64(rv.Len()))
	if err != nil {
		return err
	}

	keys := make([]string, 0, rv.Len())
	marshaledMap := make(map[string][]byte)
	for iter := rv.MapRange(); iter.Next(); {
		k := iter.Key()
		v := iter.Value()
		kb, err := Marshal(k.Interface())
		if err != nil {
			return err
		}
		vb, err := Marshal(v.Interface())
		if err != nil {
			return err
		}
		keys = append(keys, string(kb))
		marshaledMap[string(kb)] = vb
	}

	sort.Strings(keys)
	for _, k := range keys {
		e.w.Write([]byte(k))
		e.w.Write(marshaledMap[k])
	}

	return nil
}

func (e *Encoder) getEnumVariants(rv reflect.Value) map[string]map[reflect.Type]EnumKeyType {
	vv, ok := rv.Interface().(EnumTypeUser)
	if !ok {
		vv, ok = reflect.New(reflect.PtrTo(rv.Type())).Elem().Interface().(EnumTypeUser)
		if !ok {
			return nil
		}
	}
	r := make(map[string]map[reflect.Type]EnumKeyType)
	evs := vv.EnumTypes()
	for _, ev := range evs {
		evt := reflect.TypeOf(ev.Template)
		if r[ev.Name] == nil {
			r[ev.Name] = make(map[reflect.Type]EnumKeyType)
		}
		r[ev.Name][evt] = ev.Value
	}
	return r
}

func Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	e := NewEncoder(&b)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
