# Step 8: 에러 처리와 안전성

## 🎯 학습 목표
- `defer`로 안전한 리소스 관리
- `context`로 타임아웃 처리
- Go 1.13+ 에러 래핑
- 프로덕션 레벨 에러 처리

## ✅ defer - 리소스 정리의 핵심

### 기본 원리

#### defer의 실행 시점
```
함수 시작
  ↓
defer 등록
  ↓
작업 수행
  ↓
return / panic
  ↓
defer 실행 (역순!)
```

### 필수 패턴

#### 파일 처리
```go
file, err := os.Open("data.txt")
if err != nil {
    return err
}
defer file.Close()  // 👈 필수!

// 안전하게 작업...
```

#### 여러 리소스
```go
src, _ := os.Open("source.txt")
defer src.Close()  // 2번째 실행

dst, _ := os.Create("dest.txt")
defer dst.Close()   // 1번째 실행 (LIFO)

// 복사 작업...
```

### 고급 패턴: 롤백

#### 에러 시 파일 삭제
```go
func safeCopy(src, dst string) (err error) {
    source, err := os.Open(src)
    if err != nil {
        return err
    }
    defer source.Close()

    dest, err := os.Create(dst)
    if err != nil {
        return err
    }
    
    // 에러 발생 시 불완전한 파일 삭제
    defer func() {
        dest.Close()
        if err != nil {
            os.Remove(dst)  // 롤백!
        }
    }()

    _, err = io.Copy(dest, source)
    if err != nil {
        return err
    }

    // Sync로 디스크에 확실히 쓰기
    return dest.Sync()
}
```

### defer 주의사항

#### ❌ 루프 안에서 defer
```go
// 잘못된 예
for _, filename := range files {
    file, _ := os.Open(filename)
    defer file.Close()  // 루프 끝날 때까지 안 닫힘!
    // ... 작업
}
// 여기서 한꺼번에 닫힘 → 파일 핸들 고갈
```

#### ✅ 올바른 방법
```go
for _, filename := range files {
    func() {
        file, _ := os.Open(filename)
        defer file.Close()  // 함수 끝날 때 닫힘
        // ... 작업
    }()
}
```

## 🎯 context - 타임아웃과 취소

### 왜 필요한가?

#### 문제 상황
```
네트워크 요청 → 응답 없음 → 무한 대기 → 고루틴 누수
```

#### 해결
```
context.WithTimeout → 5초 제한 → 타임아웃 → 안전
```

### context 타입

| 함수 | 용도 |
|------|-----|
| `WithTimeout` | 시간 제한 |
| `WithDeadline` | 마감 시각 |
| `WithCancel` | 수동 취소 |
| `WithValue` | 값 전달 (비추천) |

### 타임아웃 패턴

#### 기본 구조
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()  // 반드시!

// 고루틴에서 작업
done := make(chan result)
go func() {
    // 시간 걸리는 작업
    done <- doWork()
}()

// 타임아웃 또는 완료 대기
select {
case <-ctx.Done():
    return fmt.Errorf("타임아웃: %w", ctx.Err())
case res := <-done:
    return res
}
```

### 실용 예시

#### 네트워크 요청
```
HTTP 요청에 10초 타임아웃
→ 서버 무응답 시 자동 중단
→ 리소스 누수 방지
```

#### 파일 처리
```
대용량 파일 처리에 30초 제한
→ 예상 시간 초과 시 중단
→ 무한 대기 방지
```

## 🔍 에러 래핑 (Go 1.13+)

### 전통적 방식 (비추천)
```go
if err != nil {
    return fmt.Errorf("파일 읽기 실패: %v", err)
    // 원본 에러 정보 손실!
}
```

### 에러 래핑 (권장)
```go
if err != nil {
    return fmt.Errorf("파일 읽기 실패: %w", err)
    // %w로 원본 에러 보존!
}
```

### 에러 체인

```
최상위: "사용자 데이터 처리 실패"
  ↓
중간: "파일 읽기 실패"
  ↓
원인: "permission denied"
```

### errors.Is - 에러 타입 확인

```go
if errors.Is(err, os.ErrNotExist) {
    // 파일이 없음
}

if errors.Is(err, io.EOF) {
    // 파일 끝
}
```

### errors.As - 에러 타입 변환

```go
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println("경로:", pathErr.Path)
    fmt.Println("작업:", pathErr.Op)
    fmt.Println("원인:", pathErr.Err)
}
```

## 🎭 커스텀 에러 타입

### 구조체 정의
```go
type FileProcessError struct {
    Filename string
    Op       string
    Err      error
}

func (e *FileProcessError) Error() string {
    return fmt.Sprintf("파일 처리 에러 [%s, %s]: %v", 
        e.Filename, e.Op, e.Err)
}

func (e *FileProcessError) Unwrap() error {
    return e.Err
}
```

### 사용
```go
if err != nil {
    return &FileProcessError{
        Filename: filename,
        Op:       "read",
        Err:      err,
    }
}
```

### 장점
- 풍부한 컨텍스트 정보
- 타입 기반 에러 처리
- 디버깅 용이

## 🛡️ 안전한 코딩 패턴

### 패턴 1: Fail Fast
```go
if err != nil {
    return err  // 빨리 실패
}
// 정상 경로 계속
```

### 패턴 2: 에러 전파
```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("작업 실패: %w", err)
}
```

### 패턴 3: 부분 성공 처리
```go
var errors []error
for _, item := range items {
    if err := process(item); err != nil {
        errors = append(errors, err)
        continue  // 다음 항목 계속
    }
}

if len(errors) > 0 {
    // 에러 요약 반환
}
```

### 패턴 4: 재시도
```go
const maxRetries = 3

for i := 0; i < maxRetries; i++ {
    err := tryOperation()
    if err == nil {
        return nil  // 성공
    }
    
    if i < maxRetries-1 {
        time.Sleep(time.Second * time.Duration(i+1))
    }
}

return fmt.Errorf("최대 재시도 횟수 초과")
```

## 🎓 실습 과제

### 과제 1: 안전한 파일 복사

**요구사항**:
1. 소스 파일 열기
2. 목적지 파일 생성
3. 복사 수행
4. 에러 발생 시 목적지 파일 삭제
5. 모든 리소스 정리

**체크리스트**:
- [ ] defer로 리소스 정리
- [ ] 에러 래핑
- [ ] 롤백 메커니즘

### 과제 2: 타임아웃이 있는 다운로드

**요구사항**:
1. HTTP에서 파일 다운로드
2. 30초 타임아웃 설정
3. 진행 상황 출력
4. 타임아웃 시 부분 파일 삭제

### 과제 3: 커스텀 에러 타입

**요구사항**:
1. FileOperationError 타입 정의
   - Filename
   - Operation (read/write/delete)
   - Timestamp
   - 원본 에러
2. Error() 메서드 구현
3. Unwrap() 메서드 구현
4. 실제 사용 예시

### 과제 4: 견고한 로그 분석기

**요구사항**:
1. 파일 열기 실패 대응
2. 읽기 에러 처리
3. 부분적 성공 처리 (일부 라인 에러)
4. 리소스 누수 방지
5. 타임아웃 설정 (선택)

## 📋 에러 처리 체크리스트

### 기본
- [ ] 모든 에러 확인
- [ ] 적절한 에러 메시지
- [ ] defer로 리소스 정리

### 고급
- [ ] 에러 래핑 (`%w`)
- [ ] 컨텍스트 타임아웃
- [ ] 커스텀 에러 타입
- [ ] 재시도 로직 (필요시)

### 프로덕션
- [ ] 로깅
- [ ] 모니터링
- [ ] 알림
- [ ] 복구 메커니즘

## 🔑 핵심 요약

### defer
```go
file, _ := os.Open("file.txt")
defer file.Close()  // 필수!

// 역순 실행 (LIFO)
defer cleanup1()  // 마지막 실행
defer cleanup2()  // 두 번째
defer cleanup3()  // 첫 번째
```

### context
```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

select {
case <-ctx.Done():
    // 타임아웃
case result := <-done:
    // 완료
}
```

### 에러 래핑
```go
// 래핑
return fmt.Errorf("작업 실패: %w", err)

// 확인
if errors.Is(err, os.ErrNotExist) { ... }

// 변환
var pathErr *os.PathError
if errors.As(err, &pathErr) { ... }
```

### 안전한 패턴
```
1. 에러 즉시 확인
2. 빨리 실패 (Fail Fast)
3. 리소스 정리 (defer)
4. 에러 래핑 (컨텍스트 보존)
```

## ➡️ 다음 단계

**Step 9: HTTP 스트리밍**
- 파일 다운로드 핸들러
- Range 요청 지원
- 파일 업로드 처리

