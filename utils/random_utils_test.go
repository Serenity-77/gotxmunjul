package utils

import (
    "io"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestRandomBytesCrypto(t *testing.T) {
    length := []int{8, 16, 32, 64}
    for _, l := range length {
        b, _:= _randomBytesCrypto(l)
        assert.NotNil(t, b)
        assert.Len(t, b, l)
    }
}


func TestRandomBytesMath(t *testing.T) {
    length := []int{8, 16, 32, 64}
    for _, l := range length {
        b := _randomBytesMath(l)
        assert.NotNil(t, b)
        assert.Len(t, b, l)
    }
}


func TestRandomBytes(t *testing.T) {
    length := []int{8, 16, 32, 64}
    for _, l := range length {
        b := RandomBytes(l)
        assert.NotNil(t, b)
        assert.Len(t, b, l)
    }
}

func TestRandomBytesUseMathRand(t *testing.T) {
    var oldCryptoRand = _cryptoRandFunc
    var oldMathRand = _mathRandFunc

    defer func() {
        _cryptoRandFunc = oldCryptoRand
        _mathRandFunc = oldMathRand
    }()

    _cryptoRandFunc = func(b []byte) (int, error) {
        return 0, io.ErrUnexpectedEOF
    }

    called := false

    _mathRandFunc = func(b []byte) (int, error) {
        called = true
        return oldMathRand(b)
    }

    length := []int{8, 16, 32, 64}
    for _, l := range length {
        b := RandomBytes(l)
        assert.NotNil(t, b)
        assert.Len(t, b, l)
    }

    assert.True(t, called)
}

func TestRandomString(t *testing.T) {
    length := []int{8, 16, 32, 64}
    for _, l := range length {
        b := RandomString(l)
        assert.NotNil(t, b)
        assert.Len(t, b, l)
    }
}
