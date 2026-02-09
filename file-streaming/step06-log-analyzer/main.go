package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

// 1. 버퍼링된 I/O - bufio.Reader로 효율적인 읽기
// 2. 스트리밍 처리 - 한 줄씩 읽어서 메모리 절약
// 3. 정규표현식 - 패턴 매칭으로 로그 분석
// 4. 진행률 표시 - 사용자 경험 개선
// 5. 구조화된 데이터 - 통계를 구조체로 관리
// 6. 파일 쓰기 - 분석 결과를 파일로 저장

// 로그 통계 구조체
type LogStats struct {
	TotalLines    int
	ErrorCount    int
	WarningCount  int
	InfoCount     int
	UniqueIPs     map[string]int
	ErrorMessages []string
}

// 로그 분석기
type LogAnalyzer struct {
	stats        *LogStats
	errorRegex   *regexp.Regexp
	warningRegex *regexp.Regexp
	ipRegex      *regexp.Regexp
}

// 스트리밍 방식으로 로그 파일 분석
func (la *LogAnalyzer) AnalyzerFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Errorf("파일열기 실패 : %w", err)
	}
	defer file.Close()

	// 진행상황 표시를 위한 파일 크기 확인
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	// 버퍼링된 Reader 사용
	reader := bufio.NewReader(file)
	var processedBytes int64

	fmt.Println("로그 파일 분석 시작...")
	startTime := time.Now()

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("읽기 에러: %w", err)
		}

		if len(line) > 0 {
			la.processLine(line)
			processedBytes += int64(len(line))

			// 진행률 표시 매 1000줄마다
			if la.stats.TotalLines%1000 == 0 {
				progress := float64(processedBytes) / float64(fileSize) * 100
				fmt.Printf("\r진행률: %.2f%% (%d 줄 처리)", progress, la.stats.TotalLines)
			}
		}

		if err == io.EOF {
			break
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n\n분석 완료! 소요 시간: %v\n", elapsed)
	return nil

}

// 한줄씩 처리
func (la *LogAnalyzer) processLine(line string) {
	la.stats.TotalLines++

	// 에러 체크
	if la.errorRegex.MatchString(line) {
		la.stats.ErrorCount++
		// 에러 메시지 저장 (최대 10개)
		if len(la.stats.ErrorMessages) < 10 {
			la.stats.ErrorMessages = append(la.stats.ErrorMessages, strings.TrimSpace(line))
		}
	}

	// 경고 체크
	if la.warningRegex.MatchString(line) {
		la.stats.WarningCount++
	}

	// INFO 체크
	if strings.Contains(line, "INFO") {
		la.stats.InfoCount++
	}

	// IP 주소 추출
	ips := la.ipRegex.FindAllString(line, -1)
	for _, ip := range ips {
		la.stats.UniqueIPs[ip]++
	}
}

// 결과 출력
func (la *LogAnalyzer) PrintReport() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 로그 분석 보고서")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\n총 라인 수: %d\n", la.stats.TotalLines)
	fmt.Printf("에러 수: %d (%.2f%%)\n",
		la.stats.ErrorCount,
		float64(la.stats.ErrorCount)/float64(la.stats.TotalLines)*100)
	fmt.Printf("경고 수: %d (%.2f%%)\n",
		la.stats.WarningCount,
		float64(la.stats.WarningCount)/float64(la.stats.TotalLines)*100)
	fmt.Printf("정보 수: %d (%.2f%%)\n",
		la.stats.InfoCount,
		float64(la.stats.InfoCount)/float64(la.stats.TotalLines)*100)

	fmt.Printf("\n고유 IP 주소 수: %d\n", len(la.stats.UniqueIPs))

	// 가장 많이 나타난 IP 찾기
	if len(la.stats.UniqueIPs) > 0 {
		maxIP := ""
		maxCount := 0
		for ip, count := range la.stats.UniqueIPs {
			if count > maxCount {
				maxIP = ip
				maxCount = count
			}
		}
		fmt.Printf("가장 빈번한 IP: %s (%d회)\n", maxIP, maxCount)
	}

	// 에러 메시지 샘플
	if len(la.stats.ErrorMessages) > 0 {
		fmt.Println("\n최근 에러 메시지 샘플:")
		for i, msg := range la.stats.ErrorMessages {
			fmt.Printf("%d. %s\n", i+1, msg)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

// 결과를 파일로 저장
func (la *LogAnalyzer) SaveReport(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// 보고서 작성
	fmt.Fprintf(writer, "로그 분석 보고서\n")
	fmt.Fprintf(writer, "생성 시간: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(writer, "총 라인 수: %d\n", la.stats.TotalLines)
	fmt.Fprintf(writer, "에러 수: %d\n", la.stats.ErrorCount)
	fmt.Fprintf(writer, "경고 수: %d\n", la.stats.WarningCount)
	fmt.Fprintf(writer, "정보 수: %d\n", la.stats.InfoCount)
	fmt.Fprintf(writer, "\n고유 IP 주소 목록:\n")

	for ip, count := range la.stats.UniqueIPs {
		fmt.Fprintf(writer, "%s: %d회\n", ip, count)
	}

	return nil
}

func NewLogAnalyzer() *LogAnalyzer {
	return &LogAnalyzer{
		stats: &LogStats{
			UniqueIPs:     make(map[string]int),
			ErrorMessages: make([]string, 0),
		},
		errorRegex:   regexp.MustCompile(`ERROR|Error|error`),
		warningRegex: regexp.MustCompile(`WARNING|Warning|warning`),
		ipRegex:      regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("사용법 : go run main.go <로그파일 경로>")
		return
	}

	wh := os.Args[0]
	fmt.Println("wh : ", wh)
	logFile := os.Args[1]

	analyzer := NewLogAnalyzer()

	// 파일 분석
	if err := analyzer.AnalyzerFile(logFile); err != nil {
		fmt.Printf("분석 실패 : %v\n", err)
		return
	}

	// 결과 출력
	analyzer.PrintReport()

	// 결과 저장
	reportFile := "log_analysis_reporter.txt"
	if err := analyzer.SaveReport(reportFile); err != nil {
		fmt.Printf("보고서 저장 실패: %v\n", err)
	} else {
		fmt.Printf("\n보고서가 %s에 저장되었습니다.\n", reportFile)
	}

}
