package gostreamencoder

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

const maxblocksize = 1024

const (
	finfoblock = iota
	datablock
	endblock
)

type Encoder struct {
	input  io.Reader
	output io.Writer
	block  block
}
type Decoder struct {
	input io.Reader
	block block
}
type StreamFile struct {
	Name string `json:"Name"`
	Size int64  `json:"Size"`
}

const blockHeaderSize = 6

type block struct {
	blocktype uint16
	blocksize uint32
	data      []byte
}
type ReadError struct {
}

func (re ReadError) Error() string {
	return "todo"
}

type BlockSizeOverflowError struct {
}

func (bsoe BlockSizeOverflowError) Error() string {
	return "Block Size too small"
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		output: w,
		block: block{
			data: make([]byte, maxblocksize),
		},
	}
}
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		input: r,
		block: block{
			data: make([]byte, maxblocksize),
		},
	}
}
func (b block) encodeHeader() []byte {
	header := make([]byte, blockHeaderSize)
	binary.BigEndian.PutUint16(header, b.blocktype)
	binary.BigEndian.PutUint32(header[2:], b.blocksize)
	return header

}

func (b block) write(w io.Writer) error {
	header := b.encodeHeader()
	_, err := w.Write(header)
	if err != nil {
		return err
	}
	_, err = w.Write(b.data)
	return err
}
func (b *block) decodeBlockHeader(header []byte) {
	b.blocktype = binary.BigEndian.Uint16(header)
	b.blocksize = binary.BigEndian.Uint32(header[2:])

}
func (d *Decoder) DecodeStream(w io.Writer) error {
	header := make([]byte, blockHeaderSize)
	while := true
	for while {
		n, err := d.input.Read(header)
		if err != nil {
			return err
		}
		if n < len(header) {
			return BlockSizeOverflowError{}
		}
		d.block.decodeBlockHeader(header)
		switch d.block.blocktype {
		case datablock:
		case endblock:
			while = false
		default:
			return ReadError{}

		}

		n, err = d.input.Read(d.block.data[:d.block.blocksize])
		if err != nil {
			return err
		}
		if n < int(d.block.blocksize) {
			return BlockSizeOverflowError{}
		}
		_, err = w.Write(d.block.data[:d.block.blocksize])
		if err != nil {
			return BlockSizeOverflowError{}
		}

	}
	return nil
}

func (d *Decoder) DecodeFinfo() (StreamFile, error) {
	sf := StreamFile{}
	header := make([]byte, blockHeaderSize)
	n, err := d.input.Read(header)
	if err != nil {
		return sf, err
	}
	if n < blockHeaderSize {
		return sf, ReadError{}
	}
	d.block.decodeBlockHeader(header)
	if d.block.blocktype != finfoblock {
		return sf, ReadError{}
	}
	n, err = d.input.Read(d.block.data[:d.block.blocksize])
	if err != nil {
		return sf, err
	}
	if n < int(d.block.blocksize) {
		return sf, ReadError{}
	}
	err = json.Unmarshal(d.block.data[:d.block.blocksize], &sf)
	return sf, err

}

func (e *Encoder) EncodeFinfo(name string, size int64) error {
	sf := StreamFile{Name: name, Size: size}
	data, err := json.Marshal(sf)
	if err != nil {
		return err
	}
	if len(data) > maxblocksize {
		return BlockSizeOverflowError{}
	}
	b := block{
		blocktype: finfoblock,
		blocksize: uint32(len(data)),
		data:      data,
	}
	return b.write(e.output)
	//return nil
}

func (e *Encoder) EncodeStream(r io.Reader) error {

	e.block.blocktype = datablock
	while := true
	header := e.block.encodeHeader()
	for while {
		n, err := r.Read(e.block.data)
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			e.block.blocktype = endblock
			binary.BigEndian.PutUint16(header, e.block.blocktype)
			while = false
		}

		e.block.blocksize = uint32(n)
		binary.BigEndian.PutUint32(header[2:], e.block.blocksize)
		_, err = e.output.Write(header)
		if err != nil {
			return err
		}
		_, err = e.output.Write(e.block.data[:n])
		if err != nil {
			return err
		}

	}
	return nil

}
