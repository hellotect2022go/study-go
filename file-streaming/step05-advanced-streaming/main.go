package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
)

// ⭐ io.Pipe는 Reader와 Writer를 연결해주는 메모리 파이프
func main() {
	//ioPipePattern()
	//customReaderWriterPattern()
	//limitReaderPattern()
	teeReaderPattern()
}

func ioPipePattern() {
	// 파이프 생성
	// ⭐ pr & pw 는 동일 메모리 버퍼를 공유한다.
	pr, pw := io.Pipe()

	// 고루틴에서 데이터 쓰기
	go func() {
		defer pw.Close()

		// 원본 파일 읽기
		file, err := os.Open("fake.log")

		if err != nil {
			pw.CloseWithError(err)
			return
		}
		defer file.Close()

		// 파이프로 복사
		_, err = io.Copy(pw, file)
		if err != nil {
			pw.CloseWithError(err)
		}

	}()

	// 메인 고루틴에서 압축하며 읽기
	outFile, err := os.Create("compressed.zip")
	if err != nil {
		fmt.Printf("출력 파일 생성 실패: %v\n", err)
		return
	}
	defer outFile.Close()

	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// 파이프에서 읽으면서 동시에 압축
	written, err := io.Copy(gzipWriter, pr)
	if err != nil {
		fmt.Printf("압축 실패: %v\n", err)
		return
	}

	fmt.Printf("총 %d 바이트를 압축했어요!\n", written)
}

// 대문자로 변환하는 Reader
type UpperCaseReader struct {
	source io.Reader
}

func (u *UpperCaseReader) Read(p []byte) (int, error) {
	n, err := u.source.Read(p)
	if err != nil {
		return n, err
	}
	copy(p, bytes.ToUpper(p[:n]))
	return n, nil
}

// 각 줄에 번호를 붙이는 Writer
type LineNumberWriter struct {
	dest       io.Writer
	lineNumber int
	newLine    bool
}

func (l *LineNumberWriter) Write(p []byte) (int, error) {
	written := 0
	for i, b := range p {
		if l.newLine {
			prefix := fmt.Sprintf("%d: ", l.lineNumber)
			l.dest.Write([]byte(prefix))
			l.lineNumber++
			l.newLine = false
		}

		// 줄바꿈처리
		if b == '\n' {
			l.dest.Write([]byte(p[written : i+1]))
			written = i + 1

			l.newLine = true
		}
	}

	// 마지막 줄 처리
	if written < len(p) {
		n, _ := l.dest.Write([]byte(p[written:]))
		written += n
	}

	return written, nil
}

func customReaderWriterPattern() {
	// 테스트
	testData := "Hello, World!\nThis is a test.\nThis is another test.\nLast line."

	// 대문자 변환 Reader
	upperReader := &UpperCaseReader{
		source: strings.NewReader(testData),
	}

	// 줄번호 Writer
	lineNumberWriter := &LineNumberWriter{
		dest:       os.Stdout,
		lineNumber: 1,
		newLine:    true,
	}

	// ⭐ 내부적을 read()  write() 메서드를 호출하며 데이터를 처리
	// 즉 데이타를 읽는 과정에서 대문자 처리 , 쓰는 과정에서 줄번호 처리
	io.Copy(lineNumberWriter, upperReader)
}

func limitReaderPattern() {
	// 긴문자열
	longText := strings.Repeat("Hello, World! ", 1000000)
	reader := strings.NewReader(longText)

	// 최대 100바이트만 읽기
	// ⭐ 읽는 양을 제한하는 Reader 를 반환 즉 100byte 이후에 EOF 가 반환됨
	limited := io.LimitReader(reader, 100)
	// ⭐ 읽은 데이터를 모두 읽어서 반환
	b, _ := io.ReadAll(limited)

	fmt.Println(string(b))
	fmt.Println("읽은 바이트:", len(b))

	fmt.Println("--------------------------------")

	buff := make([]byte, 100)

	// ⭐ buff 만큼 읽은 다음 cursor 는 100byte 이후에 있음
	reader.Read(buff)
	fmt.Println(string(buff))
	fmt.Println("읽은 바이트:", len(buff))

}

// ⭐ io.TeeReader는 Reader 를 읽으면서 동시에 Writer 에 씀
func teeReaderPattern() {
	data := "이 데이터의 체크섬을 계산하면서 파일에도 저장할 거예요!"
	reader := strings.NewReader(data)

	// 체크썸 계산을 위한 해시
	hash := md5.New()

	// ⭐ TeeReader: reader 를 읽으면서 hash에도 쓰기
	// ⭐ 일종의 파이프 reader -> hash , teeReader(원본데이타)
	teeReader := io.TeeReader(reader, hash)

	// 파일에 저장
	file, _ := os.Create("tee_reader.txt")
	defer file.Close()

	// 복사하면 팡일에 쓰이면 서 동시에 해시도 계산
	written, _ := io.Copy(file, teeReader)

	// 체크썸 출력
	checksum := hash.Sum(nil)
	fmt.Printf("체크썸: %x\n", checksum)
	fmt.Printf("쓴 바이트: %d\n", written)

}
