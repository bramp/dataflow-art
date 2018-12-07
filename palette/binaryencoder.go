package palette

import (
	"bytes"
	"encoding/binary"
)

// BinaryEncoder is a simple helper type for writing Varints, etc to a []byte.

type BinaryEncoder struct {
	b bytes.Buffer
}

// Varint writes a int64 in varint encoding.
func (e *BinaryEncoder) Varint(x int64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	length := binary.PutVarint(buf, x)
	_, err := e.b.Write(buf[:length])
	return err
}

// Uvarint writes a uint64 in varint encoding.
func (e *BinaryEncoder) Uvarint(x uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	length := binary.PutUvarint(buf, x)
	_, err := e.b.Write(buf[:length])
	return err
}

// Float64 writes a float64 in IEEE 754 binary representation.
func (e *BinaryEncoder) Float64(x float64) error {
	return binary.Write(&e.b, binary.LittleEndian, x)
}

// Bytes returns the []byte with the encoding data
func (e *BinaryEncoder) Bytes() []byte {
	return e.b.Bytes()
}
