# Step 5: 스트리밍 고급 기법

## 🎯 학습 목표
- `io.Pipe`로 고루틴 간 데이터 스트리밍
- 커스텀 Reader/Writer 구현
- `io.LimitReader`로 DoS 방지
- `io.TeeReader`로 복제 처리

## 🔄 io.Pipe - 동시성 처리의 핵심

### 개념
- Reader와 Writer를 메모리로 연결
- 한쪽에서 Write하면 다른쪽에서 Read
- **고루틴 간 데이터 파이프라인 구축**

### 기본 패턴
```
고루틴 A (Writer)  →  Pipe  →  고루틴 B (Reader)
     파일 읽기           ↓           압축
                    실시간 전달
```

### 장점
✅ 메모리에 전체 데이터 안 올림
✅ 동시에 처리 (파이프라인)
✅ 자연스러운 백프레셔 (버퍼 없음)

### 특징
- **버퍼 없음**: Write가 Read를 기다림
- **동기적**: 반드시 고루틴과 함께 사용
- **에러 전파**: `CloseWithError()` 사용 가능

## 💡 실용 예시: 파일 읽으면서 압축

### 흐름
```
1. 고루틴 A: 파일 → Pipe Writer
              ↓
2. 고루틴 B: Pipe Reader → gzip → 출력 파일
```

### 장점
- 10GB 파일을 메모리 1MB로 처리
- 읽기와 압축이 동시 진행
- 전체 시간 단축

## 🎭 커스텀 Reader/Writer 구현

### io.Reader 구현
```go
type MyReader struct {
    // 필드들...
}

func (m *MyReader) Read(p []byte) (n int, err error) {
    // 구현...
    return n, err
}
```

### io.Writer 구현
```go
type MyWriter struct {
    // 필드들...
}

func (m *MyWriter) Write(p []byte) (n int, err error) {
    // 구현...
    return n, err
}
```

## 🎨 커스텀 Reader 예시

### UpperCaseReader - 대문자 변환
**기능**: 읽는 데이터를 모두 대문자로 변환

**구조**:
```
원본 Reader → UpperCaseReader → 대문자 데이터
```

**사용 사례**:
- 텍스트 정규화
- 대소문자 무시 검색
- 데이터 변환 파이프라인

### LineNumberWriter - 줄 번호 추가
**기능**: 각 줄 앞에 번호 추가

**출력 예시**:
```
1: 첫 번째 줄
2: 두 번째 줄
3: 세 번째 줄
```

**사용 사례**:
- 로그 파일 번호 매기기
- 소스 코드 출력
- 디버깅

## ⚡ io.LimitReader - 안전한 읽기

### 보안 필수!
```go
func LimitReader(r Reader, n int64) Reader
```

### DoS 공격 방지

#### 공격 시나리오
```
악의적 사용자 → 무한 데이터 전송 → 서버 메모리 고갈
```

#### 방어
```go
limited := io.LimitReader(request.Body, 10*1024*1024) // 10MB 제한
data, _ := io.ReadAll(limited)  // 안전!
```

### 적용 분야
| 분야 | 제한 |
|-----|------|
| 파일 업로드 | 10MB ~ 100MB |
| HTTP 요청 본문 | 1MB ~ 10MB |
| 웹소켓 메시지 | 64KB ~ 1MB |
| API 요청 | 1MB |

### 중요성
- ✅ **필수 보안 조치**
- ✅ 서버 리소스 보호
- ✅ 서비스 안정성 확보

## 🔀 io.TeeReader - 복제 처리

### 개념
```go
func TeeReader(r Reader, w Writer) Reader
```

### 동작
```
Reader → TeeReader → 원본 데이터 반환
           ↓
         Writer (복사본)
```

### 실용 예시

#### 1. 파일 다운로드 + 체크섬
```
네트워크 → TeeReader → 파일 저장
              ↓
         MD5 Hash 계산
```

#### 2. 데이터 처리 + 로깅
```
입력 → TeeReader → 처리
          ↓
       로그 파일
```

#### 3. 백업 + 전송
```
파일 → TeeReader → 네트워크 전송
          ↓
      로컬 백업
```

## 🎪 고급 패턴 조합

### 체인 연결
```go
// 파일 → 10MB 제한 → 체크섬 계산 → 출력
limited := io.LimitReader(file, 10*1024*1024)
teed := io.TeeReader(limited, hash)
io.Copy(output, teed)
```

### 레이어 쌓기
```
파일 Reader
  ↓
LimitReader (보안)
  ↓
CustomReader (변환)
  ↓
TeeReader (체크섬)
  ↓
최종 출력
```

## 🎓 실습 과제

### 과제 1: 파일 읽으면서 gzip 압축

**요구사항**:
- 원본 파일을 읽어서 압축
- `io.Pipe` 사용
- 메모리에 전체 파일 올리지 않기

**구조**:
```
고루틴 1: 파일 → Pipe Writer
고루틴 2: Pipe Reader → gzip → 출력
```

### 과제 2: ROT13 Reader 구현

**기능**: 
- ROT13 암호화 (알파벳 13칸 이동)
- A → N, B → O, ..., N → A, ...

**구현**:
```go
type ROT13Reader struct {
    r io.Reader
}

func (rot *ROT13Reader) Read(p []byte) (n int, err error) {
    // TODO: 구현
}
```

### 과제 3: 파일 다운로드 + 검증

**요구사항**:
1. HTTP에서 파일 다운로드
2. 동시에 SHA256 체크섬 계산
3. 다운로드 완료 후 체크섬 출력

**힌트**: `io.TeeReader` + `crypto/sha256`

### 과제 4: 안전한 업로드 핸들러

**요구사항**:
1. HTTP 파일 업로드 받기
2. 10MB 크기 제한
3. 허용된 확장자만 (jpg, png, pdf)
4. 안전하게 저장

**보안 체크리스트**:
- ✅ `io.LimitReader` 사용
- ✅ 파일 이름 검증
- ✅ 확장자 확인
- ✅ 에러 처리

## 🔍 성능 비교

### 압축 방식 비교

#### 방법 A: 전체 로드 후 압축
```
파일 → 메모리 (10GB) → 압축 → 출력
```
- 메모리: 10GB
- 시간: 60초

#### 방법 B: 스트리밍 압축 (io.Pipe)
```
파일 → Pipe → 압축 (동시) → 출력
```
- 메모리: 1MB
- 시간: 40초

**스트리밍이 메모리도 적게 쓰고 더 빠름!** 🚀

## 🔑 핵심 요약

### io.Pipe
```go
pr, pw := io.Pipe()

go func() {
    defer pw.Close()
    // pw에 데이터 쓰기
}()

// pr에서 데이터 읽기
```

### 커스텀 Reader/Writer
```go
type MyReader struct {
    source io.Reader
}

func (m *MyReader) Read(p []byte) (n int, err error) {
    n, err = m.source.Read(p)
    // 데이터 변환
    return n, err
}
```

### 보안 필수
```go
// 반드시 크기 제한!
limited := io.LimitReader(untrustedReader, maxBytes)
```

### 복제 처리
```go
// 읽으면서 동시에 다른 곳에 쓰기
teed := io.TeeReader(reader, writer)
```

### 조합의 힘
```
인터페이스 조합 → 강력한 파이프라인 구축
```

## ➡️ 다음 단계

**Step 6: 실전 프로젝트 - 로그 파일 분석기**
- 배운 모든 기술 종합 적용
- 대용량 로그 파일 스트리밍 분석
- 정규표현식, 통계, 리포트 생성

