package tools

import "crypto/sha256"
import "encoding/hex"

import "io/ioutil"
import "mime/multipart"

func Sha256(buf []byte) string {
	s256 := sha256.New()
	s256.Write(buf)

	return hex.EncodeToString(s256.Sum(nil))
}

func Sha256FileData(h *multipart.FileHeader) (string, error) {
	f, _ := h.Open()
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	sha256Str := Sha256(buf)
	return sha256Str, nil
}
