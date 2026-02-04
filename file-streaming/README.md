# 🚀 Go 언어 파일 스트리밍 학습 로드맵

> intro.txt 기반 체계적 학습 가이드

## 📚 학습 개요

이 프로젝트는 Go 언어의 파일 처리와 스트리밍 데이터 처리를 단계별로 학습하기 위한 로드맵입니다.
대용량 파일을 메모리 효율적으로 처리하고, 실무에서 바로 활용할 수 있는 패턴을 익힙니다.

## 🎯 학습 목표

- ✅ Go의 `io.Reader`/`io.Writer` 인터페이스 완전 이해
- ✅ 스트리밍 방식의 데이터 처리 마스터
- ✅ 버퍼링을 통한 성능 최적화
- ✅ 대용량 파일 처리 실전 기술
- ✅ 동시성 처리와 고루틴 활용
- ✅ 프로덕션 레벨 에러 처리

---

## 📖 학습 로드맵

### 1단계: 스트리밍 기초 개념 이해
**디렉토리**: `step01-streaming-basics/`

**학습 내용**:
- 왜 스트리밍 처리가 필요한가?
- 10GB 파일을 메모리에 올리면 안 되는 이유
- 스트리밍: 데이터를 조금씩 읽고, 처리하고, 버리는 방식
- 전체 데이터를 메모리에 올리지 않고 필요한 부분만 순차 처리

**핵심 개념**:
```
대용량 파일 → 읽기(Read) → 처리(Process) → 출력(Write)
메모리 효율적 처리
```

---

### 2단계: io 패키지 핵심 인터페이스
**디렉토리**: `step02-io-interfaces/`

**학습 내용**:
- `io.Reader` 인터페이스
  - `Read(p []byte) (n int, err error)`
  - 단 하나의 메서드로 모든 입력 처리
- `io.Writer` 인터페이스
  - `Write(p []byte) (n int, err error)`
  - 파일, 네트워크, 버퍼 모두 같은 방식
- EOF(End of File) 처리 패턴
- 버퍼를 통한 데이터 읽기

**실습 과제**:
- 10바이트씩 문자열 읽어보기
- EOF까지 모든 데이터 처리하기
- `strings.Reader`로 스트리밍 체험

---

### 3단계: io 패키지 유용한 함수들
**디렉토리**: `step03-io-functions/`

**학습 내용**:
- `io.Copy(dst, src)` - 32KB 버퍼 자동 사용
- `io.CopyN(dst, src, n)` - 정확한 양만큼 복사
- `io.ReadAll(r)` - 전부 읽기 (⚠️ 대용량 파일 주의)
- `io.MultiReader()` - 여러 Reader를 하나로
- `io.MultiWriter()` - 한 번에 여러 곳에 쓰기

**Pro Tip**:
- `io.Copy`는 `io.WriterTo`나 `io.ReaderFrom`을 구현하면 자동 최적화됨
- 로깅 시 콘솔과 파일에 동시 출력 가능

**실습 과제**:
- 파일 복사 구현
- 여러 파일을 하나로 합치기
- 로그를 여러 곳에 동시 출력

---

### 4단계: 파일 처리 실전 테크닉
**디렉토리**: `step04-file-operations/`

**학습 내용**:
- 파일 열기/닫기
  - `os.Open(filename)` - 읽기 전용
  - `os.Create(filename)` - 생성/덮어쓰기
  - `os.OpenFile(name, flag, perm)` - 세밀한 제어
- `defer file.Close()` - 리소스 관리 (필수!)
- 파일 정보: `file.Stat()`

**os.OpenFile 플래그**:
- `os.O_RDONLY` - 읽기 전용
- `os.O_WRONLY` - 쓰기 전용
- `os.O_RDWR` - 읽기/쓰기
- `os.O_APPEND` - 파일 끝에 추가
- `os.O_CREATE` - 파일 없으면 생성
- `os.O_TRUNC` - 파일 열 때 내용 비우기

**버퍼링된 I/O**:
- `bufio.Scanner` - 줄 단위 읽기 (효율적)
- `bufio.Reader` - 버퍼링된 읽기
- `bufio.Writer` - 버퍼링된 쓰기

**청크 단위 대용량 파일 처리**:
- 고정 크기 버퍼로 읽기
- 메모리 사용량이 파일 크기와 무관
- 10GB 파일을 1MB 메모리로 처리 가능!

**실습 과제**:
- 파일 정보 출력 프로그램
- 1MB 청크로 대용량 파일 처리
- 진행률 표시 기능 추가

---

### 5단계: 스트리밍 고급 기법
**디렉토리**: `step05-advanced-streaming/`

**학습 내용**:

#### io.Pipe를 활용한 동시성 처리
- `io.Pipe()` - Reader와 Writer 연결
- 고루틴과 함께 사용하여 실시간 처리
- 파일 읽으면서 동시에 압축 가능
- 메모리에 전체 파일 올리지 않음

#### 커스텀 Reader/Writer 구현
- `io.Reader` 인터페이스 구현
- `io.Writer` 인터페이스 구현
- 데코레이터 패턴으로 기능 추가

**예시**:
- `UpperCaseReader` - 대문자로 변환
- `LineNumberWriter` - 줄 번호 추가

#### 고급 io 패턴
- `io.LimitReader(r, n)` - DoS 공격 방지 (최대 n바이트만 읽기)
- `io.TeeReader(r, w)` - 읽으면서 동시에 쓰기 (체크섬 계산)
- `io.SectionReader` - 파일의 특정 부분만 읽기

**실습 과제**:
- 파일 읽으면서 동시에 gzip 압축
- 파일 다운로드 중 MD5 체크섬 계산
- 커스텀 데이터 변환 Reader 구현

---

### 6단계: 실전 프로젝트 - 로그 파일 분석기
**디렉토리**: `step06-log-analyzer/`

**프로젝트 목표**:
대용량 로그 파일을 스트리밍 방식으로 분석하는 완전한 프로그램 구현

**주요 기능**:
1. 에러/경고/정보 레벨 통계 수집
2. IP 주소 추출 및 빈도 분석
3. 정규표현식 패턴 매칭
4. 실시간 진행률 표시 (1000줄마다)
5. 분석 결과 리포트 생성

**기술 스택**:
- `bufio.Reader` - 줄 단위 읽기
- `regexp` - 패턴 매칭
- 구조체로 통계 관리
- 스트리밍으로 메모리 효율성 유지

**핵심 구현**:
```
LogAnalyzer 구조체
├── LogStats (통계 데이터)
│   ├── TotalLines
│   ├── ErrorCount
│   ├── WarningCount
│   ├── UniqueIPs (map)
│   └── ErrorMessages
└── 정규표현식 패턴들
```

**실습 과제**:
- 로그 파일 분석 프로그램 완성
- 1GB 로그 파일로 테스트
- HTML/JSON 리포트 생성

---

### 7단계: 성능 최적화
**디렉토리**: `step07-performance/`

**학습 내용**:

#### 버퍼 크기 최적화
- 버퍼가 너무 작으면: 시스템 콜 많음 → 느림 🐌
- 버퍼가 너무 크면: 메모리 낭비 ⚠️
- **권장 크기: 32KB ~ 64KB** 🚀
- 상황에 따라 벤치마크로 최적값 찾기

**성능 비교**:
```
1KB 버퍼:   느림
4KB 버퍼:   보통
32KB 버퍼:  최적
64KB 버퍼:  최적
1MB 버퍼:   메모리 낭비
```

#### 병렬 처리로 속도 높이기
- 워커 풀 패턴
- 작업 채널 + 결과 채널
- `sync.WaitGroup`로 대기
- CPU 코어 최대 활용

**성능 향상**:
```
순차 처리 (1 워커): 40초
병렬 처리 (4 워커): 12초
병렬 처리 (8 워커): 8초
```

#### 메모리 풀 활용
- `sync.Pool`로 버퍼 재사용
- GC 압력 감소
- 메모리 할당 최소화

**효과**:
```
풀 없이: 할당 10,000회, GC 50회
풀 사용: 할당 100회, GC 5회
```

**실습 과제**:
- 다양한 버퍼 크기 성능 벤치마크
- 여러 파일 병렬 압축 구현
- sync.Pool 적용 전/후 비교

---

### 8단계: 에러 처리와 안전성
**디렉토리**: `step08-error-handling/`

**학습 내용**:

#### defer를 활용한 리소스 정리
```go
defer file.Close()  // 함수 종료 시 자동 닫기
```

#### 롤백 메커니즘
- 에러 발생 시 불완전한 파일 삭제
- 트랜잭션 패턴
- `file.Sync()` - 디스크에 확실히 쓰기

#### 컨텍스트를 활용한 타임아웃
- `context.WithTimeout()` - 타임아웃 설정
- `context.WithCancel()` - 수동 취소
- 네트워크/느린 I/O 작업에 필수

#### 에러 래핑 (Go 1.13+)
- `fmt.Errorf("설명: %w", err)` - 에러 래핑
- `errors.Is(err, target)` - 에러 타입 확인
- `errors.As(err, &target)` - 에러 타입 변환
- 커스텀 에러 타입 구현

**실습 과제**:
- 안전한 파일 복사 함수 구현
- 타임아웃이 있는 파일 읽기
- 커스텀 에러 타입으로 상세 정보 제공

---

### 9단계: HTTP 스트리밍
**디렉토리**: `step09-http-streaming/`

**학습 내용**:

#### 파일 다운로드 핸들러
- 스트리밍 방식으로 전송
- `io.Copy(w, file)` - 메모리에 올리지 않음
- Content-Disposition, Content-Length 헤더 설정

#### Range 요청 지원 (이어받기)
- HTTP Range 헤더 파싱
- `file.Seek(start, 0)` - 파일 포인터 이동
- `io.CopyN(w, file, length)` - 부분 전송
- 206 Partial Content 응답

#### 파일 업로드 핸들러
- 멀티파트 폼 파싱 (10MB 메모리 제한)
- `r.FormFile("file")` - 파일 가져오기
- 스트리밍 방식으로 저장

**실습 과제**:
- HTTP 파일 서버 구현
- Range 요청으로 이어받기 지원
- 진행률 추적 API 추가

---

### 10단계: 테스트와 벤치마킹
**디렉토리**: `step10-testing-benchmark/`

**학습 내용**:

#### 단위 테스트 작성
```bash
go test ./...
```
- 테이블 기반 테스트
- 다양한 케이스 검증
- 에러 처리 테스트

#### 벤치마크 작성
```bash
go test -bench=. -benchmem
```
- `func BenchmarkXxx(b *testing.B)`
- `b.ResetTimer()` - 초기화 시간 제외
- `b.SetBytes(n)` - 처리량 측정

#### 프로파일링
```bash
# CPU 프로파일
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# 메모리 프로파일
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

**실습 과제**:
- 다양한 복사 방법 벤치마크
- 버퍼 크기별 성능 측정
- 메모리 풀 효과 검증

---

### 11단계: 고급 패턴과 베스트 프랙티스
**디렉토리**: `step11-advanced-patterns/`

**학습 내용**:

#### 파이프라인 패턴
```
파일 읽기 → 압축 → 체크섬 계산 → 파일 쓰기
```
- 각 단계를 독립적으로 구현
- `io.Pipe()`로 단계 연결
- 모든 단계가 동시에 실행
- 재사용 가능한 컴포넌트

#### 어댑터 패턴
- `ProgressReader` - 진행률 추적
- `ThrottledReader` - 속도 제한 (1MB/s)
- 여러 Reader/Writer 조합 가능

#### 베스트 프랙티스 체크리스트
- ✅ 항상 `defer`로 리소스 정리
- ✅ 적절한 버퍼 크기 사용 (32-64KB)
- ✅ 에러를 무시하지 않기
- ✅ 컨텍스트로 타임아웃 설정
- ✅ 대용량은 스트리밍으로
- ✅ 병렬 처리 활용
- ✅ 테스트와 벤치마크 작성
- ✅ 인터페이스 활용

**실습 과제**:
- 다단계 파이프라인 구현
- 조합 가능한 Reader/Writer 세트 만들기
- 실무 프로젝트에 패턴 적용

---

## 📁 프로젝트 디렉토리 구조

```
file-streaming/
├── README.md                        # 이 파일
├── go.mod                          # Go 모듈 파일
├── intro.txt                       # 원본 학습 자료
├── 01_대용량처리.png               # 참고 다이어그램
├── 02_버퍼링_vs_논버퍼링.png       # 참고 다이어그램
├── 03_성능최적화전략.png           # 참고 다이어그램
│
├── step01-streaming-basics/        # 1단계: 스트리밍 기초
│   └── README.md
│
├── step02-io-interfaces/           # 2단계: io 인터페이스
│   └── README.md
│
├── step03-io-functions/            # 3단계: io 함수들
│   └── README.md
│
├── step04-file-operations/         # 4단계: 파일 작업
│   └── README.md
│
├── step05-advanced-streaming/      # 5단계: 고급 스트리밍
│   └── README.md
│
├── step06-log-analyzer/            # 6단계: 로그 분석기
│   └── README.md
│
├── step07-performance/             # 7단계: 성능 최적화
│   └── README.md
│
├── step08-error-handling/          # 8단계: 에러 처리
│   └── README.md
│
├── step09-http-streaming/          # 9단계: HTTP 스트리밍
│   └── README.md
│
├── step10-testing-benchmark/       # 10단계: 테스트/벤치마크
│   └── README.md
│
└── step11-advanced-patterns/       # 11단계: 고급 패턴
    └── README.md
```

---

## 🎯 학습 진행 방법

### 1. 순차적 학습
- Step 1부터 순서대로 진행
- 각 단계의 README.md를 읽고 이해
- 개념을 완전히 이해한 후 다음 단계로

### 2. 실습 중심
- 각 단계마다 제시된 실습 과제 수행
- 직접 코드를 작성하며 체득
- 다양한 변형 시도

### 3. 참고 자료 활용
- intro.txt의 예제 코드 참고
- 다이어그램 이미지로 시각적 이해
- 공식 Go 문서 병행 학습

### 4. 벤치마크로 검증
- 성능 개선 효과를 숫자로 확인
- 다양한 버퍼 크기 테스트
- 최적화 전/후 비교

---

## 📊 핵심 개념 요약

### 필수 암기 사항

#### io.Reader 인터페이스
```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

#### io.Writer 인터페이스
```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

#### 버퍼링 vs 논버퍼링
- **논버퍼링**: 시스템 콜 많음 → 느림 🐌
- **버퍼링**: 시스템 콜 최소화 → 빠름 🚀

#### 권장 버퍼 크기
- 일반 파일: **32KB ~ 64KB**
- 네트워크: 4KB ~ 8KB
- SSD: 64KB ~ 128KB

#### 메모리 사용량
- 전체 로드: 10GB 파일 = 10GB 메모리 ❌
- 청크 처리: 10GB 파일 = 1MB 메모리 ✅

---

## 🚀 학습 완료 후

### 다음 단계
1. **실전 프로젝트 적용**
   - 로그 처리 시스템
   - 파일 전송 서비스
   - 데이터 ETL 파이프라인

2. **심화 학습**
   - 네트워크 프로그래밍 (TCP/HTTP)
   - 분산 시스템
   - 데이터베이스 스트리밍

3. **오픈소스 기여**
   - Go 프로젝트 분석
   - 버그 수정 및 기능 추가

---

## 📚 참고 자료

### 공식 문서
- [Go io 패키지](https://pkg.go.dev/io)
- [Go bufio 패키지](https://pkg.go.dev/bufio)
- [Go os 패키지](https://pkg.go.dev/os)

### 학습 자료
- `intro.txt` - 전체 학습 내용
- `01_대용량처리.png` - 대용량 파일 처리 플로우
- `02_버퍼링_vs_논버퍼링.png` - 성능 비교
- `03_성능최적화전략.png` - 최적화 전략

---

## 🌟 마무리

이 로드맵을 완료하면:
- ✅ Go 파일 스트리밍 완전 마스터
- ✅ 대용량 데이터 처리 능력
- ✅ 프로덕션 레벨 코드 작성
- ✅ 성능 최적화 기법 습득
- ✅ 실무 프로젝트 포트폴리오

**즐거운 Go 프로그래밍 되세요!** 🚀✨

---

*Last Updated: 2026-02-04*

