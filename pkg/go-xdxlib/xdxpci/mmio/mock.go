package mmio

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/xdxpci/bytes"
)

type mockMmio struct {
	mmio
	source *[]byte
	offset int
	rw     bool
}

func mockOpen(source *[]byte, offset int, size int, rw bool) (Mmio, error) {
	if size < 0 {
		size = len(*source) - offset
	}
	if (offset + size) > len(*source) {
		return nil, fmt.Errorf("offset+size out of range")
	}

	data := append([]byte{}, (*source)[offset:offset+size]...)

	m := &mockMmio{}
	m.Bytes = bytes.New(&data).LittleEndian()
	m.source = source
	m.offset = offset
	m.rw = rw

	return m, nil
}

// MockOpenRO open read only
func MockOpenRO(source *[]byte, offset int, size int) (Mmio, error) {
	return mockOpen(source, offset, size, false)
}

// MockOpenRW open read write
func MockOpenRW(source *[]byte, offset int, size int) (Mmio, error) {
	return mockOpen(source, offset, size, true)
}

func (m *mockMmio) Close() error {
	m = &mockMmio{}
	return nil
}

func (m *mockMmio) Sync() error {
	if !m.rw {
		return fmt.Errorf("opened read-only")
	}
	for i := range *m.Bytes.Raw() {
		(*m.source)[m.offset+i] = (*m.Bytes.Raw())[i]
	}
	return nil
}
