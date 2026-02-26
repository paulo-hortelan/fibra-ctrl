package cmd

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var serverURL string

var rootCmd = &cobra.Command{
	Use:   "fibra-ctrl",
	Short: "Gerenciamento de OLTs via CLI interativo",
	Long:  `fibra-ctrl — CLI interativo para gerenciar OLTs através da API do fibra-ctrl.`,
	RunE:  runMainMenu,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&serverURL,
		"server",
		getEnvOrDefault("FIBRA_CTRL_SERVER", "http://localhost:8080"),
		"URL base do servidor (ou env FIBRA_CTRL_SERVER)",
	)
	rootCmd.AddCommand(oltCmd)
}

func runMainMenu(_ *cobra.Command, _ []string) error {
	for {
		prompt := promptui.Select{
			Label: "fibra-ctrl — Menu principal",
			Items: []string{
				"Gerenciar OLTs",
				"Executar comando em OLT",
				"Sair",
			},
		}

		_, choice, err := prompt.Run()
		if err != nil {
			return nil // Ctrl+C
		}

		switch choice {
		case "Gerenciar OLTs":
			if err := runOLTMenu(); err != nil {
				fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			}
		case "Executar comando em OLT":
			if err := actionExecuteCommand(); err != nil {
				fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			}
		case "Sair":
			fmt.Println("Até logo!")
			return nil
		}
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
