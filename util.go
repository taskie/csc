package csc

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func ToHexString(bs []byte) string {
	return fmt.Sprintf("%x", bs)
}

func CalcSha256(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := bufio.NewReader(file)
	digest := sha256.New()
	_, err = io.Copy(digest, r)
	if err != nil {
		return nil, err
	}
	return digest.Sum(nil), nil
}

func CalcSha256HexString(path string) (string, error) {
	bs, err := CalcSha256(path)
	if err != nil {
		return "", err
	}
	return ToHexString(bs), nil
}
