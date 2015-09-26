package main

import (
	"encoding/binary"
	"unsafe"
)

const (
	MAX_VERSION = 3

	NGX_HTTP_CACHE_KEY_LEN  = 16
	NGX_HTTP_CACHE_ETAG_LEN = 42
	NGX_HTTP_CACHE_VARY_LEN = 42
)

var (
	byteOrder      = binary.LittleEndian
	ver0HEaderSize = int(unsafe.Sizeof(Ver0Header{}))
	ver3HEaderSize = int(unsafe.Sizeof(Ver3Header{}))
)

type Ver0Header struct {
	ValidSec     uint64
	LastModified uint64
	Date         uint64
	CRC32        uint32
	ValidMsec    uint16
	HeaderStart  uint16
	BodyStart    uint16
}

type Ver3Header struct {
	Version      uint64
	ValidSec     uint64
	LastModified uint64
	Date         uint64
	CRC32        uint32
	ValidMsec    uint16
	HeaderStart  uint16
	BodyStart    uint16
	ETagLen      byte
	ETag         [NGX_HTTP_CACHE_ETAG_LEN]byte
	VaryLen      byte
	Vary         [NGX_HTTP_CACHE_VARY_LEN]byte
	Variant      [NGX_HTTP_CACHE_KEY_LEN]byte
}
