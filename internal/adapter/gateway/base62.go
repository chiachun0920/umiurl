package gateway

import (
	"crypto/rand"
	"math/big"
)

const Base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type Base62CodeGenerator struct {
	length   int
	alphabet string
}

func NewBase62CodeGenerator(length int) *Base62CodeGenerator {
	if length <= 0 {
		length = 7
	}
	return &Base62CodeGenerator{
		length:   length,
		alphabet: Base62Alphabet,
	}
}

func (g *Base62CodeGenerator) Generate() (string, error) {
	out := make([]byte, g.length)
	max := big.NewInt(int64(len(g.alphabet)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = g.alphabet[n.Int64()]
	}
	return string(out), nil
}
