package unserializer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Unserializer struct {
}

func NewUnserializer() *Unserializer {
	return &Unserializer{}
}

func readVarint(buffer *bytes.Buffer) (uint64, error) {
	value, err := binary.ReadUvarint(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		return 0, err
	}

	n := 0
	temp := value
	for temp > 0 {
		temp >>= 7
		n++
	}

	buffer.Next(n)
	return value, nil
}

func (u *Unserializer) parseProtobufUnknown(data []byte) (map[uint64]interface{}, error) {
	buffer := bytes.NewBuffer(data)
	result := make(map[uint64]interface{})

	for {
		key, err := readVarint(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		fieldNumber := key >> 3
		wireType := key & 0x7

		var value interface{}
		switch wireType {
		case 0: // Varint
			value, err = readVarint(buffer)
		case 1: // Fixed64
			err = binary.Read(buffer, binary.LittleEndian, &value)
		case 2: // Length-delimited
			length, _ := readVarint(buffer)
			rawData := buffer.Next(int(length))
			// 这里递归下降查找,如果出错不意味着最终错误,说明此处消息并非嵌套消息
			res, err := u.parseProtobufUnknown(rawData)
			if err != nil {
				value = string(rawData)
			} else {
				value = res
			}
		case 5: // Fixed32
			var fixedValue uint32
			err = binary.Read(buffer, binary.LittleEndian, &fixedValue)
			value = fixedValue
		default:
			return nil, fmt.Errorf("unknown wire type: %d", wireType)
		}

		if err != nil {
			return nil, err
		}

		if existing, ok := result[fieldNumber]; ok {
			switch existing := existing.(type) {
			case []interface{}:
				result[fieldNumber] = append(existing, value)
			default:
				result[fieldNumber] = []interface{}{existing, value}
			}
		} else {
			result[fieldNumber] = value
		}
	}

	return result, nil
}

func (u *Unserializer) Unserialize(data []byte) (map[uint64]interface{}, error) {
	return u.parseProtobufUnknown(data)
}
