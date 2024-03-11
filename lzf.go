/*
 * Public Domain (-) 2010-2011 The zenor Authors.
 * Copyright (c) 2024 ZengLy <zenor.zly@gmail.com>
 * All rights reserved.
 *
 * A rewrite of the C# version of Chase Pettit lzf with Go:
 * https://github.com/Chaser324/LZF/blob/master/CLZF2.cs
 * Copyright (c) 2010 Chase Pettit <chasepettit@gmail.com>
 */

package lzf

import "sync"

const (
	BUFFER_SIZE_ESTIMATE        = 2
	HLOG                 uint32 = 14
	HSIZE                uint32 = 1 << HLOG
	MAX_LIT              uint32 = 1 << 5
	MAX_OFF              uint32 = 1 << 13
	MAX_REF              uint32 = (1 << 8) + (1 << 3)
)

var HashTable []int64 = make([]int64, HSIZE)
var locker = sync.Mutex{}

// 压缩
func Compress(input []byte) (output []byte) {

	return compressTemp(input, len(input))
}

func compressTemp(inputBytes []byte, inputLength int) []byte {
	var tempBuffer []byte
	byteCount := compressBuff(inputBytes, &tempBuffer, inputLength)
	outputBytes := make([]byte, byteCount)
	_ = copy(outputBytes, tempBuffer)
	return outputBytes[:byteCount]
}

func compressBuff(inputBytes []byte, outputBuff *[]byte, inputLength int) int {
	outputByteCountGuess := inputLength * BUFFER_SIZE_ESTIMATE
	if outputBuff == nil || len(*outputBuff) < outputByteCountGuess {
		*outputBuff = make([]byte, outputByteCountGuess)
	}
	byteCount := lzf_compress(inputBytes, outputBuff, inputLength)
	for byteCount == 0 {
		outputByteCountGuess *= 2
		*outputBuff = make([]byte, outputByteCountGuess)
		byteCount = lzf_compress(inputBytes, outputBuff, inputLength)
	}
	return byteCount
}

// 解压缩
func Decompress(inputBytes []byte) (output []byte) {
	return decompressTemp(inputBytes, len(inputBytes))
}

func decompressTemp(inputBytes []byte, inputLength int) (output []byte) {
	tempBuffer := []byte{}
	byteCount := decompressBuff(inputBytes, &tempBuffer, inputLength)

	outputBytes := make([]byte, byteCount)
	_ = copy(outputBytes, tempBuffer)
	return outputBytes
}

func decompressBuff(inputBytes []byte, outputBuff *[]byte, inputLength int) int {
	outputByteCountGuess := inputLength * BUFFER_SIZE_ESTIMATE
	if outputBuff == nil || len(*outputBuff) < outputByteCountGuess {
		*outputBuff = make([]byte, outputByteCountGuess)
	}
	byteCount := lzf_decompress(inputBytes, outputBuff, inputLength)
	for byteCount == 0 {
		outputByteCountGuess *= 2
		*outputBuff = make([]byte, outputByteCountGuess)
		byteCount = lzf_decompress(inputBytes, outputBuff, inputLength)
	}
	return byteCount
}

func lzf_compress(input []byte, output *[]byte, inputLength int) int {
	outputLength := len(*output)
	var hslot int64
	iidx := uint32(0)
	oidx := uint32(0)
	reference := int64(0)
	hval := uint32(((input[iidx]) << 8) | input[iidx+1]) // FRST(in_data, iidx);
	off := int64(0)
	lit := 0
	locker.Lock() // lock
	HashTable = make([]int64, HSIZE)
	for {
		if int(iidx) < inputLength-2 {
			hval = (hval << 8) | (uint32)(input[iidx+2])
			hslot = int64((hval ^ (hval << 5)) >> int64((3*8-HLOG)-hval*5) & (HSIZE - 1))
			reference = HashTable[hslot]
			HashTable[hslot] = int64(iidx)
			if off = int64(iidx) - reference - 1; off < int64(MAX_OFF) && int(iidx)+4 < inputLength && reference > 0 && input[reference+0] == input[iidx+0] && input[reference+1] == input[iidx+1] && input[reference+2] == input[iidx+2] {
				len := uint32(2)
				maxlen := uint32(inputLength) - iidx - len
				if maxlen > MAX_REF {
					maxlen = MAX_REF
				}
				if (int(oidx) + lit + 1 + 3) >= outputLength {
					return 0
				}
				for len < maxlen && input[reference+int64(len)] == input[iidx+len] {
					len++
				}
				if lit != 0 {
					(*output)[oidx] = byte(lit - 1)
					oidx++
					lit = -lit
					for lit != 0 {
						(*output)[oidx] = input[int(iidx)+lit]
						oidx++
						lit++
					}
				}
				len -= 2
				iidx++

				if len < 7 {
					(*output)[oidx] = (byte)((off >> 8) + (int64(len) << 5))
					oidx++
				} else {
					(*output)[oidx] = (byte)((off >> 8) + (7 << 5))
					oidx++
					(*output)[oidx] = (byte)(len - 7)
					oidx++
				}
				(*output)[oidx] = byte(off)
				oidx++

				iidx += len - 1
				hval = uint32(((input[iidx]) << 8) | input[iidx+1])

				iidx += 2
				hval = (hval << 8) | uint32(input[iidx+2])
				HashTable[((hval ^ (hval << 5)) >> (int)((3*8-HLOG)-hval*5) & (HSIZE - 1))] = int64(iidx)
				iidx++

				hval = (hval << 8) | uint32(input[iidx+2])
				HashTable[((hval ^ (hval << 5)) >> (int)((3*8-HLOG)-hval*5) & (HSIZE - 1))] = int64(iidx)
				iidx++
				continue
			} else if int(iidx) == inputLength {
				break
			}
			/* one more literal byte we must copy */
			lit++
			iidx++
			if lit == int(MAX_LIT) {
				if int(oidx+1+MAX_LIT) >= outputLength {
					return 0
				}
				(*output)[oidx] = (byte)(MAX_LIT - 1)
				oidx++
				lit = -lit
				for lit != 0 {
					(*output)[oidx] = input[int(iidx)+lit]
					oidx++
					lit++
				}
			}
		} // for
	} // lock
	locker.Unlock()

	if lit != 0 {
		if int(oidx)+lit+1 >= outputLength {
			return 0
		}
		oidx++
		(*output)[oidx] = byte(lit - 1)
		lit = -lit
		for lit != 0 {
			(*output)[oidx] = input[int(iidx)+lit]
			oidx++
			lit++
		}
	}
	return int(oidx)
}

func lzf_decompress(input []byte, output *[]byte, inputLength int) int {
	outputLength := len(*output)
	iidx := uint32(0)
	oidx := uint32(0)
	for iidx < uint32(inputLength) {

		ctrl := uint32(input[iidx])
		iidx++

		if ctrl < (1 << 5) { /* literal run */
			ctrl++
			if oidx+ctrl > uint32(outputLength) {
				//SET_ERRNO (E2BIG);
				return 0
			}
			for ctrl != 0 {
				(*output)[oidx] = input[iidx]
				oidx++
				iidx++
				ctrl--
			}
		} else { /* back reference */
			len := uint32(ctrl >> 5)
			reference := int(oidx - ((ctrl & 0x1f) << 8) - 1)
			if len == 7 {
				len += uint32(input[iidx])
				iidx++
			}
			reference -= int(input[iidx])
			iidx++

			if int(oidx+len+2) > outputLength {
				return 0
			}

			if reference < 0 {
				return 0
			}
			(*output)[oidx] = (*output)[reference]
			oidx++
			reference++
			(*output)[oidx] = (*output)[reference]
			oidx++
			reference++

			for len != 0 {
				(*output)[oidx] = (*output)[reference]
				oidx++
				reference++
				len--
			}
		}
	}
	return int(oidx)
}
