package utils

import (
    "fmt"
    "strings"
    "unicode/utf16"
    // "unicode/utf8"
)


func UnescapeUnicode(source string) string {
    result := source

    for {
        index := strings.Index(source, `\u`)

        if index == -1 {
            break
        }

        var r1, r2 rune

        if len(source) < index + 6 {
            break
        }

        _, err := fmt.Sscanf(source[index + 2:index + 6], "%x", &r1)

        next := 0

        if utf16.IsSurrogate(r1) {
            if len(source) >= index + 12 {
                fmt.Sscanf(source[index + 8:index + 12], "%x", &r2)
                if utf16.IsSurrogate(r2) {
                    repl := "\\u" + source[index + 2:index + 6] + "\\u" + source[index + 8:index + 12]
                    result = strings.ReplaceAll(result, repl, string(utf16.DecodeRune(r1, r2)))
                } else {
                    result = strings.ReplaceAll(result, "\\u" + source[index + 8:index + 12], string(r2))
                }
                next = index + 12
            } else {
                next = index + 2
            }
        } else {
            if err == nil {
                result = strings.ReplaceAll(result, "\\u" + source[index + 2:index + 6], string(r1))
                next = index + 6
            } else {
                next = index + 2
            }
        }

        source = source[next:]
    }

    return result
}
