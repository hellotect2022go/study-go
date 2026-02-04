# Step 10: 테스트와 벤치마킹

## 🎯 학습 목표
- Go 테스트 프레임워크 활용
- 테이블 기반 테스트 작성
- 벤치마크로 성능 측정
- 프로파일링으로 병목 찾기

## 🧪 단위 테스트 작성

### 테스트 파일 규칙

| 규칙 | 예시 |
|------|------|
| 파일명 | `xxx_test.go` |
| 함수명 | `TestXxx(t *testing.T)` |
| 위치 | 같은 패키지 |

### 기본 구조

```go
import "testing"

func TestReadFile(t *testing.T) {
    // 준비 (Arrange)
    filename := "test.txt"
    
    // 실행 (Act)
    result, err := ReadFile(filename)
    
    // 검증 (Assert)
    if err != nil {
        t.Errorf("에러 발생: %v", err)
    }
    if result != expected {
        t.Errorf("예상: %v, 실제: %v", expected, result)
    }
}
```

### 실행

```bash
# 모든 테스트
go test

# 상세 출력
go test -v

# 특정 테스트만
go test -run TestReadFile

# 커버리지
go test -cover
```

## 📊 테이블 기반 테스트

### 장점
✅ 여러 케이스를 깔끔하게 관리
✅ 새 케이스 추가 쉬움
✅ 가독성 좋음

### 패턴

```go
func TestUpperReader(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"소문자", "hello", "HELLO"},
        {"대문자", "WORLD", "WORLD"},
        {"혼합", "Hello World", "HELLO WORLD"},
        {"숫자", "test123", "TEST123"},
        {"빈문자열", "", ""},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            reader := &UpperReader{
                r: strings.NewReader(tt.input),
            }
            
            result, err := io.ReadAll(reader)
            if err != nil {
                t.Fatalf("읽기 실패: %v", err)
            }
            
            if string(result) != tt.expected {
                t.Errorf("예상: %s, 실제: %s", 
                    tt.expected, string(result))
            }
        })
    }
}
```

### 출력 예시
```
=== RUN   TestUpperReader
=== RUN   TestUpperReader/소문자
=== RUN   TestUpperReader/대문자
=== RUN   TestUpperReader/혼합
=== RUN   TestUpperReader/숫자
=== RUN   TestUpperReader/빈문자열
--- PASS: TestUpperReader (0.00s)
    --- PASS: TestUpperReader/소문자 (0.00s)
    --- PASS: TestUpperReader/대문자 (0.00s)
    --- PASS: TestUpperReader/혼합 (0.00s)
    --- PASS: TestUpperReader/숫자 (0.00s)
    --- PASS: TestUpperReader/빈문자열 (0.00s)
PASS
```

## 🎯 에러 케이스 테스트

### 에러 발생 테스트

```go
func TestErrorReader(t *testing.T) {
    errorReader := &ErrorReader{
        err: io.ErrUnexpectedEOF,
    }
    
    _, err := io.ReadAll(errorReader)
    
    if err != io.ErrUnexpectedEOF {
        t.Errorf("예상 에러: %v, 실제: %v", 
            io.ErrUnexpectedEOF, err)
    }
}
```

### 경계 조건 테스트

```go
tests := []struct {
    name     string
    bufSize  int
    dataSize int
}{
    {"버퍼 == 데이터", 100, 100},
    {"버퍼 > 데이터", 100, 50},
    {"버퍼 < 데이터", 50, 100},
    {"빈 데이터", 100, 0},
    {"최소 버퍼", 1, 100},
}
```

## ⚡ 벤치마크 작성

### 기본 구조

```go
func BenchmarkCopy(b *testing.B) {
    data := bytes.Repeat([]byte("x"), 1024)  // 1KB
    
    b.ResetTimer()  // 준비 시간 제외
    
    for i := 0; i < b.N; i++ {
        var buf bytes.Buffer
        reader := bytes.NewReader(data)
        io.Copy(&buf, reader)
    }
}
```

### 실행

```bash
# 벤치마크 실행
go test -bench=.

# 메모리 포함
go test -bench=. -benchmem

# 실행 시간 조정
go test -bench=. -benchtime=10s
```

### 출력 해석

```
BenchmarkCopy-8    1000000    1234 ns/op    512 B/op    4 allocs/op
                   ^^^^^^^^   ^^^^^^^^^^    ^^^^^^^^    ^^^^^^^^^^^^
                   실행 횟수   ns/실행      바이트/실행  할당/실행
```

## 📊 다양한 벤치마크 패턴

### 1. 버퍼 크기 비교

```go
func BenchmarkBufferSizes(b *testing.B) {
    data := bytes.Repeat([]byte("x"), 1024*1024)  // 1MB
    sizes := []int{512, 1024, 4096, 32768, 65536}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
            buffer := make([]byte, size)
            b.ResetTimer()
            b.SetBytes(int64(len(data)))
            
            for i := 0; i < b.N; i++ {
                reader := bytes.NewReader(data)
                var out bytes.Buffer
                io.CopyBuffer(&out, reader, buffer)
            }
        })
    }
}
```

### 출력
```
BenchmarkBufferSizes/size_512-8     1000  1200 ns/op  854 MB/s
BenchmarkBufferSizes/size_1024-8    2000   600 ns/op  1707 MB/s
BenchmarkBufferSizes/size_4096-8    3000   400 ns/op  2560 MB/s
BenchmarkBufferSizes/size_32768-8   5000   200 ns/op  5120 MB/s
BenchmarkBufferSizes/size_65536-8   5000   190 ns/op  5389 MB/s
```

### 2. 메모리 풀 비교

```go
func BenchmarkWithPool(b *testing.B) {
    pool := sync.Pool{
        New: func() interface{} {
            buf := make([]byte, 4096)
            return &buf
        },
    }
    data := bytes.Repeat([]byte("x"), 1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bufPtr := pool.Get().(*[]byte)
        reader := bytes.NewReader(data)
        var out bytes.Buffer
        io.CopyBuffer(&out, reader, *bufPtr)
        pool.Put(bufPtr)
    }
}

func BenchmarkWithoutPool(b *testing.B) {
    data := bytes.Repeat([]byte("x"), 1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        buffer := make([]byte, 4096)
        reader := bytes.NewReader(data)
        var out bytes.Buffer
        io.CopyBuffer(&out, reader, buffer)
    }
}
```

### 출력
```
BenchmarkWithPool-8      2000000   800 ns/op   0 B/op  0 allocs/op
BenchmarkWithoutPool-8   1000000  1500 ns/op  4096 B/op  1 allocs/op
```

**풀 사용 시: 2배 빠르고 할당 0번!**

## 🔍 프로파일링

### CPU 프로파일

```bash
# 프로파일 생성
go test -bench=. -cpuprofile=cpu.prof

# 분석
go tool pprof cpu.prof

# 인터랙티브 모드
(pprof) top
(pprof) list functionName
(pprof) web  # 그래프 (graphviz 필요)
```

### 메모리 프로파일

```bash
# 프로파일 생성
go test -bench=. -memprofile=mem.prof

# 분석
go tool pprof mem.prof

# 할당 확인
(pprof) top -cum
(pprof) list functionName
```

### 웹 UI

```bash
# 웹 인터페이스
go tool pprof -http=:8080 cpu.prof
```

**브라우저에서 플레임 그래프 확인!** 🔥

## 📈 커버리지 확인

### 실행

```bash
# 커버리지 측정
go test -coverprofile=coverage.out

# 결과 보기
go tool cover -func=coverage.out

# HTML 리포트
go tool cover -html=coverage.out
```

### 출력 예시

```
file.go:10:   ReadFile    100.0%
file.go:20:   WriteFile    85.7%
file.go:30:   ProcessData  66.7%
total:                     84.1%
```

### 목표

| 프로젝트 타입 | 목표 커버리지 |
|------------|-------------|
| 라이브러리 | 80% 이상 |
| 서비스 | 70% 이상 |
| 내부 도구 | 50% 이상 |

## 🎓 실습 과제

### 과제 1: Reader 테스트

**요구사항**:
커스텀 Reader의 테스트 작성
1. 정상 케이스
2. EOF 처리
3. 에러 케이스
4. 경계 조건

### 과제 2: 버퍼 크기 벤치마크

**요구사항**:
다양한 버퍼 크기 성능 비교
- 512B, 1KB, 4KB, 8KB, 16KB, 32KB, 64KB, 128KB, 1MB
- 처리량 (MB/s) 계산
- 최적 크기 찾기

### 과제 3: 메모리 풀 효과 측정

**요구사항**:
`sync.Pool` 사용 전/후 비교
1. 실행 시간
2. 메모리 할당 횟수
3. 할당 바이트 수
4. 결론 도출

### 과제 4: 로그 분석기 벤치마크

**요구사항**:
Step 6의 로그 분석기 성능 측정
1. 다양한 크기 파일 (1MB, 10MB, 100MB)
2. 다양한 버퍼 크기
3. 처리 속도 (MB/s)
4. 병목 지점 프로파일링

## 🔧 테스트 헬퍼

### 임시 파일 생성

```go
func createTempFile(t *testing.T, content string) string {
    t.Helper()
    
    tmpfile, err := os.CreateTemp("", "test")
    if err != nil {
        t.Fatal(err)
    }
    
    defer tmpfile.Close()
    
    if _, err := tmpfile.Write([]byte(content)); err != nil {
        t.Fatal(err)
    }
    
    // 테스트 종료 시 삭제
    t.Cleanup(func() {
        os.Remove(tmpfile.Name())
    })
    
    return tmpfile.Name()
}
```

### 테이블 검증 헬퍼

```go
func assertEqual(t *testing.T, got, want interface{}) {
    t.Helper()
    
    if got != want {
        t.Errorf("예상: %v, 실제: %v", want, got)
    }
}
```

## 🔑 핵심 요약

### 테스트

```go
func TestXxx(t *testing.T) {
    // 테이블 기반
    tests := []struct{
        name     string
        input    interface{}
        expected interface{}
    }{
        // 케이스들...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 테스트 로직
        })
    }
}
```

### 벤치마크

```go
func BenchmarkXxx(b *testing.B) {
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        // 측정할 코드
    }
}
```

### 프로파일링

```bash
# CPU
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# 메모리
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

### 커버리지

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## ➡️ 다음 단계

**Step 11: 고급 패턴과 베스트 프랙티스**
- 파이프라인 패턴
- 어댑터 패턴
- 실무 베스트 프랙티스

