package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// 파일 다운로드 핸들러
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "파일명이 필요합니다", http.StatusBadRequest)
		return
	}

	// 파일 열기
	safeFilename := filepath.Base(filename) // " ../../etc/passwd" -> "passwd"로 변경됨
	file, err := os.Open("./uploads/" + safeFilename)

	if err != nil {
		http.Error(w, "파일을 찾을 수 없습니다", http.StatusNotFound)
		return
	}
	defer file.Close()

	// 파일 정보 가져오기
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "파일 정보를 가져올 수 없습니다", http.StatusInternalServerError)
		return
	}

	// 헤더 설정
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// 스트리밍 전송
	written, err := io.Copy(w, file)
	if err != nil {
		log.Printf("전송 중 에러: %v\n", err)
		return
	}

	log.Printf("%s 파일 전송 완료: %d 바이트\n", filename, written)
}

// Range 요청을 지원하는 핸들러 (이어받기 지원)
func rangeDownloadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "파일명이 필요합니다", http.StatusBadRequest)
		return
	}

	// 파일 열기
	safeFilename := filepath.Base(filename) // " ../../etc/passwd" -> "passwd"로 변경됨
	file, err := os.Open("./uploads/" + safeFilename)
	if err != nil {
		http.Error(w, "파일을 찾을 수 없습니다", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "파일 정보를 가져올 수 없습니다", http.StatusInternalServerError)
		return
	}

	fmt.Println("fileInfo : ", fileInfo)
	// Content-Disposition 설정 (다운로드 창이 뜨게 함)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", safeFilename))

	// http.ServeContent가 Range 헤더를 자동으로 확인하여
	// 전체 전송(200 OK) 또는 부분 전송(206 Partial Content)을 알아서 처리합니다.
	http.ServeContent(w, r, safeFilename, fileInfo.ModTime(), file)

	// // Range 헤더 확인
	// rangeHeader := r.Header.Get("Range")
	// fmt.Println("rangeHeader :", rangeHeader)
	// if rangeHeader == "" {
	// 	// 전체 파일 전송
	// 	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	// 	w.Header().Set("Content-Type", "application/octet-stream")
	// 	io.Copy(w, file)
	// 	return
	// }

	// // Range 요청 처리 (간단한 구현)
	// // 실제로는 더 복잡한 파싱이 필요해
	// var start, end int64
	// fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)

	// fmt.Println("start : ", start, " end : ", end)

	// if end == 0 || end >= fileInfo.Size() {
	// 	end = fileInfo.Size() - 1
	// }

	// // 파일 포인터 이동
	// file.Seek(start, 0)

	// // 헤더 설정
	// w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
	// w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	// w.Header().Set("Content-Type", "application/octet-stream")
	// w.WriteHeader(http.StatusPartialContent)

	// // 부분 전송
	// io.CopyN(w, file, end-start+1)
}

// 업로드 핸들러
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST 메서드만 허용됩니다", http.StatusMethodNotAllowed)
		return
	}

	// 멀티파트 폼 파싱 (최대 10MB 메모리 사용)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "폼 파싱 실패", http.StatusBadRequest)
		return
	}

	// 파일 가져오기
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "파일을 가져올 수 없습니다", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 저장할 파일 생성
	dst, err := os.Create("uploads/" + header.Filename)
	if err != nil {
		http.Error(w, "파일 생성 실패", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 스트리밍 방식으로 저장
	written, err := io.Copy(dst, file)
	if err != nil {
		http.Error(w, "파일 저장 실패", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "파일 업로드 성공: %s (%d 바이트)\n", header.Filename, written)
	log.Printf("파일 업로드: %s (%d 바이트)\n", header.Filename, written)
}

// 진행률을 보여주는 업로드 핸들러
type ProgressReader struct {
	reader     io.Reader
	total      int64
	current    int64
	onProgress func(current, total int64)
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.current += int64(n)

	if pr.onProgress != nil {
		pr.onProgress(pr.current, pr.total)
	}

	return n, err
}

func main() {
	// uploads 디렉토리 생성
	os.MkdirAll("uploads", 0755)

	// 1. 루트 경로("/") 접속 시 index.html 파일 서빙
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 경로가 정확히 "/" 일 때만 index.html을 보여줌 (안그러면 모든 경로에서 보임)
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html") // index.html 파일 경로
	})

	// 핸들러 등록
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/range-download", rangeDownloadHandler)
	http.HandleFunc("/upload", uploadHandler)

	// 정적 파일 서빙
	http.Handle("/files/", http.StripPrefix("/files", http.FileServer(http.Dir("./uploads"))))

	fmt.Println("서버 시작: http://localhost:8080")
	fmt.Println("다운로드: http://localhost:8080/download?file=example.txt")
	fmt.Println("업로드: http://localhost:8080/upload")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
