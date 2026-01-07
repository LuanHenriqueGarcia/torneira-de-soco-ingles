package cmd

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
	"torneira-de-soco-ingles/internal"

	"github.com/spf13/cobra"
)

var (
	scanTarget  string
	scanPorts   string
	scanAll     bool
	scanTop     bool
	scanThreads int
	scanTimeout int
	scanOutput  string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scanner de portas TCP",
	Long: `Módulo de scanner de portas TCP.

Exemplos:
  # Scan nas portas comuns
  skynet scan -t 192.168.1.1 --top

  # Scan em todas as portas
  skynet scan -t 192.168.1.1 --all

  # Scan em portas específicas
  skynet scan -t 192.168.1.1 -p 80,443,8080

  # Scan em range de portas
  skynet scan -t 192.168.1.1 -p 1-1000`,
	Run: func(cmd *cobra.Command, args []string) {
		runScan()
	},
}

func init() {
	scanCmd.Flags().StringVarP(&scanTarget, "target", "t", "", "Host alvo (IP ou domínio) (obrigatório)")
	scanCmd.Flags().StringVarP(&scanPorts, "ports", "p", "", "Portas para scan (ex: 80,443 ou 1-1000)")
	scanCmd.Flags().BoolVar(&scanAll, "all", false, "Scan em todas as portas (1-65535)")
	scanCmd.Flags().BoolVar(&scanTop, "top", false, "Scan nas portas mais comuns")
	scanCmd.Flags().IntVar(&scanThreads, "threads", 100, "Número de threads")
	scanCmd.Flags().IntVar(&scanTimeout, "timeout", 2, "Timeout em segundos")
	scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", "Arquivo de saída")

	scanCmd.MarkFlagRequired("target")

	rootCmd.AddCommand(scanCmd)
}

func runScan() {
	internal.PrintBanner()
	internal.PrintInfo("Iniciando módulo de Port Scanner...")

	var ports []int
	if scanAll {
		ports = internal.GetAllPorts()
	} else if scanTop {
		ports = internal.GetCommonPorts()
	} else if scanPorts != "" {
		ports = parsePorts(scanPorts)
	} else {
		ports = internal.GetCommonPorts()
	}

	if len(ports) == 0 {
		internal.PrintError("Nenhuma porta especificada!")
		return
	}

	internal.PrintInfo("Target: %s", scanTarget)
	internal.PrintInfo("Portas: %d", len(ports))
	internal.PrintInfo("Threads: %d", scanThreads)
	internal.PrintInfo("Timeout: %ds", scanTimeout)
	fmt.Println()

	startTime := time.Now()

	jobs := make(chan int, scanThreads*2)
	var openPorts []int
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var scanned int32
	total := int32(len(ports))

	for i := 0; i < scanThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range jobs {
				if scanPort(scanTarget, port, scanTimeout) {
					mutex.Lock()
					openPorts = append(openPorts, port)
					mutex.Unlock()
					internal.PrintSuccess("Porta %d aberta - %s", port, getServiceName(port))
				}
				current := atomic.AddInt32(&scanned, 1)
				if current%100 == 0 || current == total {
					fmt.Printf("\r[*] Progresso: %d/%d (%.1f%%)", current, total, float64(current)/float64(total)*100)
				}
			}
		}()
	}

	for _, port := range ports {
		jobs <- port
	}
	close(jobs)

	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Println()
	fmt.Println()

	sort.Ints(openPorts)

	internal.PrintInfo("===============================================")
	internal.PrintInfo("  RESULTADO DO SCAN")
	internal.PrintInfo("===============================================")
	internal.PrintInfo("Host: %s", scanTarget)
	internal.PrintInfo("Tempo: %v", elapsed)
	internal.PrintInfo("Portas escaneadas: %d", len(ports))
	internal.PrintSuccess("Portas abertas: %d", len(openPorts))
	fmt.Println()

	if len(openPorts) > 0 {
		fmt.Println("PORT      STATE    SERVICE")
		fmt.Println("----      -----    -------")
		for _, port := range openPorts {
			fmt.Printf("%-9d open     %s\n", port, getServiceName(port))
		}
	}

	if scanOutput != "" && len(openPorts) > 0 {
		var lines []string
		lines = append(lines, fmt.Sprintf("# Scan de %s em %s", scanTarget, time.Now().Format("2006-01-02 15:04:05")))
		for _, port := range openPorts {
			lines = append(lines, fmt.Sprintf("%d/tcp open %s", port, getServiceName(port)))
		}
		if err := internal.SaveToFile(scanOutput, lines); err != nil {
			internal.PrintError("Erro ao salvar: %v", err)
		} else {
			internal.PrintSuccess("Resultados salvos em: %s", scanOutput)
		}
	}
}

func scanPort(host string, port int, timeout int) bool {
	ports := internal.PortScanner(host, []int{port}, 1, timeout)
	return len(ports) > 0
}

func parsePorts(portsStr string) []int {
	var ports []int
	
	parts := splitPorts(portsStr)
	
	for _, part := range parts {
		if idx := indexOf(part, '-'); idx != -1 {
			var start, end int
			fmt.Sscanf(part, "%d-%d", &start, &end)
			if start > 0 && end > 0 && start <= end && end <= 65535 {
				for i := start; i <= end; i++ {
					ports = append(ports, i)
				}
			}
		} else {
			var port int
			fmt.Sscanf(part, "%d", &port)
			if port > 0 && port <= 65535 {
				ports = append(ports, port)
			}
		}
	}
	
	return ports
}

func splitPorts(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else if c != ' ' {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func indexOf(s string, c rune) int {
	for i, char := range s {
		if char == c {
			return i
		}
	}
	return -1
}

func getServiceName(port int) string {
	services := map[int]string{
		21:    "ftp",
		22:    "ssh",
		23:    "telnet",
		25:    "smtp",
		53:    "dns",
		80:    "http",
		110:   "pop3",
		111:   "rpcbind",
		135:   "msrpc",
		139:   "netbios-ssn",
		143:   "imap",
		443:   "https",
		445:   "microsoft-ds",
		993:   "imaps",
		995:   "pop3s",
		1433:  "mssql",
		1521:  "oracle",
		1723:  "pptp",
		3306:  "mysql",
		3389:  "rdp",
		5432:  "postgresql",
		5900:  "vnc",
		6379:  "redis",
		8080:  "http-proxy",
		8443:  "https-alt",
		8888:  "http-alt",
		9090:  "zeus-admin",
		27017: "mongodb",
	}
	
	if name, ok := services[port]; ok {
		return name
	}
	return "unknown"
}
