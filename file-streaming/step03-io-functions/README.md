# Step 3: io 패키지 유용한 함수들

## 🎯 학습 목표
- `io.Copy()` 완벽하게 사용하기
- `io.CopyN()`, `io.ReadAll()` 활용
- `io.MultiReader`, `io.MultiWriter` 이해
- 고급 io 함수들 마스터

## 📚 개요

`io` 패키지는 Reader/Writer를 다루는 편리한 함수들을 제공합니다.
이 함수들을 알면 코드가 훨씬 깔끔해집니다! 🎨

## 📋 io.Copy - 데이터 복사의 마법사

### 시그니처
```go
func Copy(dst Writer, src Reader) (written int64, err error)
```

### 동작 방식
- `src`에서 데이터를 읽어 `dst`에 씀
- EOF까지 모든 데이터 복사
- 내부적으로 32KB 버퍼 사용
- 총 복사한 바이트 수 반환

### 핵심 장점
✅ 버퍼를 직접 관리할 필요 없음
✅ 자동으로 효율적인 크기 사용
✅ 코드가 매우 간결함

### 최적화 자동 적용
`io.Copy`는 다음을 자동 감지하여 최적화:
- `src`가 `io.WriterTo` 구현 시 → `WriteTo()` 사용
- `dst`가 `io.ReaderFrom` 구현 시 → `ReadFrom()` 사용

**이것이 Go 인터페이스 조합의 힘!** ⚡

## 🎯 io.CopyN - 정확한 양만큼

### 시그니처
```go
func CopyN(dst Writer, src Reader, n int64) (written int64, err error)
```

### 동작 방식
- 정확히 `n` 바이트만 복사
- `n` 바이트를 다 읽기 전에 EOF 만나면 에러

### 사용 예시
```
파일의 처음 1MB만 복사하고 싶을 때
네트워크에서 정확히 100바이트만 읽을 때
헤더 크기가 고정된 프로토콜 처리
```

## 📖 io.ReadAll - 전부 다 읽기

### 시그니처
```go
func ReadAll(r Reader) ([]byte, error)
```

### 동작 방식
- Reader에서 EOF까지 모두 읽기
- 결과를 바이트 슬라이스로 반환
- Go 1.16 이전에는 `ioutil.ReadAll`

### ⚠️ 주의사항
**대용량 파일에 절대 사용 금지!**

| 파일 크기 | ReadAll 사용 | 결과 |
|---------|------------|------|
| 10KB | ✅ OK | 안전 |
| 10MB | ⚠️ 주의 | 가능하지만... |
| 10GB | ❌ 금지 | 메모리 터짐 💥 |

### 적합한 사용처
- 설정 파일 (몇 KB)
- HTTP 응답 본문 (크기 제한된)
- 작은 JSON/XML 데이터
- 테스트 데이터

## 🔗 io.MultiReader - 여러 Reader를 하나로

### 시그니처
```go
func MultiReader(readers ...Reader) Reader
```

### 동작 방식
- 여러 Reader를 순차적으로 연결
- 첫 번째가 EOF면 두 번째로 자동 이동
- 하나의 Reader처럼 사용 가능

### 사용 예시
```
파일 헤더 + 본문 + 푸터 합치기
여러 파일을 하나로 연결
고정 데이터 + 가변 데이터 조합
```

### 실용 사례
```
로그 파일 합치기:
  2024-01.log
+ 2024-02.log  → 하나의 Reader로 처리
+ 2024-03.log
```

## 🔀 io.MultiWriter - 한 번에 여러 곳에 쓰기

### 시그니처
```go
func MultiWriter(writers ...Writer) Writer
```

### 동작 방식
- 하나의 Write 호출로 여러 Writer에 동시에 쓰기
- 모든 Writer에 성공해야 성공
- 하나라도 실패하면 에러 반환

### 사용 예시
```
로그를 파일 + 콘솔에 동시 출력
데이터를 원본 + 백업에 동시 저장
네트워크로 전송하면서 로컬에도 저장
```

### 실용 사례
**로깅 시스템**:
```
로그 메시지 → MultiWriter
              ├─→ 파일 (logs/app.log)
              ├─→ 콘솔 (os.Stdout)
              └─→ 네트워크 (로그 서버)
```

## 🔄 io.Pipe - 메모리 파이프

### 시그니처
```go
func Pipe() (*PipeReader, *PipeWriter)
```

### 동작 방식
- 메모리 상에서 Reader와 Writer 연결
- Write한 데이터를 Read로 바로 읽음
- 고루틴 간 데이터 전달에 유용

### 특징
- 버퍼 없음 (동기적)
- Writer가 Write할 때 Reader가 Read해야 함
- 고루틴과 함께 사용 필수

### 사용 패턴
```
고루틴 A: Writer에 데이터 씀
    ↓ (파이프)
고루틴 B: Reader에서 데이터 읽음
```

## 📏 io.LimitReader - 읽기 제한

### 시그니처
```go
func LimitReader(r Reader, n int64) Reader
```

### 동작 방식
- 최대 `n` 바이트만 읽도록 제한
- `n` 바이트 읽으면 EOF 반환

### 보안 활용
**DoS 공격 방지**:
```
악의적 사용자가 무한 데이터 전송
    ↓
LimitReader로 1MB 제한
    ↓
안전하게 처리 🛡️
```

### 사용 예시
```
HTTP 요청 본문 크기 제한
파일 업로드 크기 제한
네트워크 데이터 제한
```

## 🔀 io.TeeReader - 복제 읽기

### 시그니처
```go
func TeeReader(r Reader, w Writer) Reader
```

### 동작 방식
- Reader에서 읽으면서 동시에 Writer에 씀
- 원본 데이터는 그대로 반환
- 체크섬 계산, 로깅 등에 유용

### 사용 예시
```
파일 다운로드하면서 MD5 체크섬 계산
네트워크 데이터 읽으면서 로깅
파일 복사하면서 검증
```

### 동작 흐름
```
Reader → TeeReader → 데이터 반환
           ↓
         Writer (체크섬, 로그 등)
```

## 🎓 실습 과제

### 과제 1: io.Copy vs 수동 복사

다음 두 방법의 차이는?

**방법 A (io.Copy)**:
```go
io.Copy(dst, src)
```

**방법 B (수동)**:
```go
buffer := make([]byte, 4096)
for {
    n, err := src.Read(buffer)
    if n > 0 {
        dst.Write(buffer[:n])
    }
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }
}
```

**차이점**:
- A는 간결하고 자동 최적화
- B는 세밀한 제어 가능
- 대부분 A를 사용하는 것이 좋음

### 과제 2: MultiWriter 활용

로그를 3곳에 동시에 저장하려면?
1. 콘솔 (os.Stdout)
2. 파일 (logs/app.log)
3. 네트워크 (TCP 연결)

**힌트**: `io.MultiWriter` 사용

### 과제 3: LimitReader로 안전하게

사용자가 업로드하는 파일을 10MB로 제한하려면?

**힌트**: `io.LimitReader(file, 10*1024*1024)`

## 📊 함수 선택 가이드

| 상황 | 사용할 함수 |
|-----|----------|
| 전체 복사 | `io.Copy` |
| 일부만 복사 | `io.CopyN` |
| 작은 데이터 전부 읽기 | `io.ReadAll` |
| 여러 소스 연결 | `io.MultiReader` |
| 동시에 여러 곳 쓰기 | `io.MultiWriter` |
| 읽기 크기 제한 | `io.LimitReader` |
| 읽으면서 복제 | `io.TeeReader` |

## 🔑 핵심 요약

### 가장 많이 사용
1. **io.Copy** - 99% 상황에서 사용
2. **io.LimitReader** - 보안 필수
3. **io.MultiWriter** - 로깅 시스템

### 주의할 것
- ❌ 대용량에 `io.ReadAll` 사용 금지
- ✅ `io.Copy`는 자동 최적화됨
- ✅ `io.LimitReader`로 DoS 방지

### 조합의 힘
```go
// 파일 읽으면서 10MB 제한, 동시에 체크섬 계산
limited := io.LimitReader(file, 10*1024*1024)
teed := io.TeeReader(limited, hash)
io.Copy(output, teed)
```

**인터페이스 조합으로 강력한 기능 구현!** 🚀

## ➡️ 다음 단계

**Step 4: 파일 처리 실전 테크닉**
- `os` 패키지로 파일 열기/닫기
- `bufio`로 효율적인 읽기/쓰기
- 청크 단위 대용량 파일 처리

