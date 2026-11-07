package cast

import (
	"encoding/binary"
	"errors"
)

// CastMessage protobuf encoding. Only the fields goplay uses:
// 1 protocol_version (varint), 2 source_id, 3 destination_id,
// 4 namespace, 5 payload_type (varint, 0 = string), 6 payload_utf8.

type castMessage struct {
	source, dest, namespace, payload string
}

func encodeCastMessage(m castMessage) []byte {
	var b []byte
	b = appendVarintField(b, 1, 0)
	b = appendStringField(b, 2, m.source)
	b = appendStringField(b, 3, m.dest)
	b = appendStringField(b, 4, m.namespace)
	b = appendVarintField(b, 5, 0)
	b = appendStringField(b, 6, m.payload)
	return b
}

func appendVarintField(b []byte, field int, v uint64) []byte {
	b = binary.AppendUvarint(b, uint64(field<<3|0))
	return binary.AppendUvarint(b, v)
}

func appendStringField(b []byte, field int, s string) []byte {
	b = binary.AppendUvarint(b, uint64(field<<3|2))
	b = binary.AppendUvarint(b, uint64(len(s)))
	return append(b, s...)
}

func decodeCastMessage(b []byte) (castMessage, error) {
	var m castMessage
	for len(b) > 0 {
		key, n := binary.Uvarint(b)
		if n <= 0 {
			return m, errors.New("cast: bad message")
		}
		b = b[n:]
		field, wire := int(key>>3), key&7
		switch wire {
		case 0:
			_, n := binary.Uvarint(b)
			if n <= 0 {
				return m, errors.New("cast: bad varint")
			}
			b = b[n:]
		case 2:
			l, n := binary.Uvarint(b)
			if n <= 0 || uint64(len(b)-n) < l {
				return m, errors.New("cast: bad length")
			}
			v := string(b[n : n+int(l)])
			b = b[n+int(l):]
			switch field {
			case 2:
				m.source = v
			case 3:
				m.dest = v
			case 4:
				m.namespace = v
			case 6:
				m.payload = v
			}
		default:
			return m, errors.New("cast: unexpected wire type")
		}
	}
	return m, nil
}
