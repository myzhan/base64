package base64

/*
#cgo CFLAGS: -I../deps/base64/include
#cgo LDFLAGS: -L../deps/base64/lib -lbase64

#include "libbase64.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

// The values below force the use of a given codec, even if that codec
// is a no-op in the current build. Used in testing. Set to 0 for the
// default behavior, which is runtime feature detection on x86, a
// compile-time fixed codec on ARM, and the plain codec on other platforms.
const (
	Base64ForceAVX2 = 1 << iota
	Base64ForceNeon32
	Base64ForceNeon64
	Base64ForcePlain
	Base64ForceSSE3
	Base64ForceSSE41
	Base64ForceSSE42
	Base64ForceAVX
)

// ErrInvalidInput means that the input is invalid while decoding.
var ErrInvalidInput = errors.New("invalid input")

// ErrMissingCodec means that the chosen codec is not included in the current build.
var ErrMissingCodec = errors.New("the chosen codec is missing in the current build")

// ErrUnknown means something unknown happens, you may report an issue.
var ErrUnknown = errors.New("unknown error")

// A Codec is a base64 codec using the standard encoding with padding.
// It behaves like Go's builtin base64.StdEncoding.
type Codec struct {
	flag int
}

func isCodecSupported(flag int) bool {
	testString := "aGVsbG8="
	srcSize := len(testString)
	dstBuff := make([]byte, 10)
	var outSize int
	// Check if given codec is supported by trying to decode a test string.
	ret := C.base64_decode((*C.char)(unsafe.Pointer(&([]byte(testString)[0]))),
		C.size_t(srcSize),
		(*C.char)(unsafe.Pointer(&dstBuff[0])),
		(*C.size_t)(unsafe.Pointer(&outSize)),
		C.int(flag))
	return ret != -1
}

// DefaultCodec will choose the underlying codec at runtime.
var DefaultCodec = NewCodec(0)

// NewCodec creates a codec by specifying the flag.
func NewCodec(flag int) *Codec {
	if !isCodecSupported(flag) {
		panic(ErrMissingCodec.Error())
	}
	return &Codec{
		flag: flag,
	}
}

// EncodedLen returns the length in bytes of the base64 encoding
// of an input buffer of length n.
func (c *Codec) EncodedLen(n int) int {
	return (n + 2) / 3 * 4
}

// EncodeToString returns the base64 encoding of src.
func (c *Codec) EncodeToString(src []byte) string {
	srcSize := len(src)
	if srcSize == 0 {
		return ""
	}

	dstSize := c.EncodedLen(srcSize)
	var outSize int
	dstBuff := make([]byte, dstSize)
	C.base64_encode((*C.char)(unsafe.Pointer(&src[0])),
		C.size_t(srcSize),
		(*C.char)(unsafe.Pointer(&dstBuff[0])),
		(*C.size_t)(unsafe.Pointer(&outSize)),
		C.int(c.flag))

	return string(dstBuff[:outSize])
}

// DecodedLen returns the maximum length in bytes of the decoded data
// corresponding to n bytes of base64-encoded data.
func (c *Codec) DecodedLen(n int) int {
	// Padded base64 should always be a multiple of 4 characters in length.
	return n / 4 * 3
}

// DecodeString returns the bytes represented by the base64 string s.
func (c *Codec) DecodeString(src string) ([]byte, error) {
	srcSize := len(src)
	if srcSize == 0 {
		return []byte(""), nil
	}

	var outSize int
	dstSize := c.DecodedLen(srcSize)
	dstBuff := make([]byte, dstSize)
	ret := C.base64_decode(
		(*C.char)(unsafe.Pointer(&([]byte(src)[0]))),
		C.size_t(srcSize),
		(*C.char)(unsafe.Pointer(&dstBuff[0])),
		(*C.size_t)(unsafe.Pointer(&outSize)),
		C.int(c.flag))

	switch ret {
	case 1:
		return dstBuff[:outSize], nil
	case 0:
		return nil, ErrInvalidInput
	case -1:
		return nil, ErrMissingCodec
	default:
		return nil, ErrUnknown
	}
}
