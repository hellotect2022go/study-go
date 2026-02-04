# Step 9: HTTP 스트리밍

## 🎯 학습 목표
- HTTP 파일 다운로드 구현
- Range 요청으로 이어받기 지원
- 멀티파트 파일 업로드 처리
- 스트리밍 방식으로 메모리 절약

## 🌐 HTTP 스트리밍이란?

### 전통적 방식 (비효율)
```
파일 → 메모리 전체 로드 → HTTP 응답
```
- 10GB 파일 = 10GB 메모리 필요
- 메모리 부족 위험
- 첫 바이트까지 지연 시간 길음

### 스트리밍 방식 (효율적)
```
파일 → 청크 읽기 → 즉시 전송 → 다음 청크...
```
- 10GB 파일 = 64KB 메모리만
- 메모리 효율적
- 즉시 다운로드 시작

## 📥 파일 다운로드 핸들러

### 기본 구조

```
클라이언트 요청
  ↓
파일 열기
  ↓
헤더 설정
  ↓
io.Copy(response, file) ← 스트리밍!
  ↓
자동 전송 완료
```

### 필수 헤더

| 헤더 | 용도 | 예시 |
|------|------|------|
| Content-Disposition | 다운로드 파일명 | `attachment; filename="data.pdf"` |
| Content-Type | MIME 타입 | `application/octet-stream` |
| Content-Length | 파일 크기 | `1048576` (바이트) |

### 핵심 코드 패턴

```go
// 1. 파일 열기
file, err := os.Open(filename)
if err != nil {
    http.Error(w, "파일 없음", http.StatusNotFound)
    return
}
defer file.Close()

// 2. 파일 정보
fileInfo, _ := file.Stat()

// 3. 헤더 설정
w.Header().Set("Content-Disposition", 
    fmt.Sprintf("attachment; filename=%s", filename))
w.Header().Set("Content-Type", "application/octet-stream")
w.Header().Set("Content-Length", 
    strconv.FormatInt(fileInfo.Size(), 10))

// 4. 스트리밍 전송
io.Copy(w, file)  // ← 핵심!
```

## 🔄 Range 요청 - 이어받기 지원

### HTTP Range란?

클라이언트가 파일의 **일부분만** 요청하는 기능

### 사용 사례
- 다운로드 이어받기
- 동영상 탐색 (스킵)
- 대용량 파일 분할 다운로드

### Range 헤더 형식

| 요청 | 의미 |
|------|------|
| `Range: bytes=0-499` | 처음 500바이트 |
| `Range: bytes=500-999` | 500~999 바이트 |
| `Range: bytes=500-` | 500부터 끝까지 |
| `Range: bytes=-500` | 마지막 500바이트 |

### 구현 흐름

```
1. Range 헤더 확인
   ↓
2. 범위 파싱 (start, end)
   ↓
3. file.Seek(start, 0)
   ↓
4. Content-Range 헤더 설정
   ↓
5. 206 Partial Content 응답
   ↓
6. io.CopyN(w, file, length)
```

### 응답 헤더

```
HTTP/1.1 206 Partial Content
Content-Range: bytes 0-499/1000
Content-Length: 500
```

## 📤 파일 업로드 처리

### 멀티파트 폼

#### 클라이언트 요청 형식
```
POST /upload HTTP/1.1
Content-Type: multipart/form-data; boundary=----...

------...
Content-Disposition: form-data; name="file"; filename="photo.jpg"
Content-Type: image/jpeg

[파일 데이터]
------...--
```

### 서버 처리 흐름

```
1. ParseMultipartForm(maxMemory)
   ↓
2. FormFile("file")로 파일 가져오기
   ↓
3. 검증 (크기, 확장자)
   ↓
4. 저장 위치 파일 생성
   ↓
5. io.Copy(dstFile, uploadedFile)
   ↓
6. 성공 응답
```

### 메모리 제한

```go
// 최대 10MB만 메모리에 올림
r.ParseMultipartForm(10 << 20)  // 10MB

// 나머지는 임시 파일로 저장됨
```

## 🛡️ 보안 고려사항

### 1. 파일 크기 제한

#### ❌ 위험한 코드
```go
// 무제한 업로드 허용!
io.Copy(dst, src)
```

#### ✅ 안전한 코드
```go
// 10MB 제한
limited := io.LimitReader(src, 10*1024*1024)
io.Copy(dst, limited)
```

### 2. 파일명 검증

#### 공격 예시
```
../../../etc/passwd  (디렉토리 순회)
<script>alert(1)</script>.jpg  (XSS)
```

#### 방어
```go
// 파일명 정제
filename = filepath.Base(filename)
filename = strings.ReplaceAll(filename, "..", "")

// UUID로 저장
safeFilename = uuid.New().String() + filepath.Ext(filename)
```

### 3. MIME 타입 검증

#### 확장자만 확인 (취약)
```go
ext := filepath.Ext(filename)
if ext != ".jpg" { ... }  // ❌ 우회 가능
```

#### 실제 내용 확인 (안전)
```go
buffer := make([]byte, 512)
file.Read(buffer)
mimeType := http.DetectContentType(buffer)

if mimeType != "image/jpeg" { ... }  // ✅ 안전
file.Seek(0, 0)  // 포인터 처음으로
```

### 4. 저장 위치 제한

```go
const uploadDir = "./uploads"

// 경로 벗어남 방지
safePath := filepath.Join(uploadDir, safeFilename)
if !strings.HasPrefix(safePath, uploadDir) {
    return errors.New("잘못된 경로")
}
```

## 📊 진행률 추적

### ProgressReader 패턴

```go
type ProgressReader struct {
    reader   io.Reader
    total    int64
    current  int64
    callback func(current, total int64)
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

### 사용 예시

```go
progress := &ProgressReader{
    reader: file,
    total:  fileSize,
    callback: func(current, total int64) {
        percent := float64(current) / float64(total) * 100
        fmt.Printf("\r진행률: %.2f%%", percent)
    },
}

io.Copy(w, progress)
```

## 🎓 실습 과제

### 과제 1: 기본 파일 서버

**요구사항**:
1. `/download?file=xxx` 엔드포인트
2. 파일 스트리밍 다운로드
3. 적절한 헤더 설정
4. 에러 처리

**테스트**:
```bash
curl -O http://localhost:8080/download?file=test.pdf
```

### 과제 2: Range 요청 지원

**요구사항**:
1. Range 헤더 파싱
2. 206 Partial Content 응답
3. Content-Range 헤더
4. 이어받기 테스트

**테스트**:
```bash
curl -H "Range: bytes=0-99" http://localhost:8080/download?file=test.pdf
```

### 과제 3: 안전한 업로드

**요구사항**:
1. 멀티파트 업로드 처리
2. 10MB 크기 제한
3. 파일명 검증
4. MIME 타입 확인
5. 안전한 저장

**보안 체크리스트**:
- [ ] 크기 제한 (`io.LimitReader`)
- [ ] 파일명 정제
- [ ] 확장자 화이트리스트
- [ ] 저장 경로 검증

### 과제 4: 진행률 표시 API

**요구사항**:
1. 업로드 진행률 WebSocket으로 전송
2. 다운로드 진행률 SSE로 전송
3. 백분율 계산
4. 전송 속도 표시

## 🚀 고급 기능

### 1. 청크 인코딩 전송

```go
// Transfer-Encoding: chunked
w.Header().Set("Transfer-Encoding", "chunked")

// 실시간 데이터 전송
for chunk := range dataChannel {
    w.Write(chunk)
    if f, ok := w.(http.Flusher); ok {
        f.Flush()  // 즉시 전송
    }
}
```

### 2. 다중 Range 요청

```
Range: bytes=0-99,200-299,400-499
```

복잡하지만 CDN에서 사용

### 3. ETag 캐싱

```go
// 파일 수정 시간으로 ETag 생성
etag := fmt.Sprintf(`"%x"`, fileInfo.ModTime().Unix())
w.Header().Set("ETag", etag)

// If-None-Match 확인
if r.Header.Get("If-None-Match") == etag {
    w.WriteHeader(http.StatusNotModified)
    return
}
```

## 🔑 핵심 요약

### 다운로드
```go
file, _ := os.Open(filename)
defer file.Close()

w.Header().Set("Content-Disposition", "attachment; filename=...")
w.Header().Set("Content-Length", ...)

io.Copy(w, file)  // 스트리밍!
```

### Range 요청
```
1. Range 헤더 확인
2. file.Seek(start, 0)
3. io.CopyN(w, file, length)
4. 206 Partial Content 응답
```

### 업로드
```go
r.ParseMultipartForm(10 << 20)
file, header, _ := r.FormFile("file")
defer file.Close()

// 검증 후 저장
limited := io.LimitReader(file, maxSize)
io.Copy(dst, limited)
```

### 보안
```
✅ 크기 제한 (io.LimitReader)
✅ 파일명 검증
✅ MIME 타입 확인
✅ 경로 검증
```

## ➡️ 다음 단계

**Step 10: 테스트와 벤치마킹**
- 단위 테스트 작성
- 벤치마크로 성능 측정
- 프로파일링

