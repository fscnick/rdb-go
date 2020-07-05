package rdb

import (
	"errors"
	"fmt"
)

var ErrInvalidMagicString = errors.New("invalid magic string")

type UnsupportedVersionError struct {
	Version int
}

func (u UnsupportedVersionError) Error() string {
	return fmt.Sprintf("unsupported version %d", u.Version)
}

type IntSetEncodingError struct {
	Encoding uint32
}

func (i IntSetEncodingError) Error() string {
	return fmt.Sprintf("invalid intset encoding %d", i.Encoding)
}

type LengthEncodingError struct {
	Encoding byte
}

func (l LengthEncodingError) Error() string {
	return fmt.Sprintf("invalid length encoding %d", l.Encoding)
}

type StringEncodingError struct {
	Encoding int64
}

func (s StringEncodingError) Error() string {
	return fmt.Sprintf("invalid string encoding %d", s.Encoding)
}

type UnsupportedDataTypeError struct {
	DataType byte
}

func (u UnsupportedDataTypeError) Error() string {
	return fmt.Sprintf("unsupported data type %d", u.DataType)
}

type UnexpectedZipMapEndError struct {
	Key string
}

func (u UnexpectedZipMapEndError) Error() string {
	return fmt.Sprintf("unexpected end of zipmap for key %q", u.Key)
}

type ZipListHeaderError struct {
	Header byte
}

func (z ZipListHeaderError) Error() string {
	return fmt.Sprintf("invalid ziplist entry header %d", z.Header)
}

type ConvertError struct {
	Value interface{}
	Type  string
}

func (c ConvertError) Error() string {
	return fmt.Sprintf("unable to convert value %v to %s", c.Value, c.Type)
}

type ZipListLengthError struct {
	Length      int64
	ValueLength int64
}

func (z ZipListLengthError) Error() string {
	return fmt.Sprintf("invalid ziplist length %d, expected to be divisible by %d", z.Length, z.ValueLength)
}

type ZipListEndError struct {
	Value byte
}

func (z ZipListEndError) Error() string {
	return fmt.Sprintf("invalid ziplist end %d", z.Value)
}