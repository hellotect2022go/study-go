package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	//readPattern()
	writePattern()
}

func readPattern() {

	// ⭐ strings.Reader 는 io.Reader 인터페이스를 구현
	reader := strings.NewReader("안녕하세요, Go 스트리밍의 세계에 오신 걸 환영합니다!")

	// ⭐ 256바이트씩 읽기 []byte 256 사이즈의 슬라이스를 만듬
	buf := make([]byte, 256)

	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error:", err)
			break
		}

		fmt.Println(buf[:n])
		fmt.Printf("\n읽은 바이트 수: %d, 내용: %s\n", n, string(buf[:n]))
	}
}

func writePattern() {
	data := []byte("안녕 다같이 go 의 스트리밍에 대해 공부해보자!!")

	// ⭐ 표준 출력도 io.Writer 인터페이스를 구현
	writer := os.Stdout
	n, err := writer.Write(data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("\n쓴 바이트 수: %d\n", n)
}
