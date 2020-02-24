package huffman

import "fmt"

const (
	HUFFMAN_EOF_SYMBOL int = 256

	HUFFMAN_MAX_SYMBOLS int = HUFFMAN_EOF_SYMBOL + 1
	HUFFMAN_MAX_NODES   int = HUFFMAN_MAX_SYMBOLS*2 - 1

	HUFFMAN_LUTBITS int = 10
	HUFFMAN_LUTSIZE int = (1 << HUFFMAN_LUTBITS)
	HUFFMAN_LUTMASK int = (HUFFMAN_LUTSIZE - 1)
)

type Node struct {
	Bits    uint
	NumBits uint
	Leafs   [2]uint16
	Symbol  byte
}

type Huffman struct {
	Nodes     [HUFFMAN_MAX_NODES]Node
	DecodeLut [HUFFMAN_LUTSIZE]*Node
	StartNode *Node
	NumNodes  int
}

type Frequencies = [256 + 1]uint

var FreqTable Frequencies = Frequencies{
	1 << 30, 4545, 2657, 431, 1950, 919, 444, 482, 2244, 617, 838, 542, 715, 1814, 304, 240, 754, 212, 647, 186,
	283, 131, 146, 166, 543, 164, 167, 136, 179, 859, 363, 113, 157, 154, 204, 108, 137, 180, 202, 176,
	872, 404, 168, 134, 151, 111, 113, 109, 120, 126, 129, 100, 41, 20, 16, 22, 18, 18, 17, 19,
	16, 37, 13, 21, 362, 166, 99, 78, 95, 88, 81, 70, 83, 284, 91, 187, 77, 68, 52, 68,
	59, 66, 61, 638, 71, 157, 50, 46, 69, 43, 11, 24, 13, 19, 10, 12, 12, 20, 14, 9,
	20, 20, 10, 10, 15, 15, 12, 12, 7, 19, 15, 14, 13, 18, 35, 19, 17, 14, 8, 5,
	15, 17, 9, 15, 14, 18, 8, 10, 2173, 134, 157, 68, 188, 60, 170, 60, 194, 62, 175, 71,
	148, 67, 167, 78, 211, 67, 156, 69, 1674, 90, 174, 53, 147, 89, 181, 51, 174, 63, 163, 80,
	167, 94, 128, 122, 223, 153, 218, 77, 200, 110, 190, 73, 174, 69, 145, 66, 277, 143, 141, 60,
	136, 53, 180, 57, 142, 57, 158, 61, 166, 112, 152, 92, 26, 22, 21, 28, 20, 26, 30, 21,
	32, 27, 20, 17, 23, 21, 30, 22, 22, 21, 27, 25, 17, 27, 23, 18, 39, 26, 15, 21,
	12, 18, 18, 27, 20, 18, 15, 19, 11, 17, 33, 12, 18, 15, 19, 18, 16, 26, 17, 18,
	9, 10, 25, 22, 22, 17, 20, 16, 6, 16, 15, 20, 14, 18, 24, 335, 1517}

type HuffmanConstructNode struct {
	NodeId    uint16
	Frequency int
}

// TODO: this should be something faster, but it's enough for now
func bubbleSort(list []*HuffmanConstructNode) {
	size := len(list)
	changed := true
	var temp *HuffmanConstructNode

	for changed {
		changed = false
		for i := 0; i < size-1; i++ {
			if list[i].Frequency < list[i+1].Frequency {
				temp = list[i]
				list[i] = list[i+1]
				list[i+1] = temp
				changed = true
			}
		}
		size--
	}
}

func (self *Huffman) setBits(node *Node, bits uint, depth uint) {
	if node.Leafs[1] != 0xffff {
		self.setBits(&self.Nodes[node.Leafs[1]], bits|(1<<depth), depth+1)
	}
	if node.Leafs[0] != 0xffff {
		self.setBits(&self.Nodes[node.Leafs[0]], bits, depth+1)
	}

	if node.NumBits != 0 {
		node.Bits = bits
		node.NumBits = depth
	}
}

func (self *Huffman) constructTree(frequencies *Frequencies) {
	var nodesLeftStorage [HUFFMAN_MAX_SYMBOLS]HuffmanConstructNode
	var nodesLeft [HUFFMAN_MAX_SYMBOLS]*HuffmanConstructNode
	numNodesLeft := HUFFMAN_MAX_SYMBOLS

	// add the symbols
	for i := int(0); i < HUFFMAN_MAX_SYMBOLS; i++ {
		self.Nodes[i].NumBits = 0xFFFFFFFF
		self.Nodes[i].Symbol = byte(i)
		self.Nodes[i].Leafs[0] = 0xffff
		self.Nodes[i].Leafs[1] = 0xffff

		if i == HUFFMAN_EOF_SYMBOL {
			nodesLeftStorage[i].Frequency = 1
		} else {
			nodesLeftStorage[i].Frequency = int(frequencies[i])
		}
		nodesLeftStorage[i].NodeId = uint16(i)
		nodesLeft[i] = &nodesLeftStorage[i]

	}

	self.NumNodes = HUFFMAN_MAX_SYMBOLS

	// construct the table
	for numNodesLeft > 1 {
		// we can't rely on stdlib's qsort for this, it can generate different results on different implementations
		bubbleSort(nodesLeft[:numNodesLeft])

		self.Nodes[self.NumNodes].NumBits = 0
		self.Nodes[self.NumNodes].Leafs[0] = nodesLeft[numNodesLeft-1].NodeId
		self.Nodes[self.NumNodes].Leafs[1] = nodesLeft[numNodesLeft-2].NodeId
		nodesLeft[numNodesLeft-2].NodeId = uint16(self.NumNodes)
		nodesLeft[numNodesLeft-2].Frequency = nodesLeft[numNodesLeft-1].Frequency + nodesLeft[numNodesLeft-2].Frequency

		self.NumNodes++
		numNodesLeft--
	}

	// set start node
	self.StartNode = &self.Nodes[self.NumNodes-1]

	// build symbol bits
	self.setBits(self.StartNode, 0, 0)
}

func (self *Huffman) Init(frequencies *Frequencies) {
	// make sure to cleanout every thing
	*self = Huffman{}

	// construct the tree
	if frequencies == nil {
		frequencies = &FreqTable
	}
	self.constructTree(frequencies)

	// build decode LUT
	for i := int(0); i < HUFFMAN_LUTSIZE; i++ {
		var bits uint = uint(i)
		var k int
		node := self.StartNode
		for k = 0; k < HUFFMAN_LUTBITS; k++ {
			node = &self.Nodes[node.Leafs[bits&1]]
			bits >>= 1

			if node == nil {
				break
			}

			if node.NumBits != 0 {
				self.DecodeLut[i] = node
				break
			}
		}

		if k == HUFFMAN_LUTBITS {
			self.DecodeLut[i] = node
		}
	}

}

func (self *Huffman) Compress(input []byte, output []byte) (int, error) {
	inputSize := len(input)
	outputSize := len(output)

	// setup buffer pointers
	// converted to indices for co
	iSrc := 0
	iSrcEnd := inputSize
	iDst := 0
	iDstEnd := outputSize

	// symbol variables
	var bits uint = 0
	var bitcount uint = 0

	// this macro loads a symbol for a byte into bits and bitcount
	// converted to func for go
	HUFFMAN_MACRO_LOADSYMBOL := func(sym int) {
		bits |= self.Nodes[sym].Bits << bitcount
		bitcount += self.Nodes[sym].NumBits
	}

	// this macro writes the symbol stored in bits and bitcount to the dst pointer
	// converted to func for go
	HUFFMAN_MACRO_WRITE := func() {
		for bitcount >= 8 {
			output[iDst] = byte(bits & 0xff)
			iDst++
			if iDst == iDstEnd {
				panic(fmt.Errorf("Unexpected end of output buffer"))
			}
			bits >>= 8
			bitcount -= 8
		}
	}

	// make sure that we have data that we want to compress
	if inputSize != 0 {
		// {A} load the first symbol
		symbol := int(input[iSrc])
		iSrc++

		for iSrc != iSrcEnd {
			// {B} load the symbol
			HUFFMAN_MACRO_LOADSYMBOL(symbol)

			// {C} fetch next symbol, this is done here because it will reduce dependency in the code
			symbol = int(input[iSrc])
			iSrc++

			// {B} write the symbol loaded at
			HUFFMAN_MACRO_WRITE()
		}

		// write the last symbol loaded from {C} or {A} in the case of only 1 byte input buffer
		HUFFMAN_MACRO_LOADSYMBOL(symbol)
		HUFFMAN_MACRO_WRITE()
	}

	// write EOF symbol
	HUFFMAN_MACRO_LOADSYMBOL(HUFFMAN_EOF_SYMBOL)
	HUFFMAN_MACRO_WRITE()

	// write out the last bits
	output[iDst] = byte(bits)
	iDst++

	// return the size of the output
	return iDst, nil
}

func (self *Huffman) Decompress(input []byte, output []byte) (int, error) {
	inputSize := len(input)
	outputSize := len(output)

	// setup buffer pointers
	// converted to indices for go
	iDst := 0
	iSrc := 0
	iDstEnd := outputSize
	iSrcEnd := inputSize

	var bits uint = 0
	var bitCount uint = 0

	eofNode := &self.Nodes[HUFFMAN_EOF_SYMBOL]
	var node *Node

	for {
		// {A} try to load a node now, this will reduce dependency at location {D}
		node = nil
		if bitCount >= uint(HUFFMAN_LUTBITS) {
			node = self.DecodeLut[bits&uint(HUFFMAN_LUTMASK)]
		}

		// {B} fill with new bits
		for bitCount < 24 && iSrc != iSrcEnd {
			bits |= uint(input[iSrc]) << bitCount
			iSrc++
			bitCount += 8
		}

		// {C} load symbol now if we didn't that earlier at location {A}
		if node == nil {
			node = self.DecodeLut[bits&uint(HUFFMAN_LUTMASK)]
		}

		if node == nil {
			return -1, fmt.Errorf("No node found")
		}

		// {D} check if we hit a symbol already
		if node.NumBits != 0 {
			// remove the bits for that symbol
			bits >>= node.NumBits
			bitCount -= node.NumBits
		} else {
			// remove the bits that the lut checked up for us
			bits >>= uint(HUFFMAN_LUTBITS)
			bitCount -= uint(HUFFMAN_LUTBITS)

			// walk the tree bit by bit
			for {
				// traverse tree
				node = &self.Nodes[node.Leafs[bits&1]]

				// remove bit
				bitCount--
				bits >>= 1

				// check if we hit a symbol
				if node.NumBits != 0 {
					break
				}

				// no more bits, decoding error
				if bitCount == 0 {
					return -1, fmt.Errorf("unexpected end of input")
				}
			}
		}

		// check for eof
		if node == eofNode {
			break
		}

		// output character
		if iDst == iDstEnd {
			return -1, fmt.Errorf("unexpected end of output")
		}
		output[iDst] = node.Symbol
		iDst++
	}

	// return the size of the decompressed buffer
	return iDst, nil
}
