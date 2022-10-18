package libs

import (
    "time"
    _cryptoRand "crypto/rand"
    _mathRand "math/rand"
)

func init() {
    _mathRand.Seed(time.Now().UnixNano())
}

var (
    _cryptoRandFunc = _cryptoRand.Read
    _mathRandFunc = _mathRand.Read
)

func _randomBytesCrypto(length int) ([]byte, error) {
    b := make([]byte, length)
    if _, err := _cryptoRandFunc(b); err != nil {
        return nil, err
    }
    return b, nil
}

func _randomBytesMath(length int) []byte {
    b := make([]byte, length)
    _mathRandFunc(b)
    return b
}

func RandomBytes(length int) []byte {
    if b, err := _randomBytesCrypto(length); err != nil {
        return _randomBytesMath(length)
    } else {
        return b
    }
}

var (
    _randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    _randomHexSource = "0123456789abcdef"
    _randomNumberSource = "0123456789"
)

func _doRandomString(length int, _type string) string {
    var (
        _s string
        _l int
    )

    result := ""

    b := RandomBytes(length)

    if _type == "string" {
        _s = _randomStringSource
        _l = len(_randomStringSource)
    } else if _type == "hex" {
        _s = _randomHexSource
        _l = len(_randomHexSource)
    } else if _type == "numbers" {
        _s = _randomNumberSource
        _l = len(_randomNumberSource)
    }

    for i := 0; i < length; i++ {
        c := string(_s[int(b[i]) % _l])
        result += c
    }

    return result
}


func RandomString(length int) string {
    return _doRandomString(length, "string")
}

func RandomStringHex(length int) string {
    return _doRandomString(length, "hex")
}

func RandomStringNumbers(length int) string {
    return _doRandomString(length, "numbers")
}
