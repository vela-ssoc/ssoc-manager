package service

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"golang.org/x/crypto/blake2b"
)

type Hash struct {
}

func NewHash() *Hash {
	return &Hash{}
}

func (h *Hash) Sum(b []byte) model.Checksum {
	hs := h.NewHasher()
	_, _ = hs.Write(b)
	return hs.Sum()
}

func (h *Hash) Read(r io.Reader) model.Checksum {
	hs := h.NewHasher()
	_, _ = io.Copy(hs, r)
	return hs.Sum()
}

func (*Hash) NewHasher() *HashWriter {
	m5 := md5.New()
	s1 := sha1.New()
	s2 := sha256.New()
	b2, _ := blake2b.New256(nil)

	return &HashWriter{
		md5:    m5,
		sha1:   s1,
		sha256: s2,
		blake2: b2,
		multi:  io.MultiWriter(m5, s1, s2, b2),
	}
}

type HashWriter struct {
	md5    hash.Hash
	sha1   hash.Hash
	sha256 hash.Hash
	blake2 hash.Hash
	multi  io.Writer
	count  int
}

func (hw *HashWriter) Write(p []byte) (int, error) {
	n, err := hw.multi.Write(p)
	if n > 0 {
		hw.count += n
	}
	return n, err
}

func (hw *HashWriter) Count() int {
	return hw.count
}

func (hw *HashWriter) Sum() model.Checksum {
	return model.Checksum{
		MD5:     hex.EncodeToString(hw.md5.Sum(nil)),
		SHA1:    hex.EncodeToString(hw.sha1.Sum(nil)),
		SHA256:  hex.EncodeToString(hw.sha256.Sum(nil)),
		BLAKE2b: hex.EncodeToString(hw.blake2.Sum(nil)),
	}
}
