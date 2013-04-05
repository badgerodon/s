package s

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

type (
	Encoder interface {
		EncodeS() (Expression, error)
	}
)

func encodeValue(src reflect.Value) (exp Expression, err error) {
	var ok bool

	if src.CanInterface() {
		err, ok = src.Interface().(error)
		if ok {
			exp = NewList(Identifier("error"), String(err.Error()))
			err = nil
			return
		}

		encoder, ok := src.Interface().(Encoder)
		if !ok {
			if src.Kind() != reflect.Ptr && src.CanAddr() {
				encoder, ok = src.Addr().Interface().(Encoder)
				if ok {
					src = src.Addr()
				}
			}
		}
		if ok && (src.Kind() != reflect.Ptr || !src.IsNil()) {
			exp, err = encoder.EncodeS()
			return
		}
	}

	switch src.Kind() {
	case reflect.Bool:
		if src.Bool() {
			exp = True{}
		} else {
			exp = False{}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		exp = Number(strconv.FormatInt(src.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		exp = Number(strconv.FormatUint(src.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		f := src.Float()
		if math.IsInf(f, 0) {
			err = fmt.Errorf("Infinity is not supported")
		} else if math.IsNaN(f) {
			err = fmt.Errorf("NaN is not supported")
		} else {
			exp = Number(strconv.FormatFloat(f, 'f', -1, 64))
		}
	case reflect.String:
		exp = String(src.String())
	case reflect.Struct:
		exps := make([]Expression, src.NumField())
		for i := 0; i < src.NumField(); i++ {
			var e Expression
			e, err = encodeValue(src.Field(i))
			if err != nil {
				return
			}
			exps[i] = e
		}
		exp = List(exps)
	case reflect.Map:
		if src.IsNil() {
			exp = NewList()
			break
		}
		exps := make([]Expression, src.Len())
		for i, k := range src.MapKeys() {
			var ke, ve Expression
			ke, err = encodeValue(k)
			if err != nil {
				return
			}
			ve, err = encodeValue(src.MapIndex(k))
			if err != nil {
				return
			}
			exps[i] = NewList(ke, ve)
		}
		exp = List(exps)
	case reflect.Slice:
		if src.IsNil() {
			exp = NewList()
			break
		}
		fallthrough
	case reflect.Array:
		exps := make([]Expression, src.Len())
		for i := 0; i < src.Len(); i++ {
			var ve Expression
			ve, err = encodeValue(src.Index(i))
			if err != nil {
				return
			}
			exps[i] = ve
		}
		exp = List(exps)
	case reflect.Interface, reflect.Ptr:
		if src.IsNil() {
			exp = NewList()
			break
		}
		exp, err = encodeValue(src.Elem())
	default:
		err = fmt.Errorf("Unable to convert `%v` of type `%v` into s expression", src, src.Kind())
	}
	return
}

func Encode(src interface{}) (Expression, error) {
	return encodeValue(reflect.ValueOf(src))
}
