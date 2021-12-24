package network

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"net"
	"strconv"
)

const (
	// Size is the length of the Crypto-PAn keying material.
	Size = keySize + blockSize

	blockSize = aes.BlockSize
	keySize   = 128 / 8
)

// KeySizeError is the error returned when the provided key is an invalid
// length.
type KeySizeError int

func (e KeySizeError) Error() string {
	return "invalid key size " + strconv.Itoa(int(e))
}

type bitvector [blockSize]byte

func (v *bitvector) SetBit(idx, bit uint) {
	byteIdx := idx / 8
	bitIdx := 7 - idx&7
	oldBit := uint8((v[byteIdx] & (1 << bitIdx)) >> bitIdx)
	flip := 1 ^ subtle.ConstantTimeByteEq(oldBit, uint8(bit))
	v[byteIdx] ^= byte(flip << bitIdx)
}

func (v *bitvector) Bit(idx uint) uint {
	byteIdx := idx / 8
	bitIdx := 7 - idx&7
	return uint((v[byteIdx] & (1 << bitIdx)) >> bitIdx)
}

// Cryptopan is an instance of the Crypto-PAn algorithm, initialized with a
// given key.
type Cryptopan struct {
	aesImpl cipher.Block
	pad     bitvector
}

// Anonymize anonymizes the provided IP address with the Crypto-PAn algorithm.
func (ctx *Cryptopan) Anonymize(addr net.IP) net.IP {
	var obfsAddr []byte
	if v4addr := addr.To4(); v4addr != nil {
		obfsAddr = ctx.anonymize(v4addr)
		return net.IPv4(obfsAddr[0], obfsAddr[1], obfsAddr[2], obfsAddr[3])
	} else if v6addr := addr.To16(); v6addr != nil {
		// None of the other implementations in the wild do something like
		// this, but there's no reason I can think of beyond "it'll be really
		// slow" as to why it's not valid.
		obfsAddr = ctx.anonymize(v6addr)
		addr := make(net.IP, net.IPv6len)
		copy(addr[:], obfsAddr[:])
		return addr
	}

	panic("unsupported address type")
}

func (ctx *Cryptopan) anonymize(addr net.IP) []byte {
	addrBits := uint(len(addr) * 8)
	var origAddr, input, output, toXor bitvector
	copy(origAddr[:], addr[:])
	copy(input[:], ctx.pad[:])

	// The first bit does not take any bits from orig_addr.
	ctx.aesImpl.Encrypt(output[:], input[:])
	toXor.SetBit(0, output.Bit(0))

	// The rest of the one time pad is build by copying orig_addr into the AES
	// input bit by bit (MSB first) and encrypting with ECB-AES128.
	for pos := uint(1); pos < addrBits; pos++ {
		// Copy an additional bit into input from orig_addr.
		input.SetBit(pos-1, origAddr.Bit(pos-1))

		// ECB-AES128 the input, only one bit of output is used per iteration.
		ctx.aesImpl.Encrypt(output[:], input[:])

		// Note: Per David Stott@Lucent, using the MSB of the PRF output leads
		// to weaker anonymized output.  Jinliang Fan (one of the original
		// Crypto-PAn authors) claims that a new version that incorporates one
		// of his suggested tweaks is forthcoming, but it looks like that never
		// happened, and no one else does that.
		//
		// Something like: toXor.SetBit(pos, output.Bit(pos)) will fix this,
		// but will lead to different output than every other implementation.
		toXor.SetBit(pos, output.Bit(0))
	}

	// Xor the pseudorandom one-time-pad with the address and return.
	for i := 0; i < len(addr); i++ {
		toXor[i] ^= origAddr[i]
	}
	return toXor[:len(addr)]
}

// NewCryptoPAn constructs and initializes Crypto-PAn with a given key.
func NewCryptoPAn(key []byte) (ctx *Cryptopan, err error) {
	if len(key) != Size {
		return nil, KeySizeError(len(key))
	}

	ctx = new(Cryptopan)
	if ctx.aesImpl, err = aes.NewCipher(key[0:keySize]); err != nil {
		return nil, err
	}
	ctx.aesImpl.Encrypt(ctx.pad[:], key[keySize:])

	return
}
