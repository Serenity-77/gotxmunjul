package utils

import (
    "testing"
    "github.com/stretchr/testify/assert"
)


func TestUnescapeUnicode(t *testing.T) {
    target := []struct {
        expected string
        actual string
    }{
        {"Hello World", "Hello World"},
        {"Horas", "Horas"},
        {"Hello bro 😀🙂😀", "Hello bro \\ud83d\\ude00\\ud83d\\ude42\\ud83d\\ude00"},
        {"Hello bro 😀", "Hello bro \\ud83d\\ude00"},
        {"Hello bro 😀🙂😀 hehehe", "Hello bro \\ud83d\\ude00\\ud83d\\ude42\\ud83d\\ude00 hehehe"},
        {"Hello bro 😀 Horas", "Hello bro \\ud83d\\ude00 Horas"},
        {"😀🙂😀", "\\ud83d\\ude00\\ud83d\\ude42\\ud83d\\ude00"},
        {"😀", "\\ud83d\\ude00"},
        {"😀😀", "\\ud83d\\ude00😀"},
        {"Hello Κ", "Hello \\u039a"},
        {"Κατά τον Δαίμονον Εαυτού", "\\u039a\\u03b1\\u03c4\\u03ac \\u03c4\\u03bf\\u03bd \\u0394\\u03b1\\u03af\\u03bc\\u03bf\\u03bd\\u03bf\\u03bd \\u0395\\u03b1\\u03c5\\u03c4\\u03bf\\u03cd"},
        {"Hello Κατά τον My Δαίμονον Εαυτού Brooo", "Hello \\u039a\\u03b1\\u03c4\\u03ac \\u03c4\\u03bf\\u03bd My \\u0394\\u03b1\\u03af\\u03bc\\u03bf\\u03bd\\u03bf\\u03bd \\u0395\\u03b1\\u03c5\\u03c4\\u03bf\\u03cd Brooo"},
        {"Κατά 🙂 τον Δαίμον😀ον Εαυτού", "\\u039a\\u03b1\\u03c4\\u03ac \\ud83d\\ude42 \\u03c4\\u03bf\\u03bd \\u0394\\u03b1\\u03af\\u03bc\\u03bf\\u03bd\\ud83d\\ude00\\u03bf\\u03bd \\u0395\\u03b1\\u03c5\\u03c4\\u03bf\\u03cd"},
        {"Hai ❌", "Hai \\u274c"},
        {"Hai ❌😀", "Hai \\u274c\\ud83d\\ude00"},
        {"Hai ❌ 😀", "Hai \\u274c \\ud83d\\ude00"},
        {"Hai ❌😀🙂😀", "Hai \\u274c\\ud83d\\ude00\\ud83d\\ude42\\ud83d\\ude00"},
        {"Hai ❌😀🙂😀 hohoho", "Hai \\u274c\\ud83d\\ude00\\ud83d\\ude42\\ud83d\\ude00 hohoho"},
        {"Hai ❌ 😀🙂😀", "Hai \\u274c \\ud83d\\ude00\\ud83d\\ude42\\ud83d\\ude00"},
        {"Hai ❌ 😀🙂😀 hihihih", "Hai \\u274c \\ud83d\\ude00\\ud83d\\ude42\\ud83d\\ude00 hihihih"},
        {"1234", "1234"},
        {"hahaha 1234", "hahaha 1234"},
        {"hahaha 1234 hehehe", "hahaha 1234 hehehe"},
        {"Yang udah gila\\unya???", "Yang udah gila\\unya???"},
        {"Yang udah gila\\unya??\\unya?\\unya", "Yang udah gila\\unya??\\unya?\\unya"},
        {"Yang udah gila\\u❌yy???", "Yang udah gila\\u\\u274cyy???"},
        {"Yang udah gila\\u😀yy???", "Yang udah gila\\u\\ud83d\\ude00yy???"},
        {"Yang udah gila\\u❌\\uy", "Yang udah gila\\u\\u274c\\uy"},
        {"Yang udah gila\\u😀\\uy", "Yang udah gila\\u\\ud83d\\ude00\\uy"},
        {"Yang udah gila\\u😀\\uy hahaha", "Yang udah gila\\u\\ud83d\\ude00\\uy hahaha"},
        {"Hello \\ud83d", "Hello \\ud83d"},
        {"Hello \\ud83dy\\uy", "Hello \\ud83dy\\uy"},
        {"Hello 😀🙂", "Hello \\ud83d\\ude00\\ud83d\\ude42"},
        {"Hello \\ud83d❌", "Hello \\ud83d\\u274c"},
        {"Hello \\ud83d❌ semua", "Hello \\ud83d\\u274c semua"},
        {"Hello \\ud83d❌😀", "Hello \\ud83d\\u274c\\ud83d\\ude00"},
    }

    for _, source := range target {
        assert.Equal(t, source.expected, UnescapeUnicode(source.actual))
    }
}
