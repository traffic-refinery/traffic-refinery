package flowstats

import (
	"github.com/google/gopacket"
)

const fnvBasis = 14695981039346656037
const fnvPrime = 1099511628211

// TupleFlow is a special flow that includes the 4-tuple consisting of the network and transport flows
// src and dst are IP addresses, src2 and dst2 are port numbers
type TupleFlow struct {
	typ          gopacket.EndpointType
	slen, dlen   int
	s2len, d2len int
	src, dst     [gopacket.MaxEndpointSize]byte
	src2, dst2   [gopacket.MaxEndpointSize]byte
}

// NewTupleFlow generates a TupleFlow given the 4-tuple
func NewTupleFlow(src, dst, src2, dst2 []byte) (f TupleFlow) {
	f.slen = len(src)
	f.dlen = len(dst)
	f.s2len = len(src2)
	f.d2len = len(dst2)
	if f.slen > gopacket.MaxEndpointSize || f.dlen > gopacket.MaxEndpointSize {
		panic("flow raw byte length greater than MaxEndpointSize")
	}
	if f.s2len > gopacket.MaxEndpointSize || f.d2len > gopacket.MaxEndpointSize {
		panic("flow raw byte length greater than MaxEndpointSize")
	}
	copy(f.src[:], src)
	copy(f.dst[:], dst)
	copy(f.src2[:], src2)
	copy(f.dst2[:], dst2)
	return
}

// FastHash is a special version of gopacket's fasthash as we incorporate a 4-tuple rather than only 2 endpoints
// The hash must be symmetric srcIP,srcPort->dstIP,dstPort must collide with dstIP,dstPort->srcIP,srcPort
func (f TupleFlow) FastHash() (h uint64) {
	// This combination must be commutative.  We don't use ^, since that would
	// give the same hash for all A->A flows.
	h = fnvHash(f.src[:f.slen]) + fnvHash(f.dst[:f.dlen]) + fnvHash(f.src2[:f.s2len]) + fnvHash(f.dst2[:f.d2len])
	h ^= uint64(f.typ)
	h *= fnvPrime
	return
}

// fnvHash is used by FastHash, and implements the FNV hash
// created by Glenn Fowler, Landon Curt Noll, and Phong Vo.
// See http://isthe.com/chongo/tech/comp/fnv/.
func fnvHash(s []byte) (h uint64) {
	h = fnvBasis
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= fnvPrime
	}
	return
}
