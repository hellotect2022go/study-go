package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	//openFilePattern()
	//createFilePattern()
	//bufferedFilePattern()
	chunkedFilePattern()
}

// 정말 큰 파일을 처리할 때는 청크(chunk) 단위로 나눠서 읽는 게 좋아:
func chunkedFilePattern() {
	chunkSize := 1024 * 1024 * 100 // 100MB
	file, _ := os.Open("fake.log")
	defer file.Close()

	buffer := make([]byte, chunkSize)

	totalBytes := 0
	chunkNumber := 1

	for {
		// chunkSize 만큼 읽기
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Printf("청크 %d 읽기 실패: %v\n", chunkNumber, err)
			break
		}

		if n == 0 {
			break
		}

		// 여기서 데이터 처리
		fmt.Printf("청크 %d: %d 바이트 처리\n", chunkNumber, n)
		//fmt.Println(string(buffer[:n]))
		outputFile, _ := os.Create(fmt.Sprintf("chunk_%d.txt", chunkNumber))
		outputFile.Write(buffer[:n])
		outputFile.Close()

		// 실제로는 여기서 데이터를 분석하거나 변환

		totalBytes += n
		chunkNumber++

	}
	fmt.Printf("총 %d 바이트 처리 완료!\n", totalBytes)
	return
}

func bufferedFilePattern() {
	file, _ := os.Open("README.md")
	defer file.Close()

	// bufio.Scanner 로 줄단위 읽기
	// ⭐ bufio.Scanner는 내부적으로 버퍼링을 해서 시스템 콜 횟수를 줄여줘. 대용량 파일을 읽을 때 성능이 크게 향상돼! 🚀
	scanner := bufio.NewScanner(file)
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("%d: %s\n", lineNumber, line)
		lineNumber++

		// if lineNumber > 10 {
		// 	break
		// }
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("스캔 중 에러: %v\n", err)
	}
}

func createFilePattern() {
	file1, _ := os.Create("create_file.txt")

	defer file1.Close()

	file1.WriteString("새로운 파일에 쓰는 첫 번째 줄\n")
	file1.Write([]byte("두 번째 줄은 바이트 슬라이스로\n"))

	// ⭐⭐ 2. 더 세밀한 제어
	// os.O_RDONLY - 읽기 전용
	// os.O_WRONLY - 쓰기 전용
	// os.O_RDWR - 읽기/쓰기
	// os.O_APPEND - 파일 끝에 추가
	// os.O_CREATE - 파일이 없으면 생성
	// os.O_TRUNC - 파일을 열 때 내용 비우기
	// 이 플래그들은 비트 OR 연산자(|)로 조합해서 사용

	file2, _ := os.OpenFile("create_file.txt",
		os.O_APPEND,
		//|os.O_CREATE|os.O_WRONLY, // 추가 모드
		0644, // 파일 권한
	)

	defer file2.Close()

	file2.WriteString("이 내용은 파일 끝에 추가돼요!\n")
	fmt.Println("파일 쓰기 완료!")
}

func openFilePattern() {
	// 읽기 전용으로 파일열기
	file, err := os.Open("02_버퍼링_vs_논버퍼링.png")
	if err != nil {
		fmt.Println("파일 열기 실패:", err)
		return
	}

	// ❗❗ 이걸 빼먹으면 파일 핸들이 계속 열려있어서 리소스 낭비
	defer file.Close()

	// 파일 정보 가져오기
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("파일 정보 가져오기 실패:", err)
		return
	}

	fmt.Println("파일 이름:", fileInfo.Name())
	fmt.Println("파일 크기:", fileInfo.Size())
	fmt.Println("수정 시간:", fileInfo.ModTime())
	fmt.Println("디렉토리?:", fileInfo.IsDir())
	fmt.Println("권한:", fileInfo.Mode())
}
