package rdb

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

const (
	len6Bit   = 0
	len14Bit  = 1
	len32Bit  = 0x80
	len64Bit  = 0x81
	lenEncVal = 3

	opCodeModuleAux    = 247
	opCodeIdle         = 248
	opCodeFreq         = 249
	opCodeAux          = 250
	opCodeResizeDB     = 251
	opCodeExpireTimeMS = 252
	opCodeExpireTime   = 253
	opCodeSelectDB     = 254
	opCodeEOF          = 255

	typeString          = 0
	typeList            = 1
	typeSet             = 2
	typeZSet            = 3
	typeHash            = 4
	typeZSet2           = 5
	typeModule          = 6
	typeModule2         = 7
	typeHashZipMap      = 9
	typeListZipList     = 10
	typeSetIntSet       = 11
	typeZSetZipList     = 12
	typeHashZipList     = 13
	typeListQuickList   = 14
	typeStreamListPacks = 15

	encInt8  = 0
	encInt16 = 1
	encInt32 = 2
	encLZF   = 3
)

var (
	magicString = []byte("REDIS")

	minVersion = 1
	maxVersion = 9
)

type Parser struct {
	reader      io.Reader
	db          int64
	initialized bool
	expiry      *time.Time
	dataType    *byte
	idle        int64
	freq        byte
	key         string
	iterator    iterator
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		reader: r,
		db:     -1,
	}
}

func (p *Parser) Next() (interface{}, error) {
	if !p.initialized {
		if err := p.verifyMagicString(); err != nil {
			return nil, err
		}

		if err := p.verifyVersion(); err != nil {
			return nil, err
		}

		p.initialized = true
	}

	if p.dataType != nil {
		return p.readData()
	}

	dataType, err := readByte(p.reader)

	if err != nil {
		return nil, err
	}

	switch dataType {
	case opCodeExpireTimeMS:
		if p.expiry, err = readMillisecondsTime(p.reader); err != nil {
			return nil, err
		}

		return p.Next()

	case opCodeExpireTime:
		if p.expiry, err = readSecondsTime(p.reader); err != nil {
			return nil, err
		}

		return p.Next()

	case opCodeIdle:
		if p.idle, err = readLength(p.reader); err != nil {
			return nil, err
		}

		return p.Next()

	case opCodeFreq:
		if p.freq, err = readByte(p.reader); err != nil {
			return nil, err
		}

		return p.Next()

	case opCodeSelectDB:
		if p.db, err = readLength(p.reader); err != nil {
			return nil, err
		}

		return p.Next()

	case opCodeAux:
		key, err := readString(p.reader)

		if err != nil {
			return nil, err
		}

		value, err := readString(p.reader)

		if err != nil {
			return nil, err
		}

		return &Aux{Key: key, Value: value}, nil

	case opCodeResizeDB:
		// TODO

	case opCodeModuleAux:
		// TODO

	case opCodeEOF:
		// TODO: verify checksum
		return nil, io.EOF
	}

	if p.key, err = readString(p.reader); err != nil {
		return nil, err
	}

	p.dataType = &dataType
	return p.Next()
}

func (p *Parser) verifyMagicString() error {
	buf := make([]byte, len(magicString))

	if _, err := io.ReadFull(p.reader, buf); err != nil {
		return err
	}

	if !bytes.Equal(buf, magicString) {
		return errors.New("invalid magic string")
	}

	return nil
}

func (p *Parser) verifyVersion() error {
	buf := make([]byte, 4)

	if _, err := io.ReadFull(p.reader, buf); err != nil {
		return err
	}

	version, err := strconv.Atoi(string(buf))

	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	if version < minVersion || version > maxVersion {
		return fmt.Errorf("unsupported version %d", version)
	}

	return nil
}

func (p *Parser) readData() (interface{}, error) {
	if p.iterator != nil {
		value, err := p.iterator.Next()

		if err == io.EOF {
			p.dataType = nil
			p.iterator = nil
			return p.Next()
		}

		if err != nil {
			return nil, err
		}

		return value, nil
	}

	key := DataKey{
		Key:      p.key,
		Expiry:   p.expiry,
		Database: p.db,
	}

	switch *p.dataType {
	case typeString:
		value, err := readString(p.reader)

		if err != nil {
			return nil, err
		}

		p.dataType = nil
		return &StringData{DataKey: key, Value: value}, nil

	case typeList:
		p.iterator = &seqIterator{
			DataKey:     key,
			Reader:      p.reader,
			ValueReader: stringValueReader{},
			Mapper:      listMapper{},
		}

		return p.Next()

	case typeSet:
		p.iterator = &seqIterator{
			DataKey:     key,
			Reader:      p.reader,
			ValueReader: stringValueReader{},
			Mapper:      setMapper{},
		}

		return p.Next()

	case typeZSet, typeZSet2:
		p.iterator = &seqIterator{
			DataKey:     key,
			Reader:      p.reader,
			ValueReader: sortedSetValueReader{Type: *p.dataType},
			Mapper:      sortedSetMapper{},
		}

		return p.Next()

	case typeHash:
		p.iterator = &seqIterator{
			DataKey:     key,
			Reader:      p.reader,
			ValueReader: hashValueReader{},
			Mapper:      hashMapper{},
		}

		return p.Next()

	case typeHashZipMap:
		p.iterator = &zipMapIterator{
			DataKey: key,
			Reader:  p.reader,
			Mapper:  hashMapper{},
		}

		return p.Next()
	}

	return nil, fmt.Errorf("unsupported data type %d", p.dataType)
}