package main

import (
    "os"
    "log"
)

func main() {
    f, err := os.OpenFile("/home/qyu/gvisor-test/test.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("打开文件失败: %v", err)
    }
    defer f.Close()
    if _, err := f.WriteString("追加的一行内容\n"); err != nil {
        log.Fatalf("写入失败: %v", err)
    }
}