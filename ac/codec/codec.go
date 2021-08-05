package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

//  todo:
// 		interface 兼容性(地址和对象进行区分)
//		[]interface 处理不了 需解决
// 		slice 的数目进行优化(目前报文中的slice都是在最后，解析时也是一直解析到结束)
//
//  remark:
// 		binary.Write 不可以打包变长结构
//		struct 有对齐/补齐的时候， 计算大小和输入的字节不相同

var (
	ErrDataLen    = errors.New("data len error")
	ErrNotSupport = errors.New("not support type")
)

var endian binary.ByteOrder

func init() {
	endian = binary.LittleEndian // 默认小端
}

// SetEndian 设置编码大小端
func SetEndian(e binary.ByteOrder) {
	endian = e
}
func GetEndian() binary.ByteOrder {
	return endian
}

// Marshal 编码
func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := marshal(buf, reflect.ValueOf(v))
	return buf.Bytes(), err

}

func marshal(buf io.Writer, val reflect.Value) (err error) {
	kind := val.Type().Kind()
	if val.Type().Kind() == reflect.Ptr {
		kind = val.Elem().Kind()
	}
	switch kind {
	case reflect.Struct:
		ref := val.Elem()
		for i := 0; i < ref.NumField(); i++ {
			field := ref.Field(i)
			if err = marshal(buf, field); err != nil {
				return err
			}
		}
	case reflect.Interface:
		// fmt.Println("-----> interface", val.Type(), val.Elem().Type())
		if err = marshal(buf, val.Elem()); err != nil {
			return err
		}
	default:
		return binary.Write(buf, endian, val.Interface())
	}
	return nil
}

// Unmarshal 解码
func Unmarshal(b []byte, v interface{}) error {
	if len(b) <= 0 {
		return errors.New("buf is nil")
	}
	// buf  := b
	// return unmarshal(&buf,  reflect.ValueOf(v))

	buf, offset := b, 0
	return unmarshal(&buf, &offset, reflect.ValueOf(v))
}

func unmarshal(b *[]byte, offset *int, val reflect.Value) (err error) {
	if len((*b)[*offset:]) <= 0 {
		// return errors.New("buf is nil")
		return nil // 当协议在后面添加字段时，需兼容老协议
	}
	kind := val.Type().Kind()
	if val.Type().Kind() == reflect.Ptr {
		kind = val.Elem().Kind()
	}
	// fmt.Printf("----> codec unmarshal begin: [%v][%d][% x]\r\n", kind, len((*b)[*offset:]), (*b)[*offset:])
	switch kind {
	case reflect.Int8:
		if len((*b)[*offset:]) < 1 {
			return ErrDataLen
		}
		val.SetInt(int64(int8((*b)[*offset])))
		(*offset)++
	case reflect.Uint8: // uint8
		if len((*b)[*offset:]) < 1 {
			return ErrDataLen
		}
		val.SetUint(uint64(uint8((*b)[*offset])))
		(*offset)++
	case reflect.Uint16: // uint16
		if len((*b)[*offset:]) < 2 {
			return ErrDataLen
		}
		val.SetUint(uint64(endian.Uint16((*b)[*offset : *offset+2])))
		(*offset) += 2
	case reflect.Uint32: // uint32
		if len((*b)[*offset:]) < 4 {
			return ErrDataLen
		}
		val.SetUint(uint64(endian.Uint32((*b)[*offset : *offset+4])))
		(*offset) += 4
	case reflect.Uint64: // uint64
		if len((*b)[*offset:]) < 8 {
			return ErrDataLen
		}
		val.SetUint(endian.Uint64((*b)[*offset : *offset+8]))
		(*offset) += 8
	case reflect.Array:
		typesize := int(val.Type().Size())
		if len((*b)[*offset:]) < typesize {
			return ErrDataLen
		}

		elemsize := int(val.Type().Elem().Size())
		l := typesize / elemsize
		for cur, j := 0, 0; j < l; j++ {
			tmpOffset := cur + j*elemsize + (*offset)
			switch elemsize {
			case 1:
				val.Index(j).SetUint(uint64((*b)[tmpOffset]))
			case 2:
				val.Index(j).SetUint(uint64(endian.Uint16((*b)[tmpOffset:(tmpOffset + elemsize)])))
			case 4:
				val.Index(j).SetUint(uint64(endian.Uint32((*b)[tmpOffset:(tmpOffset + elemsize)])))
			case 8:
				val.Index(j).SetUint(uint64(endian.Uint64((*b)[tmpOffset:(tmpOffset + elemsize)])))
			default:
				return ErrNotSupport
			}
		}
		// *b = (*b)[typesize:]
		(*offset) += typesize
	case reflect.Interface:
		if err := unmarshal(b, offset, val.Elem()); err != nil {
			return err
		}
	case reflect.Struct:
		ref := val.Elem()
		for i := 0; i < ref.NumField(); i++ {
			field := ref.Field(i)
			if len((*b)) <= 0 {
				break
			}
			if err := unmarshal(b, offset, field); err != nil {
				return err
			}
		}
	case reflect.Slice:
		val.Set(reflect.MakeSlice(val.Type(), 0, 0))
		isPtr := false
		destType := val.Type().Elem()
		if destType.Kind() == reflect.Ptr {
			isPtr = true
			destType = destType.Elem()
		}

		num := 0
		if *offset > 0 {
			num = int((*b)[*offset-1])
		}
		for j := 0; ; j++ { // todo: 数量优化, 目前是一直取到结束
			if l := len((*b)[*offset:]); l <= 0 {
				// fmt.Println("---> exit", l, tsize, j)
				break
			} else if num > 0 && j >= num {
				break
			}

			elem := reflect.New(destType)
			// tsize := int(elem.Elem().Type().Size())
			// fmt.Printf("-------> %v %v %v %v buf[%d][%v]\r\n", elem.Kind(), elem.Elem().Type(), elem.Elem().Kind(), tsize,  len((*b)[*offset:]) , (*b))
			// if l :=  len((*b)[*offset:]) ; l <= 0 || l-tsize < 0 { 结构体 tsize 因为有对齐/不齐影响 所以 tsize不是结构真正大小, 误差值为4

			if destType.Kind() == reflect.Struct {
				err = unmarshal(b, offset, elem)
			} else {
				err = unmarshal(b, offset, elem.Elem())
			}
			if err != nil {
				return err
			}
			if isPtr {
				val.Set(reflect.Append(val, elem.Addr()))
			} else {
				val.Set(reflect.Append(val, elem.Elem()))
			}
		}
	default:
		return fmt.Errorf("unmarshal type [%v] not support", kind)
	}
	// fmt.Printf("----> codec unmarshal end: [%d][% x]\r\n", len(*b), *b)

	return nil
}

// func unmarshal(b *[]byte, val reflect.Value) (err error) {
// 	if len(*b) <= 0 {
// 		// return errors.New("buf is nil")
// 		return nil // 当协议在后面添加字段时，需兼容老协议
// 	}
// 	kind := val.Type().Kind()
// 	if val.Type().Kind() == reflect.Ptr {
// 		kind = val.Elem().Kind()
// 	}
// 	// fmt.Printf("----> codec unmarshal begin: [%v][%d][% x]\r\n", kind, len(*b), *b)
// 	switch kind {
// 	case reflect.Int8:
// 		if len(*b) < 1 {
// 			return ErrDataLen
// 		}
// 		val.SetInt(int64(int8((*b)[0])))
// 		*b = (*b)[1:]
// 		// val.SetInt(int64(int8((*b)[*offset])))
// 		// (*offset)++
// 	case reflect.Uint8: // uint8
// 		if len(*b) < 1 {
// 			return ErrDataLen
// 		}
// 		val.SetUint(uint64(uint8((*b)[0])))
// 		*b = (*b)[1:]
// 		// val.SetUint(uint64(uint8((*b)[*offset])))
// 		// (*offset)++
// 	case reflect.Uint16: // uint16
// 		if len(*b) < 2 {
// 			return ErrDataLen
// 		}
// 		val.SetUint(uint64(endian.Uint16((*b)[0:2])))
// 		*b = (*b)[2:]
// 		// val.SetUint(uint64(endian.Uint16((*b)[*offset : *offset+2])))
// 		// (*offset) += 2
// 	case reflect.Uint32: // uint32
// 		if len(*b) < 4 {
// 			return ErrDataLen
// 		}
// 		// fmt.Printf("-------->uint32:[%x][%+v][%+v]\r\n", (*b)[:4], val.Type().Name(), endian.Uint32((*b)[:4]))
// 		val.SetUint(uint64(endian.Uint32((*b)[:4])))
// 		*b = (*b)[4:]
// 		// val.SetUint(uint64(endian.Uint32((*b)[*offset : *offset+4])))
// 		// (*offset) += 4
// 	case reflect.Uint64: // uint64
// 		if len(*b) < 8 {
// 			return ErrDataLen
// 		}
// 		val.SetUint(endian.Uint64((*b)[:8]))
// 		*b = (*b)[8:]
// 		// val.SetUint(endian.Uint64((*b)[*offset : *offset+8]))
// 		// (*offset) += 8
// 	case reflect.Array:
// 		typesize := int(val.Type().Size())
// 		if len(*b) < typesize {
// 			return ErrDataLen
// 		}

// 		elemsize := int(val.Type().Elem().Size())
// 		l := typesize / elemsize
// 		for cur, j := 0, 0; j < l; j++ {
// 			tmpOffset := cur + j*elemsize
// 			switch elemsize {
// 			case 1:
// 				val.Index(j).SetUint(uint64((*b)[tmpOffset]))
// 			case 2:
// 				val.Index(j).SetUint(uint64(endian.Uint16((*b)[tmpOffset:(tmpOffset + elemsize)])))
// 			case 4:
// 				val.Index(j).SetUint(uint64(endian.Uint32((*b)[tmpOffset:(tmpOffset + elemsize)])))
// 			case 8:
// 				val.Index(j).SetUint(uint64(endian.Uint64((*b)[tmpOffset:(tmpOffset + elemsize)])))
// 			default:
// 				return ErrNotSupport
// 			}
// 		}
// 		*b = (*b)[typesize:]
// 		// (*offset) += uint32(typesize)
// 	case reflect.Interface:
// 		if err := unmarshal(b, val.Elem()); err != nil {
// 			return err
// 		}
// 	case reflect.Struct:
// 		ref := val.Elem()
// 		for i := 0; i < ref.NumField(); i++ {
// 			field := ref.Field(i)
// 			if len((*b)) <= 0 {
// 				break
// 			}
// 			if err := unmarshal(b, field); err != nil {
// 				return err
// 			}
// 		}
// 	case reflect.Slice:
// 		val.Set(reflect.MakeSlice(val.Type(), 0, 0))
// 		isPtr := false
// 		destType := val.Type().Elem()
// 		if destType.Kind() == reflect.Ptr {
// 			isPtr = true
// 			destType = destType.Elem()
// 		}

// 		for j := 0; ; j++ { // todo: 数量优化, 目前是一直取到结束
// 			elem := reflect.New(destType)
// 			// tsize := int(elem.Elem().Type().Size())
// 			// fmt.Printf("-------> %v %v %v %v buf[%d][%v]\r\n", elem.Kind(), elem.Elem().Type(), elem.Elem().Kind(), tsize, len(*b), (*b))
// 			// if l := len(*b); l <= 0 || l-tsize < 0 { 结构体 tsize 因为有对齐/不齐影响 所以 tsize不是结构真正大小, 误差值为4
// 			if l := len(*b); l <= 0 {
// 				// fmt.Println("---> exit", l, tsize, j)
// 				break
// 			}
// 			if destType.Kind() == reflect.Struct {
// 				err = unmarshal(b, elem)
// 			} else {
// 				err = unmarshal(b, elem.Elem())
// 			}
// 			if err != nil {
// 				return err
// 			}
// 			if isPtr {
// 				val.Set(reflect.Append(val, elem.Addr()))
// 			} else {
// 				val.Set(reflect.Append(val, elem.Elem()))
// 			}
// 		}
// 	default:
// 		return fmt.Errorf("unmarshal type [%v] not support", kind)
// 	}
// 	// fmt.Printf("----> codec unmarshal end: [%d][% x]\r\n", len(*b), *b)

// 	return nil
// }
