package main

func reverse(str string) string {
    s := []rune(str)
    length := len(str)
    for i := 0;i < length / 2; i++ {
        s[i], s[length - 1 - i] = s[length -1 - i], s[i]
    }
    return string(s)
}

