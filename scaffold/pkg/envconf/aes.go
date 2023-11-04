package envconf

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

type AESMixerOption func(opt *AESCryptor) error

func WithAESMethod(m string) AESMixerOption {
	return func(opt *AESCryptor) error {
		if !(m == "ofb" || m == "cfb") {
			return fmt.Errorf("invalid aes method %s", m)
		}
		opt.method = m
		return nil
	}
}

func NewAESCryptorFromKey(v string) (*AESCryptor, error) {
	parts := strings.Split(v, " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid key %s", v)
	}

	method := strings.Replace(parts[0], "aes/", "", -1)
	key, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	return NewAESCryptor(key, WithAESMethod(method))
}

func NewAESCryptor(key []byte, opts ...AESMixerOption) (*AESCryptor, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	m := &AESCryptor{
		block: block,
	}

	switch len(key) {
	case 16:
		m.size = 128
	case 24:
		m.size = 192
	case 32:
		m.size = 256
	}

	defaults := []AESMixerOption{
		WithAESMethod("cfb"),
	}

	for _, optFn := range append(defaults, opts...) {
		if err := optFn(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

type AESCryptor struct {
	block  cipher.Block
	method string
	size   int
}

func (m *AESCryptor) EncryptFor(writer io.Writer) (io.WriteCloser, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	var iv [aes.BlockSize]byte
	var stream cipher.Stream

	switch m.method {
	case "cfb":
		stream = cipher.NewCFBEncrypter(m.block, iv[:])
	case "ofb":
		stream = cipher.NewOFB(m.block, iv[:])
	default:
		return nil, fmt.Errorf("invalid aes method %s", m.method)
	}

	return &cipher.StreamWriter{S: stream, W: writer}, nil
}

func (m *AESCryptor) DecryptFor(reader io.Reader) (io.Reader, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	var iv [aes.BlockSize]byte
	var stream cipher.Stream

	switch m.method {
	case "cfb":
		stream = cipher.NewCFBDecrypter(m.block, iv[:])
	case "ofb":
		stream = cipher.NewOFB(m.block, iv[:])
	default:
		return nil, fmt.Errorf("invalid aes method %s", m.method)
	}

	return &cipher.StreamReader{S: stream, R: reader}, nil
}
