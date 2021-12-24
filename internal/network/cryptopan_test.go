/*
 * Copyright (c) 2014, Yawning Angel <yawning at schwanenlied dot me>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *  * Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 *
 *  * Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package network

import (
	"net"
	"testing"
)

// testKey is the key used in the original Crypto-PAn source distribution
// sample.
var testKey = []byte{21, 34, 23, 141, 51, 164, 207, 128, 19, 10, 91, 22, 73, 144, 125, 16, 216, 152, 143, 131, 121, 121, 101, 39, 98, 87, 76, 45, 42, 132, 34, 2}

type testVector struct {
	origAddr string
	obfsAddr string
}

// TestCryptopanIPv4 tests against the IP addresses/sanitized output from the
// "sample_trace_raw.dat"/"sample_trace_sanitized.dat" files included in the
// Crypto-PAn source distribution.
func TestCryptopanIPv4(t *testing.T) {
	cpan, err := NewCryptoPAn(testKey)
	if err != nil {
		t.Fatal("New(testKey) failed:", err)
	}

	// The vectors were pulled out by abusing awk, paste, sort and uniq.
	// Thus ordering is different from the sample data.
	v4Vectors := []testVector{
		{"128.11.68.132", "135.242.180.132"},
		{"129.118.74.4", "134.136.186.123"},
		{"130.132.252.244", "133.68.164.234"},
		{"141.223.7.43", "141.167.8.160"},
		{"141.233.145.108", "141.129.237.235"},
		{"152.163.225.39", "151.140.114.167"},
		{"156.29.3.236", "147.225.12.42"},
		{"165.247.96.84", "162.9.99.234"},
		{"166.107.77.190", "160.132.178.185"},
		{"192.102.249.13", "252.138.62.131"},
		{"192.215.32.125", "252.43.47.189"},
		{"192.233.80.103", "252.25.108.8"},
		{"192.41.57.43", "252.222.221.184"},
		{"193.150.244.223", "253.169.52.216"},
		{"195.205.63.100", "255.186.223.5"},
		{"198.200.171.101", "249.199.68.213"},
		{"198.26.132.101", "249.36.123.202"},
		{"198.36.213.5", "249.7.21.132"},
		{"198.51.77.238", "249.18.186.254"},
		{"199.217.79.101", "248.38.184.213"},
		{"202.49.198.20", "245.206.7.234"},
		{"203.12.160.252", "244.248.163.4"},
		{"204.184.162.189", "243.192.77.90"},
		{"204.202.136.230", "243.178.4.198"},
		{"204.29.20.4", "243.33.20.123"},
		{"205.178.38.67", "242.108.198.51"},
		{"205.188.147.153", "242.96.16.101"},
		{"205.188.248.25", "242.96.88.27"},
		{"205.245.121.43", "242.21.121.163"},
		{"207.105.49.5", "241.118.205.138"},
		{"207.135.65.238", "241.202.129.222"},
		{"207.155.9.214", "241.220.250.22"},
		{"207.188.7.45", "241.255.249.220"},
		{"207.25.71.27", "241.33.119.156"},
		{"207.33.151.131", "241.1.233.131"},
		{"208.147.89.59", "227.237.98.191"},
		{"208.234.120.210", "227.154.67.17"},
		{"208.28.185.184", "227.39.94.90"},
		{"208.52.56.122", "227.8.63.165"},
		{"209.12.231.7", "226.243.167.8"},
		{"209.238.72.3", "226.6.119.243"},
		{"209.246.74.109", "226.22.124.76"},
		{"209.68.60.238", "226.184.220.233"},
		{"209.85.249.6", "226.170.70.6"},
		{"212.120.124.31", "228.135.163.231"},
		{"212.146.8.236", "228.19.4.234"},
		{"212.186.227.154", "228.59.98.98"},
		{"212.204.172.118", "228.71.195.169"},
		{"212.206.130.201", "228.69.242.193"},
		{"216.148.237.145", "235.84.194.111"},
		{"216.157.30.252", "235.89.31.26"},
		{"216.184.159.48", "235.96.225.78"},
		{"216.227.10.221", "235.28.253.36"},
		{"216.254.18.172", "235.7.16.162"},
		{"216.32.132.250", "235.192.139.38"},
		{"216.35.217.178", "235.195.157.81"},
		{"24.0.250.221", "100.15.198.226"},
		{"24.13.62.231", "100.2.192.247"},
		{"24.14.213.138", "100.1.42.141"},
		{"24.5.0.80", "100.9.15.210"},
		{"24.7.198.88", "100.10.6.25"},
		{"24.94.26.44", "100.88.228.35"},
		{"38.15.67.68", "64.3.66.187"},
		{"4.3.88.225", "124.60.155.63"},
		{"63.14.55.111", "95.9.215.7"},
		{"63.195.241.44", "95.179.238.44"},
		{"63.97.7.140", "95.97.9.123"},
		{"64.14.118.196", "0.255.183.58"},
		{"64.34.154.117", "0.221.154.117"},
		{"64.39.15.238", "0.219.7.41"},
		{"10.185.25.118", "117.161.29.118"},
	}

	for _, vec := range v4Vectors {
		origAddr := net.ParseIP(vec.origAddr)
		obfsAddr := net.ParseIP(vec.obfsAddr)
		testAddr := cpan.Anonymize(origAddr)
		if !obfsAddr.Equal(testAddr) {
			t.Errorf("%s -> %s != %s", origAddr, testAddr, obfsAddr)
		}
	}
}

// TestCryptopanIPv6 tests against a few random IPv6 addresses, just to make
// sure the code does not panic.
func TestCryptopanIPv6(t *testing.T) {
	cpan, err := NewCryptoPAn(testKey)
	if err != nil {
		t.Fatal("NewCryptoPAn(testKey) failed:", err)
	}

	v6Vectors := []testVector{
		{"::1", "78ff:f001:9fc0:20df:8380:b1f1:704:ed"},
		{"::2", "78ff:f001:9fc0:20df:8380:b1f1:704:ef"},
		{"::ffff", "78ff:f001:9fc0:20df:8380:b1f1:704:f838"},
		{"2001:db8::1", "4401:2bc:603f:d91d:27f:ff8e:e6f1:dc1e"},
		{"2001:db8::2", "4401:2bc:603f:d91d:27f:ff8e:e6f1:dc1c"},
	}

	for _, vec := range v6Vectors {
		origAddr := net.ParseIP(vec.origAddr)
		obfsAddr := net.ParseIP(vec.obfsAddr)
		testAddr := cpan.Anonymize(origAddr)
		if !obfsAddr.Equal(testAddr) {
			t.Errorf("%s -> %s != %s", origAddr, testAddr, obfsAddr)
		}
	}
}

// BenchmarkCryptopanIPv4 benchmarks annonymizing IPv4 addresses.
func BenchmarkCryptopanIPv4(b *testing.B) {
	cpan, err := NewCryptoPAn(testKey)
	if err != nil {
		b.Fatal("NewCryptoPAn(testKey) failed:", err)
	}
	b.ResetTimer()

	addr := net.ParseIP("192.168.1.1")
	for i := 0; i < b.N; i++ {
		_ = cpan.Anonymize(addr)
	}
}

// BenchmarkCryptopanIPv6 benchmarks annonymizing IPv6 addresses.
func BenchmarkCryptopanIPv6(b *testing.B) {
	cpan, err := NewCryptoPAn(testKey)
	if err != nil {
		b.Fatal("NewCryptoPAn(testKey) failed:", err)
	}
	b.ResetTimer()

	addr := net.ParseIP("2001:db8::1")
	for i := 0; i < b.N; i++ {
		_ = cpan.Anonymize(addr)
	}
}
