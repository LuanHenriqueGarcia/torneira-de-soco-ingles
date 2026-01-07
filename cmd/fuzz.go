package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"torneira-de-soco-ingles/internal"
	"torneira-de-soco-ingles/internal/payloads"

	"github.com/spf13/cobra"
)

var (
	fuzzTarget      string
	fuzzWordlist    string
	fuzzPayloadType string
	fuzzParam       string
	fuzzMethod      string
	fuzzData        string
	fuzzHeaders     []string
	fuzzThreads     int
	fuzzTimeout     int
	fuzzDelay       int
	fuzzFilterCode  []int
	fuzzFilterSize  int
	fuzzMatchCode   []int
	fuzzMatchStr    string
	fuzzVerbose     bool
	fuzzOutput      string
)

var fuzzCmd = &cobra.Command{
	Use:   "fuzz",
	Short: "Fuzzing de parâmetros e detecção de vulnerabilidades",
	Long: `Módulo de fuzzing para descoberta de diretórios, parâmetros e vulnerabilidades.

O marcador FUZZ será substituído pelos payloads da wordlist.

Exemplos:
  # Fuzzing de diretórios
  skynet fuzz -t http://target.com/FUZZ -w dirs.txt

  # Fuzzing de parâmetros com payloads de SQLi
  skynet fuzz -t "http://target.com/page?id=FUZZ" --payload sqli

  # Fuzzing POST
  skynet fuzz -t http://target.com/login -m POST -d "user=admin&pass=FUZZ" -w passwords.txt

  # Filtrar por código de status
  skynet fuzz -t http://target.com/FUZZ -w dirs.txt --fc 404,403`,
	Run: func(cmd *cobra.Command, args []string) {
		runFuzz()
	},
}

func init() {
	fuzzCmd.Flags().StringVarP(&fuzzTarget, "target", "t", "", "URL alvo com marcador FUZZ (obrigatório)")
	fuzzCmd.Flags().StringVarP(&fuzzWordlist, "wordlist", "w", "", "Wordlist para fuzzing")
	fuzzCmd.Flags().StringVar(&fuzzPayloadType, "payload", "", "Tipo de payload: sqli, xss, lfi, cmdi, etc")
	fuzzCmd.Flags().StringVarP(&fuzzParam, "param", "p", "", "Parâmetro específico para fuzz")
	fuzzCmd.Flags().StringVarP(&fuzzMethod, "method", "m", "GET", "Método HTTP (GET, POST)")
	fuzzCmd.Flags().StringVarP(&fuzzData, "data", "d", "", "Dados POST (use FUZZ como marcador)")
	fuzzCmd.Flags().StringArrayVarP(&fuzzHeaders, "header", "H", nil, "Headers customizados (pode usar múltiplas vezes)")
	fuzzCmd.Flags().IntVar(&fuzzThreads, "threads", 20, "Número de threads")
	fuzzCmd.Flags().IntVar(&fuzzTimeout, "timeout", 10, "Timeout em segundos")
	fuzzCmd.Flags().IntVar(&fuzzDelay, "delay", 0, "Delay entre requisições (ms)")
	fuzzCmd.Flags().IntSliceVar(&fuzzFilterCode, "fc", nil, "Filtrar códigos de status (ex: --fc 404,403)")
	fuzzCmd.Flags().IntVar(&fuzzFilterSize, "fs", -1, "Filtrar por tamanho de resposta")
	fuzzCmd.Flags().IntSliceVar(&fuzzMatchCode, "mc", nil, "Mostrar apenas códigos específicos (ex: --mc 200,301)")
	fuzzCmd.Flags().StringVar(&fuzzMatchStr, "ms", "", "Mostrar apenas respostas contendo string")
	fuzzCmd.Flags().BoolVarP(&fuzzVerbose, "verbose", "v", false, "Modo verboso")
	fuzzCmd.Flags().StringVarP(&fuzzOutput, "output", "o", "", "Arquivo de saída")

	fuzzCmd.MarkFlagRequired("target")

	rootCmd.AddCommand(fuzzCmd)
}

type FuzzResult struct {
	URL        string
	Payload    string
	StatusCode int
	Size       int64
	Duration   time.Duration
}

func runFuzz() {
	internal.PrintBanner()
	internal.PrintInfo("Iniciando módulo de Fuzzing...")

	if !strings.Contains(fuzzTarget, "FUZZ") && !strings.Contains(fuzzData, "FUZZ") {
		internal.PrintError("Marcador FUZZ não encontrado na URL ou dados POST!")
		return
	}

	var wordlist []string
	var err error

	if fuzzWordlist != "" {
		wordlist, err = internal.LoadWordlist(fuzzWordlist)
		if err != nil {
			internal.PrintError("Erro ao carregar wordlist: %v", err)
			return
		}
	} else if fuzzPayloadType != "" {
		wordlist = payloads.GetPayloadsByType(fuzzPayloadType)
		if wordlist == nil {
			internal.PrintError("Tipo de payload inválido! Tipos disponíveis: %v", payloads.GetAvailablePayloadTypes())
			return
		}
	} else {
		internal.PrintError("Especifique uma wordlist (-w) ou tipo de payload (--payload)")
		return
	}

	internal.PrintInfo("Target: %s", fuzzTarget)
	internal.PrintInfo("Método: %s", fuzzMethod)
	internal.PrintInfo("Payloads: %d", len(wordlist))
	internal.PrintInfo("Threads: %d", fuzzThreads)
	if len(fuzzFilterCode) > 0 {
		internal.PrintInfo("Filtrando códigos: %v", fuzzFilterCode)
	}
	if len(fuzzMatchCode) > 0 {
		internal.PrintInfo("Mostrando apenas códigos: %v", fuzzMatchCode)
	}
	fmt.Println()

	client := internal.NewHTTPClient(fuzzTimeout, false, "")

	for _, h := range fuzzHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			client.SetHeader(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	jobs := make(chan string, fuzzThreads*2)
	results := make(chan FuzzResult, fuzzThreads*2)
	var wg sync.WaitGroup
	var tested int32
	total := int32(len(wordlist))

	var foundResults []FuzzResult
	var resultsMutex sync.Mutex

	go func() {
		for result := range results {
			resultsMutex.Lock()
			foundResults = append(foundResults, result)
			resultsMutex.Unlock()
		}
	}()

	startTime := time.Now()

	for i := 0; i < fuzzThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for payload := range jobs {
				if fuzzDelay > 0 {
					time.Sleep(time.Duration(fuzzDelay) * time.Millisecond)
				}

				targetURL := strings.ReplaceAll(fuzzTarget, "FUZZ", url.PathEscape(payload))
				postData := strings.ReplaceAll(fuzzData, "FUZZ", url.QueryEscape(payload))

				var resp *internal.HTTPResponse
				var err error

				if fuzzMethod == "POST" {
					resp, err = client.Post(targetURL, postData)
				} else {
					resp, err = client.Get(targetURL)
				}

				atomic.AddInt32(&tested, 1)
				current := atomic.LoadInt32(&tested)

				if err != nil {
					if fuzzVerbose {
						internal.PrintError("Erro: %s - %v", payload, err)
					}
					continue
				}

				if shouldFilter(resp) {
					if fuzzVerbose {
						fmt.Printf("\r[*] Testando: %s [%d/%d]", truncate(payload, 30), current, total)
					}
					continue
				}

				if !shouldMatch(resp) {
					continue
				}

				result := FuzzResult{
					URL:        targetURL,
					Payload:    payload,
					StatusCode: resp.StatusCode,
					Size:       int64(len(resp.Body)),
					Duration:   resp.Duration,
				}
				results <- result

				statusColor := internal.ColorGreen
				if resp.StatusCode >= 400 {
					statusColor = internal.ColorRed
				} else if resp.StatusCode >= 300 {
					statusColor = internal.ColorYellow
				}

				fmt.Printf("\r%s[%d]%s %-50s [Size: %d] [Time: %v]\n",
					statusColor, resp.StatusCode, internal.ColorReset,
					truncate(payload, 50), len(resp.Body), resp.Duration.Round(time.Millisecond))
			}
		}()
	}

	for _, payload := range wordlist {
		jobs <- payload
	}
	close(jobs)

	wg.Wait()
	close(results)

	elapsed := time.Since(startTime)
	fmt.Println()
	internal.PrintInfo("Tempo total: %v", elapsed)
	internal.PrintInfo("Payloads testados: %d", atomic.LoadInt32(&tested))
	internal.PrintSuccess("Resultados encontrados: %d", len(foundResults))

	if fuzzOutput != "" && len(foundResults) > 0 {
		var lines []string
		for _, r := range foundResults {
			lines = append(lines, fmt.Sprintf("[%d] %s (Size: %d)", r.StatusCode, r.URL, r.Size))
		}
		if err := internal.SaveToFile(fuzzOutput, lines); err != nil {
			internal.PrintError("Erro ao salvar: %v", err)
		} else {
			internal.PrintSuccess("Resultados salvos em: %s", fuzzOutput)
		}
	}
}

func shouldFilter(resp *internal.HTTPResponse) bool {
	for _, code := range fuzzFilterCode {
		if resp.StatusCode == code {
			return true
		}
	}

	if fuzzFilterSize >= 0 && len(resp.Body) == fuzzFilterSize {
		return true
	}

	return false
}

func shouldMatch(resp *internal.HTTPResponse) bool {
	if len(fuzzMatchCode) == 0 && fuzzMatchStr == "" {
		return true
	}

	if len(fuzzMatchCode) > 0 {
		found := false
		for _, code := range fuzzMatchCode {
			if resp.StatusCode == code {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if fuzzMatchStr != "" && !strings.Contains(resp.Body, fuzzMatchStr) {
		return false
	}

	return true
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
