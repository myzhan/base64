package main

import (
	"flag"
	"io"
	"log"
	"os"

	base64 "github.com/myzhan/base64/pkg"
)

var decodeMode = false

func encode() {
	state := &base64.State{}
	codec := base64.NewCodec(0)
	codec.StreamEncodeInit(state)

	maxChunkSize := 100
	encodedChunkSize := codec.EncodedLen(maxChunkSize)

	var outSize int
	readBuff := make([]byte, maxChunkSize)
	outBuff := make([]byte, encodedChunkSize)

	in := os.Stdin
	out := os.Stdout

	for {
		nread, err := in.Read(readBuff)
		if err != nil {
			if err != io.EOF {
				log.Printf("%v\n", err)
			}
			break
		}
		codec.StreamEncode(state, readBuff, nread, outBuff, &outSize)
		out.Write(outBuff[:outSize])
	}

	codec.StreamEncodeFinal(state, outBuff, &outSize)
	if outSize > 0 {
		// write trailer if any
		out.Write(outBuff[:outSize])
	}
}

func decode() {
	state := &base64.State{}
	codec := base64.NewCodec(0)
	codec.StreamDecodeInit(state)

	maxChunkSize := 100
	encodedChunkSize := codec.DecodedLen(maxChunkSize)

	var outSize int
	readBuff := make([]byte, maxChunkSize)
	outBuff := make([]byte, encodedChunkSize)

	in := os.Stdin
	out := os.Stdout

	for {
		nread, err := in.Read(readBuff)
		if err != nil {
			if err != io.EOF {
				log.Printf("%v\n", err)
			}
			break
		}
		err = codec.StreamDecode(state, readBuff, nread, outBuff, &outSize)
		if err != nil {
			log.Printf("%v\n", err)
		} else {
			out.Write(outBuff[:outSize])
		}
	}
}

func main() {
	flag.BoolVar(&decodeMode, "d", false, "Run in decode mode.")
	flag.Parse()

	if decodeMode {
		decode()
	} else {
		encode()
	}
}
