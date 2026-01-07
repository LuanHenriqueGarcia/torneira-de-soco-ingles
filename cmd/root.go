package cmd

import (
	"fmt"
	"os"
	"torneira-de-soco-ingles/internal"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skynet",
	Short: "Skynet Offensive Toolkit - Torneira de Soco Inglês",
	Long: `
╔═══════════════════════════════════════════════════════════════╗
║                TORNEIRA DE SOCO INGLÊS                        ║
║                   Offensive Security Toolkit                  ║
╚═══════════════════════════════════════════════════════════════╝

Uma suíte modular de pentest desenvolvida para testes de segurança
com velocidade e precisão.

Módulos disponíveis:
  scan    - Scanner de portas TCP
  enum    - Enumeração de diretórios e subdomínios
  brute   - Força bruta em formulários de login
  fuzz    - Fuzzing de parâmetros e vulnerabilidades
  exploit - Teste de vulnerabilidades conhecidas

Exemplos:
  skynet scan -t 192.168.1.1 --top
  skynet enum -t http://target.com -m dir -w wordlist.txt
  skynet brute -t http://target.com/login -u admin -P passwords.txt
  skynet fuzz -t "http://target.com/page?id=FUZZ" --payload sqli
  skynet exploit -t "http://target.com/page?id=1" --type sqli

  Use apenas em sistemas que você tem autorização para testar!`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.PrintBanner()
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
