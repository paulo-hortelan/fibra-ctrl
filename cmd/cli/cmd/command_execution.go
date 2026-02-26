package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/manifoldco/promptui"
)

type FunctionDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
	Parameters  []string `json:"parameters"`
}

type CommandRequest struct {
	OltID        string                 `json:"oltId"`
	FunctionName string                 `json:"functionName"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

type CommandResponse struct {
	IDs    []string `json:"ids"`
	Status string   `json:"status"`
}

func actionExecuteCommand() error {
	olt, err := pickOLT("Selecione a OLT para executar comando")
	if err != nil {
		return err
	}

	functions, err := fetchFunctions(olt.Vendor)
	if err != nil {
		return err
	}
	if len(functions) == 0 {
		return fmt.Errorf("nenhum comando disponível para o vendor '%s'", olt.Vendor)
	}

	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Name | cyan }} {{ if .Description }}— {{ .Description | faint }}{{ end }}",
		Inactive: "  {{ .Name }} {{ if .Description }}— {{ .Description | faint }}{{ end }}",
		Selected: "  {{ .Name | green }}",
	}

	selector := promptui.Select{
		Label:     "Selecione o comando",
		Items:     functions,
		Templates: templates,
		Size:      10,
	}

	idx, _, err := selector.Run()
	if err != nil {
		return err
	}

	selectedFn := functions[idx]
	params := make(map[string]interface{}, len(selectedFn.Parameters))
	for _, parameter := range selectedFn.Parameters {
		value, promptErr := prompt(fmt.Sprintf("Parâmetro '%s'", parameter), "", validateNotEmpty)
		if promptErr != nil {
			return promptErr
		}
		params[parameter] = value
	}

	resp, err := submitCommand(CommandRequest{
		OltID:        olt.ID,
		FunctionName: selectedFn.Name,
		Parameters:   params,
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n  ✔ Comando enfileirado com sucesso!\n")
	fmt.Printf("  OLT: %s (%s/%s)\n", olt.ID, olt.Vendor, olt.Model)
	fmt.Printf("  Função: %s\n", selectedFn.Name)
	fmt.Printf("  Status: %s\n", resp.Status)
	if len(resp.IDs) > 0 {
		fmt.Printf("  IDs: %v\n", resp.IDs)
	}
	fmt.Println()

	return nil
}

func fetchFunctions(vendor string) ([]FunctionDefinition, error) {
	resp, err := http.Get(serverURL + "/functions?vendor=" + vendor)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var functions []FunctionDefinition
	if err := json.NewDecoder(resp.Body).Decode(&functions); err != nil {
		return nil, fmt.Errorf("erro ao decodificar funções: %w", err)
	}
	return functions, nil
}

func submitCommand(reqData CommandRequest) (*CommandResponse, error) {
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	resp, err := http.Post(serverURL+"/commands", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, apiError(resp)
	}

	var result CommandResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta do comando: %w", err)
	}

	return &result, nil
}
