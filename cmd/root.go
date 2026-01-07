package cmd

import (
	"fmt"
	"os"
	"torneira-de-soco-ingles/internal"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skynet",
	Short: "Skynet Offensive Toolkit - Torneira de Soco Ingles",
	Long: `
+===============================================================+
|                TORNEIRA DE SOCO INGLES                        |
|                   Offensive Security Toolkit                  |
+===============================================================+

Uma suite modular de pentest desenvolvida para testes de seguranca
com velocidade e precisao.

Modulos disponiveis:
  scan    - Scanner de portas TCP
  enum    - Enumeracao de diretorios e subdominios
  brute   - Forca bruta em formularios de login
  fuzz    - Fuzzing de parametros e vulnerabilidades
  exploit - Teste de vulnerabilidades conhecidas

Exemplos:
  skynet scan -t 192.168.1.1 --top
  skynet enum -t http://target.com -m dir -w wordlist.txt
  skynet brute -t http://target.com/login -u admin -P passwords.txt
  skynet fuzz -t "http://target.com/page?id=FUZZ" --payload sqli
  skynet exploit -t "http://target.com/page?id=1" --type sqli

  Use apenas em sistemas que voce tem autorizacao para testar!`,
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
