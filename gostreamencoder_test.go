package gostreamencoder

import (
	"bytes"
	"crypto/sha1"
	"math/rand"
	"testing"
)

type fakeReader struct {
	rnd *rand.Rand
}

func newFakeReader() *fakeReader {
	return &fakeReader{
		rnd: rand.New(rand.NewSource(100)),
	}
}
func (fr *fakeReader) Read(b []byte) (int, error) {
	return fr.rnd.Read(b)
}
func TestFifoEncoding(t *testing.T) {
	buffer := &bytes.Buffer{}
	encoder := NewEncoder(buffer)
	err := encoder.EncodeFinfo("test", 10241024)
	if err != nil {
		t.Error(err)
	}

	decoder := NewDecoder(bytes.NewReader(buffer.Bytes()))
	sf, err := decoder.DecodeFinfo()
	if err != nil {
		t.Error(err)
	}
	if sf.Name != "test" || sf.Size != 10241024 {
		t.Errorf("Wrong decoded val %s,%d", sf.Name, sf.Size)
	}

}

func TestStreamEncoding(t *testing.T) {
	block := make([]byte, 32*1024)
	fr := newFakeReader()
	fr.Read(block)
	t.Log(sha1.Sum(block))
	block2 := make([]byte, 0, 32*1024+4*1024)
	buffer := bytes.NewBuffer(block2)
	enc := NewEncoder(buffer)
	err := enc.EncodeStream(bytes.NewReader(block))
	if err != nil {
		t.Error(err)
	}
	block3 := make([]byte, 0, 32*1024)
	buffer2 := bytes.NewBuffer(block3)
	decoder := NewDecoder(buffer)
	err = decoder.DecodeStream(buffer2)
	if err != nil {
		t.Error(err)
	}
	block4 := buffer2.Bytes()[:len(block)]
	t.Log(sha1.Sum(block4))
	if !bytes.Equal(block, block4) {
		t.Errorf("Bytes not equal %x,%x\n", block, buffer2.Bytes()[:len(block)])
	}

}
