# Step 11: 고급 패턴과 베스트 프랙티스

## 🎯 학습 목표
- 파이프라인 패턴 마스터
- 어댑터 패턴 활용
- 실무 베스트 프랙티스 습득
- 프로덕션 레디 코드 작성

## 🔄 파이프라인 패턴

### 개념

여러 처리 단계를 파이프처럼 연결하여 데이터가 흐르도록 하는 패턴

### 장점
✅ 각 단계가 독립적
✅ 재사용 가능한 컴포넌트
✅ 테스트 용이
✅ 확장성 좋음
✅ 동시성 자연스럽게 적용

### 기본 구조

```
입력 → 단계1 → 단계2 → 단계3 → 출력

예시:
파일 → 읽기 → 압축 → 암호화 → 저장
```

## 🎨 파이프라인 예시: 파일 처리

### 단계별 함수

#### 1단계: 파일 읽기
```go
func readFile(filename string) (io.ReadCloser, error) {
    return os.Open(filename)
}
```

#### 2단계: 압축
```go
func compressStream(r io.Reader) io.Reader {
    pr, pw := io.Pipe()
    
    go func() {
        gzipWriter := gzip.NewWriter(pw)
        io.Copy(gzipWriter, r)
        gzipWriter.Close()
        pw.Close()
    }()
    
    return pr
}
```

#### 3단계: 체크섬 계산
```go
func checksumStream(r io.Reader) (io.Reader, *hash.Hash) {
    h := md5.New()
    teeReader := io.TeeReader(r, h)
    return teeReader, &h
}
```

#### 4단계: 파일 쓰기
```go
func writeFile(filename string, r io.Reader) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    _, err = io.Copy(file, r)
    return err
}
```

### 파이프라인 연결

```go
func processPipeline(inputFile, outputFile string) (string, error) {
    // 1. 파일 읽기
    reader, err := readFile(inputFile)
    if err != nil {
        return "", fmt.Errorf("파일 읽기 실패: %w", err)
    }
    defer reader.Close()
    
    // 2. 압축
    compressed := compressStream(reader)
    
    // 3. 체크섬 계산
    checksummed, hash := checksumStream(compressed)
    
    // 4. 파일 쓰기
    err = writeFile(outputFile, checksummed)
    if err != nil {
        return "", fmt.Errorf("파일 쓰기 실패: %w", err)
    }
    
    // 체크섬 반환
    checksum := hex.EncodeToString((*hash).Sum(nil))
    return checksum, nil
}
```

### 실행

```
input.txt (10GB)
  ↓ 읽기 (스트리밍)
  ↓ 압축 (동시)
  ↓ 체크섬 (동시)
  ↓ 쓰기
output.txt.gz (2GB)

메모리 사용: 64KB (일정!)
```

## 🎭 어댑터 패턴

### 개념

기존 인터페이스에 새로운 기능을 추가하는 래퍼

### 사용 사례
- 진행률 추적
- 속도 제한
- 로깅
- 통계 수집
- 데이터 변환

## 🎯 어댑터 예시 1: 진행률 추적

### ProgressReader

```go
type ProgressCallback func(current, total int64)

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
```

### 사용

```go
file, _ := os.Open("largefile.dat")
defer file.Close()

fileInfo, _ := file.Stat()

progress := NewProgressReader(file, fileInfo.Size(), 
    func(current, total int64) {
        percent := float64(current) / float64(total) * 100
        fmt.Printf("\r진행률: %.2f%%", percent)
    })

io.Copy(output, progress)
fmt.Println()
```

## ⏱️ 어댑터 예시 2: 속도 제한

### ThrottledReader

```go
type ThrottledReader struct {
    reader      io.Reader
    bytesPerSec int64
    lastTime    time.Time
    bucket      int64
}

func NewThrottledReader(r io.Reader, bytesPerSec int64) *ThrottledReader {
    return &ThrottledReader{
        reader:      r,
        bytesPerSec: bytesPerSec,
        lastTime:    time.Now(),
        bucket:      bytesPerSec,
    }
}

func (tr *ThrottledReader) Read(p []byte) (n int, err error) {
    // 토큰 버킷 알고리즘
    now := time.Now()
    elapsed := now.Sub(tr.lastTime)
    tr.lastTime = now
    
    // 버킷 충전
    tr.bucket += int64(float64(tr.bytesPerSec) * elapsed.Seconds())
    if tr.bucket > tr.bytesPerSec {
        tr.bucket = tr.bytesPerSec
    }
    
    // 읽을 수 있는 양 제한
    if tr.bucket <= 0 {
        time.Sleep(time.Millisecond * 10)
        return 0, nil
    }
    
    readSize := len(p)
    if int64(readSize) > tr.bucket {
        readSize = int(tr.bucket)
    }
    
    n, err = tr.reader.Read(p[:readSize])
    tr.bucket -= int64(n)
    
    return n, err
}
```

### 사용

```go
file, _ := os.Open("video.mp4")
defer file.Close()

// 1MB/s 속도 제한
throttled := NewThrottledReader(file, 1024*1024)

io.Copy(network, throttled)
```

## 🔗 어댑터 조합

여러 어댑터를 레이어처럼 쌓을 수 있습니다!

```go
file, _ := os.Open("data.bin")
defer file.Close()

fileInfo, _ := file.Stat()

// 레이어 1: 크기 제한 (보안)
limited := io.LimitReader(file, 100*1024*1024)  // 100MB

// 레이어 2: 속도 제한
throttled := NewThrottledReader(limited, 10*1024*1024)  // 10MB/s

// 레이어 3: 진행률
progress := NewProgressReader(throttled, fileInfo.Size(), 
    func(c, t int64) {
        fmt.Printf("\r%.2f%%", float64(c)/float64(t)*100)
    })

// 레이어 4: 체크섬
hash := sha256.New()
teed := io.TeeReader(progress, hash)

// 최종 처리
io.Copy(output, teed)

fmt.Printf("\nSHA256: %x\n", hash.Sum(nil))
```

**구조**:
```
파일
 ↓ LimitReader (보안)
 ↓ ThrottledReader (속도)
 ↓ ProgressReader (진행률)
 ↓ TeeReader (체크섬)
 ↓
출력
```

## 📋 베스트 프랙티스 체크리스트

### 리소스 관리
- [x] **항상 `defer`로 정리**
  ```go
  file, _ := os.Open("file.txt")
  defer file.Close()  // 필수!
  ```

- [x] **루프에서 defer 주의**
  ```go
  // ❌ 잘못된 방법
  for _, f := range files {
      file, _ := os.Open(f)
      defer file.Close()  // 루프 끝날 때까지 안 닫힘
  }
  
  // ✅ 올바른 방법
  for _, f := range files {
      func() {
          file, _ := os.Open(f)
          defer file.Close()  // 함수 끝날 때 닫힘
          // 처리...
      }()
  }
  ```

### 버퍼 크기
- [x] **적절한 버퍼 크기 사용 (32-64KB)**
  ```go
  buffer := make([]byte, 64*1024)  // 64KB
  ```

- [x] **상황에 따라 조정**
  - 네트워크: 4-8KB
  - 파일: 32-64KB
  - SSD: 64-128KB

### 에러 처리
- [x] **모든 에러 확인**
  ```go
  n, err := file.Read(buffer)
  if err != nil && err != io.EOF {
      return err
  }
  ```

- [x] **에러 래핑으로 컨텍스트 보존**
  ```go
  if err != nil {
      return fmt.Errorf("파일 처리 실패: %w", err)
  }
  ```

### 타임아웃
- [x] **네트워크 작업에 타임아웃 필수**
  ```go
  ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
  defer cancel()
  ```

### 대용량 처리
- [x] **스트리밍 방식 사용**
  ```go
  // ❌ 전체 로드
  data, _ := io.ReadAll(file)  // 10GB 파일 = 메모리 터짐
  
  // ✅ 스트리밍
  io.Copy(output, file)  // 일정한 메모리
  ```

- [x] **청크 단위 처리**
  ```go
  buffer := make([]byte, 1024*1024)  // 1MB 청크
  for {
      n, err := file.Read(buffer)
      if n > 0 {
          process(buffer[:n])
      }
      if err == io.EOF {
          break
      }
  }
  ```

### 병렬 처리
- [x] **I/O 바운드 작업은 병렬화**
  ```go
  // 워커 풀 패턴 사용
  workers := runtime.NumCPU()
  ```

- [x] **WaitGroup으로 동기화**
  ```go
  var wg sync.WaitGroup
  for _, task := range tasks {
      wg.Add(1)
      go func(t Task) {
          defer wg.Done()
          process(t)
      }(task)
  }
  wg.Wait()
  ```

### 메모리 최적화
- [x] **sync.Pool 활용**
  ```go
  var bufferPool = sync.Pool{
      New: func() interface{} {
          buf := make([]byte, 64*1024)
          return &buf
      },
  }
  ```

### 인터페이스 활용
- [x] **io.Reader/Writer 인터페이스로 유연하게**
  ```go
  func Process(r io.Reader) error {
      // 파일, 네트워크, 메모리 모두 처리 가능
  }
  ```

### 테스트
- [x] **단위 테스트 작성**
- [x] **벤치마크로 성능 검증**
- [x] **커버리지 70% 이상 목표**

## 🎓 실습 과제

### 과제 1: 완전한 파이프라인

**요구사항**:
파일 → 읽기 → 압축 → 암호화 → 체크섬 → 저장

각 단계를 독립적으로 구현하고 조합

### 과제 2: 다중 어댑터

**요구사항**:
- ProgressReader
- ThrottledReader
- LoggingReader
- StatisticsReader

모두 구현하고 조합 테스트

### 과제 3: 프로덕션 파일 서버

**요구사항**:
1. HTTP 파일 서버
2. 스트리밍 다운로드
3. Range 요청 지원
4. 업로드 (10MB 제한)
5. 진행률 API
6. 타임아웃
7. 로깅
8. 테스트

**베스트 프랙티스 체크리스트 모두 적용!**

## 🎉 최종 정리

### 학습한 핵심 기술

#### 기초
1. ✅ io.Reader/Writer 인터페이스
2. ✅ 스트리밍 vs 전체 로드
3. ✅ 버퍼링의 중요성

#### 실전
4. ✅ 파일 처리 (os, bufio)
5. ✅ io.Pipe 동시성
6. ✅ 커스텀 Reader/Writer
7. ✅ 로그 분석기 프로젝트

#### 최적화
8. ✅ 버퍼 크기 최적화
9. ✅ 병렬 처리
10. ✅ sync.Pool

#### 안전성
11. ✅ defer 리소스 관리
12. ✅ context 타임아웃
13. ✅ 에러 래핑

#### 응용
14. ✅ HTTP 스트리밍
15. ✅ Range 요청
16. ✅ 테스트/벤치마크

#### 패턴
17. ✅ 파이프라인 패턴
18. ✅ 어댑터 패턴
19. ✅ 베스트 프랙티스

### 실무 활용 분야

| 분야 | 기술 |
|------|------|
| 로그 처리 | 스트리밍, 병렬 처리 |
| 파일 전송 | HTTP 스트리밍, Range |
| 데이터 ETL | 파이프라인, 어댑터 |
| 백업 시스템 | 압축, 암호화, 체크섬 |
| API 서버 | 타임아웃, 에러 처리 |
| 모니터링 | 진행률, 통계 |

## 🚀 다음 단계

이제 여러분은:
- ✅ Go 파일 스트리밍 전문가
- ✅ 대용량 데이터 처리 가능
- ✅ 프로덕션 레벨 코드 작성
- ✅ 성능 최적화 능력

### 추가 학습 추천
1. **네트워크 프로그래밍**
   - TCP/UDP 스트리밍
   - WebSocket
   - gRPC 스트리밍

2. **분산 시스템**
   - 메시지 큐 (Kafka, RabbitMQ)
   - 분산 파일 시스템
   - 로드 밸런싱

3. **오픈소스 기여**
   - Go 표준 라이브러리 분석
   - 유명 프로젝트 기여
   - 자신의 라이브러리 제작

## 🎊 축하합니다!

**file-streaming 로드맵 완주를 축하드립니다!** 🎉

배운 내용을 실전 프로젝트에 적용하고,
계속해서 Go 프로그래밍의 세계를 탐험하세요!

**Happy Go Programming!** 🚀✨

---

*"Simple is better than complex."*
*"Concurrency is not parallelism."*
*"Don't communicate by sharing memory; share memory by communicating."*

— Go Proverbs

