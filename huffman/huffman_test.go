package huffman_test

import (
	"crypto/rand"
	"huffman/huffman"
	"testing"
)

func compareBytes(first, second []byte) bool {
	if len(first) != len(second) {
		return false
	}

	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}

	return true
}

func TestCompressDecompress(t *testing.T) {

	for exp := 0; exp < 16; exp++ {
		original := make([]byte, 1<<exp)
		rand.Read(original)

		freq := huffman.Frequencies{}
		for _, b := range original {
			freq[b]++
		}

		huffman := huffman.Huffman{}
		huffman.Init(&freq)

		copressBuffer := make([]byte, 1<<(exp+2))
		compressedSize, err := huffman.Compress(original, copressBuffer)
		if err != nil {
			t.Error(err)
		}
		
		decompressBuffer := make([]byte, len(original))
		decompressedSize, err := huffman.Decompress(copressBuffer[:compressedSize], decompressBuffer)
		if err != nil {
			t.Error(err)
		}

		if !compareBytes(original, decompressBuffer[:decompressedSize]) {
			t.Errorf("original and decompressed are not equal")
		}
	}
}

func TestCompressDecompressWithoutFrequencies(t *testing.T) {
	huffman := huffman.Huffman{}
	huffman.Init(nil)

	for exp := 0; exp < 16; exp++ {
		original := make([]byte, 1<<exp)
		rand.Read(original)

		copressBuffer := make([]byte, 1<<(exp+2))
		compressedSize, err := huffman.Compress(original, copressBuffer)
		if err != nil {
			t.Error(err)
		}

		decompressBuffer := make([]byte, len(original))
		decompressedSize, err := huffman.Decompress(copressBuffer[:compressedSize], decompressBuffer)
		if err != nil {
			t.Error(err)
		}

		if !compareBytes(original, decompressBuffer[:decompressedSize]) {
			t.Errorf("original and decompressed are not equal")
		}
	}
}
