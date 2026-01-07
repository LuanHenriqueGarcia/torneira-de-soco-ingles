package cmd

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"torneira-de-soco-ingles/internal"

	"github.com/spf13/cobra"
)

var (
	enumTarget    string
	enumMode      string
	enumWordlist  string
	enumThreads   int
	enumTimeout   int
	enumOutput    string
	enumRecursive bool
	enumExtension string
)

var enumCmd = &cobra.Command{
	Use:   "enum",
	Short: "Enumeração de diretórios e subdomínios",
	Long: `Módulo de enumeração para descoberta de diretórios e subdomínios.

Modos disponíveis:
  - dir: Enumeração de diretórios
  - sub: Enumeração de subdomínios
  - vhost: Enumeração de virtual hosts

Exemplos:
  # Enumerar diretórios
  skynet enum -t http://target.com -m dir -w dirs.txt

  # Enumerar subdomínios
  skynet enum -t target.com -m sub -w subdomains.txt

  # Enumerar diretórios com extensões
  skynet enum -t http://target.com -m dir -w dirs.txt -x php,html,txt`,
	Run: func(cmd *cobra.Command, args []string) {
		runEnum()
	},
}

func init() {
	enumCmd.Flags().StringVarP(&enumTarget, "target", "t", "", "Alvo (obrigatório)")
	enumCmd.Flags().StringVarP(&enumMode, "mode", "m", "dir", "Modo: dir, sub, vhost")
	enumCmd.Flags().StringVarP(&enumWordlist, "wordlist", "w", "", "Wordlist (obrigatório)")
	enumCmd.Flags().IntVar(&enumThreads, "threads", 30, "Número de threads")
	enumCmd.Flags().IntVar(&enumTimeout, "timeout", 10, "Timeout em segundos")
	enumCmd.Flags().StringVarP(&enumOutput, "output", "o", "", "Arquivo de saída")
	enumCmd.Flags().BoolVarP(&enumRecursive, "recursive", "r", false, "Enumeração recursiva")
	enumCmd.Flags().StringVarP(&enumExtension, "extension", "x", "", "Extensões para testar (ex: php,html,txt)")

	enumCmd.MarkFlagRequired("target")
	enumCmd.MarkFlagRequired("wordlist")

	rootCmd.AddCommand(enumCmd)
}

type EnumResult struct {
	URL        string
	StatusCode int
	Size       int64
	Redirect   string
}

func runEnum() {
	internal.PrintBanner()

	switch enumMode {
	case "dir":
		runDirEnum()
	case "sub":
		runSubEnum()
	case "vhost":
		runVhostEnum()
	default:
		internal.PrintError("Modo inválido! Use: dir, sub ou vhost")
	}
}

func runDirEnum() {
	internal.PrintInfo("Iniciando enumeração de diretórios...")

	wordlist, err := internal.LoadWordlist(enumWordlist)
	if err != nil {
		internal.PrintError("Erro ao carregar wordlist: %v", err)
		return
	}

	var extensions []string
	if enumExtension != "" {
		for _, ext := range strings.Split(enumExtension, ",") {
			ext = strings.TrimSpace(ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			extensions = append(extensions, ext)
		}
	}

	var urls []string
	baseURL := internal.NormalizeURL(enumTarget)

	for _, word := range wordlist {
		urls = append(urls, fmt.Sprintf("%s/%s", baseURL, word))
		for _, ext := range extensions {
			urls = append(urls, fmt.Sprintf("%s/%s%s", baseURL, word, ext))
		}
	}

	internal.PrintInfo("Target: %s", baseURL)
	internal.PrintInfo("Wordlist: %d palavras", len(wordlist))
	internal.PrintInfo("URLs a testar: %d", len(urls))
	internal.PrintInfo("Threads: %d", enumThreads)
	if len(extensions) > 0 {
		internal.PrintInfo("Extensões: %v", extensions)
	}
	fmt.Println()

	client := internal.NewHTTPClient(enumTimeout, false, "")

	jobs := make(chan string, enumThreads*2)
	var results []EnumResult
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var tested int32
	_ = tested
	startTime := time.Now()

	for i := 0; i < enumThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range jobs {
				resp, err := client.Get(url)
				atomic.AddInt32(&tested, 1)

				if err != nil {
					continue
				}

				if resp.StatusCode == 404 {
					continue
				}

				result := EnumResult{
					URL:        url,
					StatusCode: resp.StatusCode,
					Size:       int64(len(resp.Body)),
				}

				if resp.StatusCode >= 300 && resp.StatusCode < 400 {
					if loc := resp.Headers.Get("Location"); loc != "" {
						result.Redirect = loc
					}
				}

				mutex.Lock()
				results = append(results, result)
				mutex.Unlock()

				statusColor := internal.ColorGreen
				if resp.StatusCode >= 400 {
					statusColor = internal.ColorRed
				} else if resp.StatusCode >= 300 {
					statusColor = internal.ColorYellow
				}

				redirectInfo := ""
				if result.Redirect != "" {
					redirectInfo = fmt.Sprintf(" -> %s", result.Redirect)
				}

				fmt.Printf("%s[%d]%s %-60s [Size: %d]%s\n",
					statusColor, resp.StatusCode, internal.ColorReset,
					url, len(resp.Body), redirectInfo)
			}
		}()
	}

	for _, url := range urls {
		jobs <- url
	}
	close(jobs)

	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Println()
	internal.PrintInfo("Tempo total: %v", elapsed)
	internal.PrintSuccess("Diretórios/arquivos encontrados: %d", len(results))

	if enumOutput != "" && len(results) > 0 {
		var lines []string
		for _, r := range results {
			lines = append(lines, fmt.Sprintf("[%d] %s (Size: %d)", r.StatusCode, r.URL, r.Size))
		}
		if err := internal.SaveToFile(enumOutput, lines); err != nil {
			internal.PrintError("Erro ao salvar: %v", err)
		} else {
			internal.PrintSuccess("Resultados salvos em: %s", enumOutput)
		}
	}
}

func runSubEnum() {
	internal.PrintInfo("Iniciando enumeração de subdomínios...")

	wordlist, err := internal.LoadWordlist(enumWordlist)
	if err != nil {
		internal.PrintError("Erro ao carregar wordlist: %v", err)
		return
	}

	domain := strings.TrimPrefix(enumTarget, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimSuffix(domain, "/")

	internal.PrintInfo("Domínio: %s", domain)
	internal.PrintInfo("Wordlist: %d palavras", len(wordlist))
	internal.PrintInfo("Threads: %d", enumThreads)
	fmt.Println()

	client := internal.NewHTTPClient(enumTimeout, false, "")

	jobs := make(chan string, enumThreads*2)
	var results []EnumResult
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var tested int32

	startTime := time.Now()

	for i := 0; i < enumThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sub := range jobs {
				subdomain := fmt.Sprintf("%s.%s", sub, domain)
				url := fmt.Sprintf("http://%s", subdomain)

				resp, err := client.Get(url)
				_ = atomic.AddInt32(&tested, 1)

				if err != nil {
					url = fmt.Sprintf("https://%s", subdomain)
					resp, err = client.Get(url)
					if err != nil {
						continue
					}
				}

				result := EnumResult{
					URL:        url,
					StatusCode: resp.StatusCode,
					Size:       int64(len(resp.Body)),
				}

				mutex.Lock()
				results = append(results, result)
				mutex.Unlock()

				internal.PrintFound("%-40s [%d] [Size: %d]", subdomain, resp.StatusCode, len(resp.Body))
			}
		}()
	}

	for _, sub := range wordlist {
		jobs <- sub
	}
	close(jobs)

	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Println()
	internal.PrintInfo("Tempo total: %v", elapsed)
	internal.PrintSuccess("Subdomínios encontrados: %d", len(results))

	if enumOutput != "" && len(results) > 0 {
		var lines []string
		for _, r := range results {
			lines = append(lines, r.URL)
		}
		if err := internal.SaveToFile(enumOutput, lines); err != nil {
			internal.PrintError("Erro ao salvar: %v", err)
		} else {
			internal.PrintSuccess("Resultados salvos em: %s", enumOutput)
		}
	}
}

func runVhostEnum() {
	internal.PrintInfo("Iniciando enumeração de Virtual Hosts...")

	wordlist, err := internal.LoadWordlist(enumWordlist)
	if err != nil {
		internal.PrintError("Erro ao carregar wordlist: %v", err)
		return
	}

	baseURL := internal.NormalizeURL(enumTarget)

	internal.PrintInfo("Target: %s", baseURL)
	internal.PrintInfo("Wordlist: %d palavras", len(wordlist))
	internal.PrintInfo("Threads: %d", enumThreads)
	fmt.Println()

	client := internal.NewHTTPClient(enumTimeout, false, "")
	baseResp, err := client.Get(baseURL)
	if err != nil {
		internal.PrintError("Erro ao conectar ao alvo: %v", err)
		return
	}
	baseSize := len(baseResp.Body)
	internal.PrintInfo("Tamanho base da resposta: %d bytes", baseSize)

	jobs := make(chan string, enumThreads*2)
	var results []EnumResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < enumThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localClient := internal.NewHTTPClient(enumTimeout, false, "")

			for vhost := range jobs {
				localClient.SetHeader("Host", vhost)
				resp, err := localClient.Get(baseURL)

				if err != nil {
					continue
				}

				if len(resp.Body) != baseSize {
					result := EnumResult{
						URL:        vhost,
						StatusCode: resp.StatusCode,
						Size:       int64(len(resp.Body)),
					}

					mutex.Lock()
					results = append(results, result)
					mutex.Unlock()

					internal.PrintFound("%-40s [%d] [Size: %d]", vhost, resp.StatusCode, len(resp.Body))
				}
			}
		}()
	}

	for _, vhost := range wordlist {
		jobs <- vhost
	}
	close(jobs)

	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Println()
	internal.PrintInfo("Tempo total: %v", elapsed)
	internal.PrintSuccess("Virtual Hosts encontrados: %d", len(results))

	if enumOutput != "" && len(results) > 0 {
		var lines []string
		for _, r := range results {
			lines = append(lines, fmt.Sprintf("[%d] %s (Size: %d)", r.StatusCode, r.URL, r.Size))
		}
		if err := internal.SaveToFile(enumOutput, lines); err != nil {
			internal.PrintError("Erro ao salvar: %v", err)
		} else {
			internal.PrintSuccess("Resultados salvos em: %s", enumOutput)
		}
	}
}
