package main

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"strings"
)

func getVersion(file *os.File) int {
	var version uint64
	binary.Read(file, byteOrder, &version)
	file.Seek(0, os.SEEK_SET)
	if version > 1000 {
		version = 0
	}
	return int(version)
}

func convertFile(inFile *os.File, out io.Writer) error {
	inFile.Seek(0, os.SEEK_SET)

	h0 := new(Ver0Header)
	if err := binary.Read(inFile, byteOrder, h0); err != nil {
		return err
	}

	inFile.Seek(int64(h0.HeaderStart), os.SEEK_SET)
	htr := bufio.NewReader(io.MultiReader(
		io.LimitReader(inFile, int64(h0.BodyStart-h0.HeaderStart)),
		strings.NewReader("\r\n"),
	))
	htr.ReadBytes('\n') // skippng http response code line
	headers, err := textproto.NewReader(htr).ReadMIMEHeader()
	if err != nil {
		return err
	}

	h3 := new(Ver3Header)
	sizeShift := ver3HEaderSize - ver0HEaderSize
	h3.Version = 3
	h3.ValidSec = h0.ValidSec
	h3.LastModified = h0.LastModified
	h3.Date = h0.Date
	h3.CRC32 = h0.CRC32
	h3.ValidMsec = h0.ValidMsec
	h3.HeaderStart = uint16(int(h0.HeaderStart) + sizeShift)
	h3.BodyStart = uint16(int(h0.BodyStart) + sizeShift)

	if s := headers.Get("Etag"); s != "" {
		h3.ETagLen = byte(len(s))
		copy(h3.ETag[:], []byte(s))
	}
	if s := headers.Get("Vary"); s != "" {
		h3.VaryLen = byte(len(s))
		copy(h3.Vary[:], []byte(s))

		if len(s) > NGX_HTTP_CACHE_VARY_LEN || s == "*" {
			// особый случай, nginx такое не кэширует
			return fmt.Errorf("Non-cacheable Vary header")
		}

		// variant = hash of all Vary headers, see ngx_http_file_cache_vary
		hash := md5.New()
		for _, hName := range strings.Split(s, ",") {
			hName = strings.TrimSpace(hName)
			io.WriteString(hash, strings.ToLower(hName))
			io.WriteString(hash, ":")
			io.WriteString(hash, headers.Get(hName))
			io.WriteString(hash, "\r\n")
		}
		copy(h3.Variant[:], hash.Sum(nil))
	}

	inFile.Seek(int64(ver0HEaderSize), os.SEEK_SET)
	binary.Write(out, byteOrder, h3)
	io.Copy(out, inFile)

	return nil
}
