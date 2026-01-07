package internal

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func PrintBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════════════╗
║   ████████╗ ██████╗ ██████╗ ███╗   ██╗███████╗██╗██████╗  █████╗  ║
║   ╚══██╔══╝██╔═══██╗██╔══██╗████╗  ██║██╔════╝██║██╔══██╗██╔══██╗ ║
║      ██║   ██║   ██║██████╔╝██╔██╗ ██║█████╗  ██║██████╔╝███████║ ║
║      ██║   ██║   ██║██╔══██╗██║╚██╗██║██╔══╝  ██║██╔══██╗██╔══██║ ║
║      ██║   ╚██████╔╝██║  ██║██║ ╚████║███████╗██║██║  ██║██║  ██║ ║
║      ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ║
║                   	 DE SOCO INGLÊS                          	║
║              [ Offensive Security Toolkit v1.0 ]              	║
╚═══════════════════════════════════════════════════════════════════╝
`
	fmt.Println(ColorCyan + banner + ColorReset)
}

func PrintSuccess(format string, args ...interface{}) {
	fmt.Printf(ColorGreen+"[+] "+format+ColorReset+"\n", args...)
}

func PrintError(format string, args ...interface{}) {
	fmt.Printf(ColorRed+"[-] "+format+ColorReset+"\n", args...)
}

func PrintInfo(format string, args ...interface{}) {
	fmt.Printf(ColorBlue+"[*] "+format+ColorReset+"\n", args...)
}

func PrintWarning(format string, args ...interface{}) {
	fmt.Printf(ColorYellow+"[!] "+format+ColorReset+"\n", args...)
}

func PrintFound(format string, args ...interface{}) {
	fmt.Printf(ColorPurple+"[✓] "+format+ColorReset+"\n", args...)
}

func LoadWordlist(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir wordlist: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro ao ler wordlist: %w", err)
	}

	return lines, nil
}

func SaveToFile(filepath string, data []string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range data {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func IsValidURL(urlStr string) bool {
	pattern := `^(http|https):\/\/[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(:[0-9]+)?(\/.*)?$`
	matched, _ := regexp.MatchString(pattern, urlStr)
	return matched
}

func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func NormalizeURL(urlStr string) string {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}
	return strings.TrimSuffix(urlStr, "/")
}

func PortScanner(host string, ports []int, threads int, timeout int) []int {
	var openPorts []int
	var mutex sync.Mutex
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, threads)

	for _, port := range ports {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(p int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			address := fmt.Sprintf("%s:%d", host, p)
			conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
			if err == nil {
				conn.Close()
				mutex.Lock()
				openPorts = append(openPorts, p)
				mutex.Unlock()
			}
		}(port)
	}

	wg.Wait()
	return openPorts
}

func GetCommonPorts() []int {
	return []int{
		21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445, 993, 995,
		1723, 3306, 3389, 5432, 5900, 8080, 8443, 8888, 9090, 27017,
	}
}

func GetAllPorts() []int {
	ports := make([]int, 65535)
	for i := 1; i <= 65535; i++ {
		ports[i-1] = i
	}
	return ports
}

func GetPortRange(start, end int) []int {
	if start > end || start < 1 || end > 65535 {
		return nil
	}
	ports := make([]int, end-start+1)
	for i := start; i <= end; i++ {
		ports[i-start] = i
	}
	return ports
}

type WorkerPool struct {
	Workers   int
	Jobs      chan func()
	WaitGroup sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
	pool := &WorkerPool{
		Workers: workers,
		Jobs:    make(chan func(), workers*10),
	}

	for i := 0; i < workers; i++ {
		go pool.worker()
	}

	return pool
}

func (p *WorkerPool) worker() {
	for job := range p.Jobs {
		job()
		p.WaitGroup.Done()
	}
}

func (p *WorkerPool) Submit(job func()) {
	p.WaitGroup.Add(1)
	p.Jobs <- job
}

func (p *WorkerPool) Wait() {
	p.WaitGroup.Wait()
}

func (p *WorkerPool) Close() {
	close(p.Jobs)
}

type ProgressBar struct {
	Total   int
	Current int
	Width   int
	mutex   sync.Mutex
}

func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		Total: total,
		Width: 50,
	}
}

func (p *ProgressBar) Increment() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.Current++
	p.Print()
}

func (p *ProgressBar) Print() {
	percent := float64(p.Current) / float64(p.Total)
	filled := int(percent * float64(p.Width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.Width-filled)
	fmt.Printf("\r[%s] %d/%d (%.1f%%)", bar, p.Current, p.Total, percent*100)

	if p.Current >= p.Total {
		fmt.Println()
	}
}
