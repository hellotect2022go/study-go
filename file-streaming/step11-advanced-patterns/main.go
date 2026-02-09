package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

// 진행률 콜백 함수 타입
type ProgressCallback func(current, total int64)

// 진행률을 추적하는 Reader 어댑터
type ProgressReader struct {
	reader   io.Reader
	total    int64
	current  int64
	callback ProgressCallback
}

func NewProgressReader(r io.Reader, total int64, callback ProgressCallback) *ProgressReader {
	return &ProgressReader{
		reader:   r,
		total:    total,
		callback: callback,
	}
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.current += int64(n)

	if pr.callback != nil {
		pr.callback(pr.current, pr.total)
	}

	return n, err
}

// 속도 제한 Reader 어댑터
type ThrottledReader struct {
	reader      io.Reader
	bytesPerSec int64
	lastRead    time.Time
}

func NewThrottledReader(r io.Reader, bytesPerSec int64) *ThrottledReader {
	return &ThrottledReader{
		reader:      r,
		bytesPerSec: bytesPerSec,
		lastRead:    time.Now(),
	}
}

func (tr *ThrottledReader) Read(p []byte) (n int, err error) {
	// 읽을 수 있는 최대 바이트 계산
	elapsed := time.Since(tr.lastRead)
	allowedBytes := int64(float64(tr.bytesPerSec) * elapsed.Seconds())

	if allowedBytes < int64(len(p)) {
		p = p[:allowedBytes]
	}

	if len(p) == 0 {
		// 대기
		time.Sleep(time.Millisecond * 10)
		return 0, nil
	}

	n, err = tr.reader.Read(p)
	tr.lastRead = time.Now()

	return n, err
}

func main() {
	file, _ := os.Open("fake.log")
	defer file.Close()

	fileInfo, _ := file.Stat()

	// 진행률 콜백
	progressCallback := func(current, total int64) {
		percent := float64(current) / float64(total) * 100
		fmt.Printf("\r진행률: %.2f%%", percent)
	}

	// 진행률 추적 Reader
	progressReader := NewProgressReader(file, fileInfo.Size(), progressCallback)

	// 속도 제한 Reader (1MB/s)
	throttledReader := NewThrottledReader(progressReader, 1024*1024)

	// 데이터 읽기
	io.Copy(io.Discard, throttledReader)
	fmt.Println("\n완료!")
}
