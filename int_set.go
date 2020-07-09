package rdb

import (
	"fmt"
	"io"

	"github.com/tommy351/rdb-go/internal/reader"
)

type intSetIterator struct {
	DataKey DataKey
	Reader  reader.BytesReader
	Mapper  collectionMapper

	buf      *reader.Buffer
	done     bool
	encoding uint32
	index    int
	length   int
	values   []interface{}
}

func (i *intSetIterator) Next() (interface{}, error) {
	if i.done {
		return nil, io.EOF
	}

	if i.buf == nil {
		buf, err := readStringEncoding(i.Reader)

		if err != nil {
			return nil, fmt.Errorf("failed to read intset buffer: %w", err)
		}

		i.buf = reader.NewBuffer(buf)

		if i.encoding, err = reader.ReadUint32(i.buf); err != nil {
			return nil, fmt.Errorf("failed to read intset encoding: %w", err)
		}

		length, err := reader.ReadUint32(i.buf)

		if err != nil {
			return nil, fmt.Errorf("failed to read intset length: %w", err)
		}

		i.length = int(length)

		head, err := i.Mapper.MapHead(&collectionHead{
			DataKey: i.DataKey,
			Length:  i.length,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to map head in intset: %w", err)
		}

		return head, nil
	}

	if i.index == i.length {
		i.done = true
		i.buf = nil

		slice, err := i.Mapper.MapSlice(&collectionSlice{
			DataKey: i.DataKey,
			Value:   i.values,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to map slice in intset: %w", err)
		}

		return slice, nil
	}

	value, err := i.readValue()

	if err != nil {
		return nil, err
	}

	entry, err := i.Mapper.MapEntry(&collectionEntry{
		DataKey: i.DataKey,
		Index:   i.index,
		Length:  i.length,
		Value:   value,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to map entry in intset: %w", err)
	}

	i.index++
	i.values = append(i.values, value)

	return entry, nil
}

func (i *intSetIterator) readValue() (interface{}, error) {
	switch i.encoding {
	case 8:
		return reader.ReadInt64(i.buf)
	case 4:
		return reader.ReadInt32(i.buf)
	case 2:
		return reader.ReadInt16(i.buf)
	}

	return nil, IntSetEncodingError{Encoding: i.encoding}
}
