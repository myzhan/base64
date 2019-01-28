package base64

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func testEqual(t *testing.T, msg string, args ...interface{}) bool {
	if args[len(args)-2] != args[len(args)-1] {
		t.Errorf(msg, args...)
		return false
	}
	return true
}

func TestFlags(t *testing.T) {
	testEqual(t, "Base64ForceAVX2 = %q, want %q", Base64ForceAVX2, 1<<0)
	testEqual(t, "Base64ForceNeon32 = %q, want %q", Base64ForceNeon32, 1<<1)
	testEqual(t, "Base64ForceNeon64 = %q, want %q", Base64ForceNeon64, 1<<2)
	testEqual(t, "Base64ForcePlain = %q, want %q", Base64ForcePlain, 1<<3)
	testEqual(t, "Base64ForceSSSE3 = %q, want %q", Base64ForceSSSE3, 1<<4)
	testEqual(t, "Base64ForceSSE41 = %q, want %q", Base64ForceSSE41, 1<<5)
	testEqual(t, "Base64ForceSSE42 = %q, want %q", Base64ForceSSE42, 1<<6)
	testEqual(t, "Base64ForceAVX = %q, want %q", Base64ForceAVX, 1<<7)
}

type testpair struct {
	decoded, encoded string
}

var pairs = []testpair{
	// RFC 3548 examples
	{"\x14\xfb\x9c\x03\xd9\x7e", "FPucA9l+"},
	{"\x14\xfb\x9c\x03\xd9", "FPucA9k="},
	{"\x14\xfb\x9c\x03", "FPucAw=="},

	// RFC 4648 examples
	{"", ""},
	{"f", "Zg=="},
	{"fo", "Zm8="},
	{"foo", "Zm9v"},
	{"foob", "Zm9vYg=="},
	{"fooba", "Zm9vYmE="},
	{"foobar", "Zm9vYmFy"},

	// Wikipedia examples
	{"sure.", "c3VyZS4="},
	{"sure", "c3VyZQ=="},
	{"sur", "c3Vy"},
	{"su", "c3U="},
	{"leasure.", "bGVhc3VyZS4="},
	{"easure.", "ZWFzdXJlLg=="},
	{"asure.", "YXN1cmUu"},
	{"sure.", "c3VyZS4="},
}

func TestStreamEncodeInit(t *testing.T) {
	state := &State{}
	codec := NewCodec(0)
	codec.StreamEncodeInit(state)
}

func TestStreamEncode(t *testing.T) {
	state := &State{}
	codec := NewCodec(0)
	codec.StreamEncodeInit(state)

	file, err := os.Open("testdata/moby_dick_plain.txt")
	defer file.Close()
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	var encodedBytes []byte
	expectedEncodeBytes, err := ioutil.ReadFile("testdata/moby_dick_base64.txt")
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	maxChunkSize := 10
	encodedChunkSize := codec.encodedLen(maxChunkSize)

	var outSize int
	readBuff := make([]byte, maxChunkSize)
	outBuff := make([]byte, encodedChunkSize)

	for {
		nread, err := file.Read(readBuff)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Errorf("%v\n", err)
				return
			}
		}
		codec.StreamEncode(state, readBuff, nread, outBuff, &outSize)
		encodedBytes = append(encodedBytes, outBuff[:outSize]...)
	}
	codec.StreamEncodeFinal(state, outBuff, &outSize)
	if outSize > 0 {
		// write trailer if any
		encodedBytes = append(encodedBytes, outBuff[:outSize]...)
	}
	if !bytes.Equal(expectedEncodeBytes, encodedBytes) {
		t.Error("not equal")
	}
}

func TestStreamDecode(t *testing.T) {
	state := &State{}
	codec := NewCodec(0)
	codec.StreamDecodeInit(state)

	file, err := os.Open("testdata/moby_dick_base64.txt")
	defer file.Close()
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	var decodedBytes []byte
	expectedDecodeBytes, err := ioutil.ReadFile("testdata/moby_dick_plain.txt")
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	maxChunkSize := 10
	decodedChunkSize := codec.decodedLen(maxChunkSize)

	var outSize int
	readBuff := make([]byte, maxChunkSize)
	outBuff := make([]byte, decodedChunkSize)

	for {
		nread, err := file.Read(readBuff)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Errorf("%v\n", err)
				return
			}
		}
		err = codec.StreamDecode(state, readBuff, nread, outBuff, &outSize)
		if err != nil {
			t.Errorf("%v\n", err)
			return
		}
		decodedBytes = append(decodedBytes, outBuff[:outSize]...)
	}
	if !bytes.Equal(expectedDecodeBytes, decodedBytes) {
		t.Error("not equal")
	}
}

func TestEncodeToString(t *testing.T) {
	for _, pair := range pairs {
		got := DefaultCodec.EncodeToString([]byte(pair.decoded))
		if got != pair.encoded {
			t.Errorf("Encode(%q) = %q, want %q", pair.decoded, got, pair.encoded)
		}
	}
}

func TestDecodeString(t *testing.T) {
	for _, pair := range pairs {
		got, err := DefaultCodec.DecodeString(pair.encoded)
		if err != nil {
			t.Errorf("Decode(%q) error, %v", pair.encoded, err)
		}
		if !bytes.Equal(got, []byte(pair.decoded)) {
			t.Errorf("Decode(%q) = %q, want %q", pair.encoded, got, pair.decoded)
		}
	}
}

func IgnoredTestNewLineCharacters(t *testing.T) {
	// Each of these should decode to the string "sure", without errors.
	const expected = "sure"
	examples := []string{
		"c3VyZQ==",
		"c3VyZQ==\r",
		"c3VyZQ==\n",
		"c3VyZQ==\r\n",
		"c3VyZ\r\nQ==",
		"c3V\ryZ\nQ==",
		"c3V\nyZ\rQ==",
		"c3VyZ\nQ==",
		"c3VyZQ\n==",
		"c3VyZQ=\n=",
		"c3VyZQ=\r\n\r\n=",
	}
	codec := NewCodec(Base64ForceAVX)
	for _, e := range examples {
		log.Println("decoding:", e)
		buf, err := codec.DecodeString(e)
		if err != nil {
			t.Errorf("Decode(%q) failed: %v", e, err)
			continue
		}
		if s := string(buf); s != expected {
			t.Errorf("Decode(%q) = %q, want %q", e, s, expected)
		}
	}
}

func BenchmarkEncodeToString(b *testing.B) {
	sizes := []int{2, 4, 8, 64, 8192}
	codecs := map[string]*Codec{
		"avx2": NewCodec(Base64ForceAVX2),
		// "neon32": NewCodec(Base64ForceNeon32),
		// "neon64": NewCodec(Base64ForceNeon64),
		"plain": NewCodec(Base64ForcePlain),
		"ssse3": NewCodec(Base64ForceSSSE3),
		"sse41": NewCodec(Base64ForceSSE41),
		"sse42": NewCodec(Base64ForceSSE42),
		"avx":   NewCodec(Base64ForceAVX),
	}
	benchFunc := func(b *testing.B, codec *Codec, benchSize int) {
		data := make([]byte, benchSize)
		b.ResetTimer()
		b.SetBytes(int64(benchSize))
		for i := 0; i < b.N; i++ {
			codec.EncodeToString(data)
		}
	}
	for name, codec := range codecs {
		for _, size := range sizes {
			b.Run(fmt.Sprintf("%s-%d", name, size), func(b *testing.B) {
				benchFunc(b, codec, size)
			})
		}
	}
}

func BenchmarkDecodeString(b *testing.B) {
	sizes := []int{2, 4, 8, 64, 8192}
	codecs := map[string]*Codec{
		"avx2": NewCodec(Base64ForceAVX2),
		// "neon32": NewCodec(Base64ForceNeon32),
		// "neon64": NewCodec(Base64ForceNeon64),
		"plain": NewCodec(Base64ForcePlain),
		"ssse3": NewCodec(Base64ForceSSSE3),
		"sse41": NewCodec(Base64ForceSSE41),
		"sse42": NewCodec(Base64ForceSSE42),
		"avx":   NewCodec(Base64ForceAVX),
	}
	benchFunc := func(b *testing.B, codec *Codec, benchSize int) {
		data := DefaultCodec.EncodeToString(make([]byte, benchSize))
		b.ResetTimer()
		b.SetBytes(int64(benchSize))
		for i := 0; i < b.N; i++ {
			codec.DecodeString(data)
		}
	}
	for name, codec := range codecs {
		for _, size := range sizes {
			b.Run(fmt.Sprintf("%s-%d", name, size), func(b *testing.B) {
				benchFunc(b, codec, size)
			})
		}
	}
}
