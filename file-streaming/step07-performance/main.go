package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	// 버퍼 크기는 성능에 큰 영향을 미쳐. 너무 작으면 시스템 콜이 많아지고, 너무 크면 메모리 낭비야:
	//bufferTestPattern()

	// 여러 파일을 동시에 처리하거나, 파이프라인을 구성해서 병렬 처리할 수 있어:
	//compressTestPattern()

	// sync.Pool을 사용하면 버퍼를 재사용해서 GC 압력을 줄일 수 있어:
	syncPoolTestPattern()

}

func copyWithBuffer(src, dst string, bufferSize int) (time.Duration, error) {
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	dest, err := os.Create(dst)

	if err != nil {
		return 0, err
	}
	defer dest.Close()

	buffer := make([]byte, bufferSize)

	start := time.Now()

	_, err = io.CopyBuffer(dest, source, buffer)
	elapsed := time.Since(start)

	return elapsed, err
}

func bufferTestPattern() {
	testFile := "test_large_file.dat"

	// 다양한 버퍼 크기 테스트
	bufferSizes := []int{
		1024,    // 1KB
		4096,    // 4KB
		8192,    // 8KB
		32768,   // 32KB
		65536,   // 64KB
		131072,  // 128KB
		1048576, // 1MB
	}

	fmt.Println("버퍼 크기별 성능 테스트")
	fmt.Println(strings.Repeat("-", 50))

	for _, size := range bufferSizes {
		elapsed, err := copyWithBuffer(testFile, "output.tmp", size)
		if err != nil {
			fmt.Printf("에러: %v\n", err)
			continue
		}

		fmt.Printf("버퍼 크기: %7d 바이트 -> 소요 시간: %v\n", size, elapsed)
		os.Remove("output.tmp")
	}
}

// 파일 압축 작업
func compressFile(inputPath, outputPath string) error {
	input, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	gzipWriter := gzip.NewWriter(output)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, input)
	return err
}

// 병렬로 여러 파일 압축
func compressFilesParallel(files []string, workers int) error {
	// 작업 채널
	jobs := make(chan string, len(files))
	// 결과 채널
	results := make(chan error, len(files))

	// 워커 고루틴 시작
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for inputFile := range jobs {
				outputFile := inputFile + ".gz"
				fmt.Printf("워커 %d: %s 압축 중...\n", workerID, inputFile)

				err := compressFile(inputFile, outputFile)
				results <- err

				if err != nil {
					fmt.Printf("워커 %d: 에러 - %v\n", workerID, err)
				} else {
					fmt.Printf("워커 %d: %s 완료!\n", workerID, inputFile)
				}
			}
		}(i)
	}

	// 작업 전송
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// 워커들이 끝날 때까지 대기
	wg.Wait()
	close(results)

	// 결과 확인
	errorCount := 0
	for err := range results {
		if err != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%d개 파일 압축 실패", errorCount)
	}

	return nil
}

func compressTestPattern() {
	// 압축할 파일 목록
	files := []string{
		"file1.txt",
		"file2.txt",
		"file3.txt",
		"file4.txt",
		"file5.txt",
	}

	// 4개의 워커로 병렬 처리
	fmt.Println("병렬 압축 시작...")
	err := compressFilesParallel(files, 4)
	if err != nil {
		fmt.Printf("압축 실패: %v\n", err)
		return
	}

	fmt.Println("모든 파일 압축 완료!")
}

// 버퍼 풀
var bufferPool = sync.Pool{
	New: func() interface{} {
		// 64KB 버퍼 생성
		buffer := make([]byte, 64*1024)
		return &buffer
	},
}

// 풀을 사용한 파일 복사
func copyFileWithPool(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	// 풀에서 버퍼 가져오기
	bufferPtr := bufferPool.Get().(*[]byte)
	buffer := *bufferPtr
	defer bufferPool.Put(bufferPtr) // 사용 후 반환

	// 복사
	_, err = io.CopyBuffer(dest, source, buffer)
	return err
}

func syncPoolTestPattern() {
	files := []string{"file1.txt", "file2.txt", "file3.txt"}

	var wg sync.WaitGroup
	for i, file := range files {
		wg.Add(1)
		go func(idx int, f string) {
			defer wg.Done()

			output := fmt.Sprintf("copy_%d.txt", idx)
			err := copyFileWithPool(f, output)
			if err != nil {
				fmt.Printf("복사 실패 %s: %v\n", f, err)
			} else {
				fmt.Printf("복사 완료: %s -> %s\n", f, output)
			}
		}(i, file)
	}

	wg.Wait()
	fmt.Println("모든 복사 완료!")
}
