package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	//copyPattern()
	//copyNPattern()
	//readAllPattern()
	multiReaderPattern()
}

func multiReaderPattern() {
	r1 := strings.NewReader("READER 1")
	r2 := strings.NewReader("READER 2")
	r3 := strings.NewReader("READER 3")

	multiReader := io.MultiReader(r1, r2, r3)

	file, _ := os.Create("multi_reader.txt")

	// 표준출력 , 파일 에 동시에 쓰기
	// ⭐ 로깅할 때 특히 유용해. 콘솔과 파일에 동시에 로그를 남길 수 있
	multiWriter := io.MultiWriter(os.Stdout, file)

	//io.Copy(file, multiReader)
	io.Copy(multiWriter, multiReader)
	fmt.Println("완료!")
}

func copyPattern() {
	// 소스 Reader
	reader := strings.NewReader("이 데이터를 복사할 거예요!")

	// 목적지 Writer (파일)
	dest, err := os.Create("output.txt")
	if err != nil {
		fmt.Println("파일 생성 실패:", err)
		return
	}
	defer dest.Close()

	// 복사 실행
	// ⭐ io.Copy는 내부적으로 32KB 버퍼를 사용해서 효율적으로 데이터를 전송
	written, err := io.Copy(dest, reader)
	if err != nil {
		fmt.Println("복사 실패:", err)
		return
	}
	fmt.Printf("총 %d 바이트를 복사했어요!\n", written)
}

func copyNPattern() {
	// 소스 Reader
	reader := strings.NewReader("이 데이터를 복사할 거예요!")

	// 목적지 Writer (파일)
	dest, err := os.Create("output.txt")
	if err != nil {
		fmt.Println("파일 생성 실패:", err)
		return
	}
	defer dest.Close()

	// 복사 실행
	// ⭐ io.CopyN은 정확히 n 바이트만 복사
	written, err := io.CopyN(dest, reader, 20)
	if err != nil {
		fmt.Println("복사 실패:", err)
		return
	}
	fmt.Printf("총 %d 바이트를 복사했어요!\n", written)
}

func readAllPattern() {
	// 소스 Reader
	reader := strings.NewReader("이 데이터를 복사할 거예요!")

	// ⭐ io.ReadAll은 Reader에서 EOF까지 모두 읽어서 바이트 슬라이스로 반환
	// 💥 ❗❗ io.ReadAll은 모든 데이터를 메모리에 올려. 대용량 파일에는 사용하지 마! 메모리가 터질 수 있어.
	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("읽기 실패:", err)
		return
	}
	fmt.Println(string(data))
}
