package db

import (
	"fmt"
	"reflect"
)

type rule struct {
	check  func(reflect.Type) bool
	decode func(Converter, reflect.Value) (interface{}, error)
	encode func(Converter, interface{}, reflect.Value) error
}

func (r *rule) Check(tpe reflect.Type) bool {
	return r.check(tpe)
}

func (r *rule) Decode(conv Converter, value reflect.Value) (interface{}, error) {
	return r.decode(conv, value)
}

func (r *rule) Encode(conv Converter, i interface{}, value reflect.Value) error {
	return r.encode(conv, i, value)
}

type ptrRule struct{}

func PtrRule() ConvertRule {
	return &ptrRule{}
}

func (*ptrRule) Check(tpe reflect.Type) bool {
	return tpe.Kind() == reflect.Ptr
}

func (*ptrRule) Decode(conv Converter, value reflect.Value) (interface{}, error) {
	if value.IsNil() {
		return nil, nil
	}
	return conv.Decode(value.Elem())
}

func (*ptrRule) Encode(conv Converter, i interface{}, value reflect.Value) error {
	if i == nil {
		value.Set(reflect.Zero(value.Type()))
		return nil
	}
	evalue := value.Elem()
	etype := value.Type().Elem()
	for _, rule := range conv.Rules() {
		if rule.Check(etype) {
			return rule.Encode(conv, i, evalue)
		}
	}
	return nil
}

type structRule struct{}

func StructRule() ConvertRule {
	return &structRule{}
}

func (*structRule) Check(tpe reflect.Type) bool {
	return tpe.Kind() == reflect.Struct
}

func (*structRule) Decode(conv Converter, value reflect.Value) (interface{}, error) {
	result := make(map[string]interface{})

	itype := value.Type()

	fnames := conv.FieldNames(itype)
	for mname := range fnames {
		sname := fnames[mname]
		if _, ok := itype.FieldByName(sname); ok {
			fvalue := value.FieldByName(sname)
			i, err := conv.Decode(fvalue)
			if err != nil {
				return result, err
			}
			result[mname] = i
		}
	}

	return result, nil
}

func (*structRule) Encode(conv Converter, i interface{}, value reflect.Value) error {
	itype := reflect.TypeOf(i)
	if itype.Kind() == reflect.Map && itype.Key().Kind() == reflect.String {
		ivalue := reflect.ValueOf(i)
		vtype := value.Type()

		fnames := conv.FieldNames(vtype)
		for mname := range fnames {
			if mapvalue := ivalue.MapIndex(reflect.ValueOf(mname)); mapvalue.IsValid() {
				sname := fnames[mname]
				if _, ok := vtype.FieldByName(sname); ok {
					fvalue := value.FieldByName(sname)
					if err := conv.Encode(mapvalue.Interface(), fvalue); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}
	return fmt.Errorf("Only maps of string keys and interface{} values can encoded to struct, but got %T", i)
}

func SliceRule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Slice
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			result := make([]interface{}, value.Len())
			for i := 0; i < value.Len(); i++ {
				vitem := value.Index(i)
				res, err := conv.Decode(vitem)
				if err != nil {
					return nil, err
				}
				result[i] = res
			}
			return result, nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ivalue := reflect.ValueOf(i)
			vtype := value.Type()
			etype := vtype.Elem()
			ekind := etype.Kind()
			if ekind == reflect.Ptr {
				etype = etype.Elem()
			}
			result := reflect.MakeSlice(vtype, 0, 0)
			for i := 0; i < ivalue.Len(); i++ {
				itemvalue := ivalue.Index(i)
				ptrevalue := reflect.New(etype)
				evalue := ptrevalue.Elem()
				if err := conv.Encode(itemvalue.Interface(), evalue); err != nil {
					return err
				}
				if ekind == reflect.Ptr {
					result = reflect.Append(result, ptrevalue)
				} else {
					result = reflect.Append(result, evalue)
				}
			}
			value.Set(result)
			return nil
		},
	}
}

/*func BinaryRule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return (tpe.Kind() == reflect.Array || tpe.Kind() == reflect.Slice) && tpe.Elem().Kind() == reflect.Uint8
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {

		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {

		},
	}
}*/

func StringKeyMapRule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Map && tpe.Key().Kind() == reflect.String
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			vtype := value.Type()
			keytype := vtype.Key()
			if keytype.Kind() != reflect.String {
				return nil, fmt.Errorf("Unsupported key-type: %v", keytype)
			}
			result := make(map[string]interface{})
			for _, key := range value.MapKeys() {
				mkey := key.String()
				mvalue, err := conv.Decode(value.MapIndex(key))
				if err != nil {
					return nil, err
				}
				result[mkey] = mvalue
			}
			return result, nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			itype := reflect.TypeOf(i)
			if itype.Kind() == reflect.Map && itype.Key().Kind() == reflect.String {
				ivalue := reflect.ValueOf(i)
				vtype := value.Type()
				valuetype := vtype.Elem()
				valuekind := valuetype.Kind()
				result := reflect.MakeMap(vtype)
				if valuekind == reflect.Ptr {
					valuetype = valuetype.Elem()
				}
				for _, key := range ivalue.MapKeys() {
					if mapvalue := ivalue.MapIndex(key); mapvalue.IsValid() {
						ptritemvalue := reflect.New(valuetype)
						itemvalue := ptritemvalue.Elem()
						if err := conv.Encode(mapvalue.Interface(), itemvalue); err != nil {
							return err
						}
						if valuekind == reflect.Ptr {
							result.SetMapIndex(key, ptritemvalue)
						} else {
							result.SetMapIndex(key, itemvalue)
						}
					}
				}
				value.Set(result)
				return nil
			}
			return fmt.Errorf("Could not cast %T to map[string]interface{}", i)
		},
	}
}

func MapRule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Map
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			vtype := value.Type()
			keytype := vtype.Key()
			if keytype.Kind() == reflect.Ptr || keytype.Kind() == reflect.Struct || keytype.Kind() == reflect.Slice || keytype.Kind() == reflect.Chan {
				return nil, fmt.Errorf("Unsupported key-type: %v", keytype)
			}
			result := make(map[interface{}]interface{})
			for _, key := range value.MapKeys() {
				mkey, err := conv.Decode(key)
				if err != nil {
					return nil, err
				}
				mvalue, err := conv.Decode(value.MapIndex(key))
				if err != nil {
					return nil, err
				}
				result[mkey] = mvalue
			}
			return result, nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			switch i.(type) {
			case map[interface{}]interface{}:
				smap := i.(map[interface{}]interface{})
				vtype := value.Type()
				keytype := vtype.Key()
				valuetype := vtype.Elem()
				valuekind := valuetype.Kind()
				result := reflect.MakeMap(vtype)
				if valuekind == reflect.Ptr {
					valuetype = valuetype.Elem()
				}
				for key := range smap {
					item := smap[key]
					ptrkeyvalue := reflect.New(keytype)
					keyvalue := ptrkeyvalue.Elem()
					if err := conv.Encode(key, keyvalue); err != nil {
						return err
					}
					ptritemvalue := reflect.New(valuetype)
					itemvalue := ptritemvalue.Elem()
					if err := conv.Encode(item, itemvalue); err != nil {
						return err
					}
					if valuekind == reflect.Ptr {
						result.SetMapIndex(keyvalue, ptritemvalue)
					} else {
						result.SetMapIndex(keyvalue, itemvalue)
					}
				}
				value.Set(result)
				return nil
			default:
				return fmt.Errorf("Could not cast %T to map[interface{}]interface{}", i)
			}
		},
	}
}

func StringRule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.String
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return value.String(), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			switch i.(type) {
			case string:
				value.Set(reflect.ValueOf(i))
				return nil
			case fmt.Stringer:
				str := i.(fmt.Stringer)
				value.Set(reflect.ValueOf(str.String()))
				return nil
			default:
				return fmt.Errorf("Could not encode %T to string", i)
			}
		},
	}
}

func BoolRule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Bool
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return value.Bool(), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			switch i.(type) {
			case bool:
				value.Set(reflect.ValueOf(i))
				return nil
			default:
				return fmt.Errorf("Could not encode %T to bool", i)
			}
		},
	}
}

func IntRule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Int
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return value.Int(), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func Int8Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Int8
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return int8(value.Int()), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func Int16Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Int16
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return int16(value.Int()), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func Int32Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Int32
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return int32(value.Int()), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func Int64Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Int64
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return int64(value.Int()), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func UintRule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Uint
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return uint(value.Uint()), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func Uint8Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Uint8
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return uint8(value.Uint()), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func Uint16Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Uint16
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return uint16(value.Uint()), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func Uint32Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Uint32
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return uint32(value.Uint()), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func Uint64Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Uint64
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return value.Uint(), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			ival, err := CastI(i, value.Kind())
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(ival))
			return nil
		},
	}
}

func Float32Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Float32
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return float32(value.Float()), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			switch i.(type) {
			case float32:
				value.Set(reflect.ValueOf(i))
				return nil
			default:
				return fmt.Errorf("Could not encode %T to float32", i)
			}
		},
	}
}

func Float64Rule() ConvertRule {
	return &rule{
		check: func(tpe reflect.Type) bool {
			return tpe.Kind() == reflect.Float64
		},
		decode: func(conv Converter, value reflect.Value) (interface{}, error) {
			return value.Float(), nil
		},
		encode: func(conv Converter, i interface{}, value reflect.Value) error {
			switch i.(type) {
			case float64:
				value.Set(reflect.ValueOf(i))
				return nil
			default:
				return fmt.Errorf("Could not encode %T to float64", i)
			}
		},
	}
}
