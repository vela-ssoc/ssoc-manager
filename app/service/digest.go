package service

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"hash"

	"golang.org/x/crypto/bcrypt"
)

type DigestService interface {
	Hashed(passwd string) (string, error)
	Compare(hashed, plaintext string) bool
	SHA1() hash.Hash
	SumMD5([]byte) string
}

func Digest() DigestService {
	return &digestService{}
}

type digestService struct{}

func (dig *digestService) Hashed(passwd string) (string, error) {
	pwd, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	return string(pwd), err
}

func (dig *digestService) Compare(hashed, plaintext string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext))
	return err == nil
}

func (dig *digestService) SHA1() hash.Hash {
	return sha1.New()
}

func (dig *digestService) SumMD5(dat []byte) string {
	sum := md5.Sum(dat)
	return hex.EncodeToString(sum[:])
}
