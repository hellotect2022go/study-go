package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// 파일과 스트림을 다룰 때는 에러 처리가 정말 중요해.
// 실무에서 쓰이는 안전한 패턴들을 알아보자! 🛡️

func main() {
	// 이 패턴은 에러가 발생해도 리소스가 제대로 정리되도록 보장해줘! ✨
	// deferDeletePattern()

	// 네트워크 스트림이나 느린 I/O 작업에는 타임아웃이 필수야:
	// contextTimeoutPattern()

	errorWrappingPattern()
}

// 안전한 파일 복사 함수
func safeCopyFile(src, dst string) (err error) {
	// 소스 파일 열기
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("소스 파일 열기 실패: %w", err)
	}
	defer sourceFile.Close() // 함수 종료 시 자동으로 닫힘

	// 목적지 파일 생성
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("목적지 파일 생성 실패: %w", err)
	}
	// 에러가 발생하면 파일 삭제
	defer func() {
		destFile.Close()
		if err != nil {
			os.Remove(dst) // 실패 시 불완전한 파일 삭제
		}
	}()

	// 복사
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("복사 실패: %w", err)
	}

	// Sync로 디스크에 확실히 쓰기
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("동기화 실패: %w", err)
	}

	return nil
}

func deferDeletePattern() {
	err := safeCopyFile("source.txt", "destination.txt")
	if err != nil {
		fmt.Printf("파일 복사 실패: %v\n", err)
		return
	}

	fmt.Println("파일 복사 성공!")
}

// 타임아웃이 있는 파일 읽기
func readFileWithTimeout(filename string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 결과를 받을 채널
	resultChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	// 고루틴에서 파일 읽기
	go func() {
		file, err := os.Open(filename)
		if err != nil {
			errorChan <- err
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			errorChan <- err
			return
		}

		resultChan <- data
	}()

	// 컨텍스트 타임아웃 또는 결과 대기
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("타임아웃: %w", ctx.Err())
	case err := <-errorChan:
		return nil, err
	case data := <-resultChan:
		return data, nil
	}
}

func contextTimeoutPattern() {
	// 5초 타임아웃으로 파일 읽기
	data, err := readFileWithTimeout("large_file.txt", 5*time.Second)
	if err != nil {
		fmt.Printf("읽기 실패: %v\n", err)
		return
	}

	fmt.Printf("읽은 데이터 크기: %d 바이트\n", len(data))
}

// 커스텀 에러 타입
type FileProcessError struct {
	Filename string
	Op       string
	Err      error
}

func (e *FileProcessError) Error() string {
	return fmt.Sprintf("파일 처리 에러 [%s, %s]: %v", e.Filename, e.Op, e.Err)
}

func (e *FileProcessError) Unwrap() error {
	return e.Err
}

// 파일 처리 함수
func processFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return &FileProcessError{
			Filename: filename,
			Op:       "open",
			Err:      err,
		}
	}
	defer file.Close()

	data := make([]byte, 1024)
	_, err = file.Read(data)
	if err != nil && err != io.EOF {
		return &FileProcessError{
			Filename: filename,
			Op:       "read",
			Err:      err,
		}
	}

	// 데이터 처리...
	fmt.Println(string(data))

	return nil
}

func errorWrappingPattern() {
	err := processFile("nonexistent.txt")
	if err != nil {
		// 에러 타입 확인
		var fileErr *FileProcessError
		if errors.As(err, &fileErr) {
			fmt.Printf("파일 에러 발생!\n")
			fmt.Printf("파일명: %s\n", fileErr.Filename)
			fmt.Printf("작업: %s\n", fileErr.Op)
			fmt.Printf("원인: %v\n", fileErr.Err)
		}

		// 특정 에러 확인
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("파일이 존재하지 않습니다.")
		}
	}
}
