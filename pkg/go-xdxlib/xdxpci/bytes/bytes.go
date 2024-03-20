 package bytes

 import (
	 "encoding/binary"
	 "unsafe"
 )
 
 // Raw returns just the bytes without any assumptions about layout
 type Raw interface {
	 Raw() *[]byte
 }
 
 // Reader used to read various data sizes in the byte array
 type Reader interface {
	 Read8(pos int) uint8
	 Read16(pos int) uint16
	 Read32(pos int) uint32
	 Read64(pos int) uint64
	 Len() int
 }
 
 // Writer used to write various sizes of data in the byte array
 type Writer interface {
	 Write8(pos int, value uint8)
	 Write16(pos int, value uint16)
	 Write32(pos int, value uint32)
	 Write64(pos int, value uint64)
	 Len() int
 }
 
 // Bytes object for manipulating arbitrary byte arrays
 type Bytes interface {
	 Raw
	 Reader
	 Writer
	 Slice(offset int, size int) Bytes
	 LittleEndian() Bytes
	 BigEndian() Bytes
 }
 
 var nativeByteOrder binary.ByteOrder
 
 func init() {
	 buf := [2]byte{}
	 *(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0x00FF)
 
	 switch buf {
	 case [2]byte{0xFF, 0x00}:
		 nativeByteOrder = binary.LittleEndian
	 case [2]byte{0x00, 0xFF}:
		 nativeByteOrder = binary.BigEndian
	 default:
		 panic("Unable to infer byte order")
	 }
 }
 
 // New raw bytearray
 func New(data *[]byte) Bytes {
	 return (*native)(data)
 }
 
 // NewLittleEndian little endian ordering of bytes
 func NewLittleEndian(data *[]byte) Bytes {
	 if nativeByteOrder == binary.LittleEndian {
		 return (*native)(data)
	 }
 
	 return (*swapbo)(data)
 }
 
 // NewBigEndian big endian ordering of bytes
 func NewBigEndian(data *[]byte) Bytes {
	 if nativeByteOrder == binary.BigEndian {
		 return (*native)(data)
	 }
 
	 return (*swapbo)(data)
 }
 