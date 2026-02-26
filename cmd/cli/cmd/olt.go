package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// OLTDetails mirrors the domain struct for JSON (de)serialization.
type OLTDetails struct {
	ID       string `json:"id,omitempty"`
	Title    string `json:"title"`
	IP       string `json:"ip"`
	Vendor   string `json:"vendor"`
	Model    string `json:"model"`
	Protocol string `json:"protocol"`
}

func (o OLTDetails) String() string {
	return fmt.Sprintf("%-36s  %-20s  %-15s  %-15s  %-10s  %s", o.ID, o.Title, o.IP, o.Vendor, o.Model, o.Protocol)
}

var oltCmd = &cobra.Command{
	Use:   "olt",
	Short: "Gerenciar OLTs (modo interativo)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOLTMenu()
	},
}

func runOLTMenu() error {
	for {
		prompt := promptui.Select{
			Label: "OLTs — selecione uma ação",
			Items: []string{
				"Listar todas",
				"Buscar por ID",
				"Criar nova",
				"Atualizar",
				"Deletar",
				"← Voltar",
			},
		}

		_, choice, err := prompt.Run()
		if err != nil {
			return nil // Ctrl+C
		}

		var actionErr error
		switch choice {
		case "Listar todas":
			actionErr = actionList()
		case "Buscar por ID":
			actionErr = actionGet()
		case "Criar nova":
			actionErr = actionCreate()
		case "Atualizar":
			actionErr = actionUpdate()
		case "Deletar":
			actionErr = actionDelete()
		case "← Voltar":
			return nil
		}

		if actionErr != nil {
			fmt.Fprintf(os.Stderr, "\n  ✖ %v\n\n", actionErr)
		}
	}
}

/* Actions */
func actionList() error {
	olts, err := fetchOLTs()
	if err != nil {
		return err
	}
	if len(olts) == 0 {
		fmt.Println("\n  Nenhuma OLT cadastrada.")
		return nil
	}
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "  ID\tTÍTULO\tIP\tTIPO\tMODELO\tPROTOCOLO")
	fmt.Fprintln(w, "  ──────────────────────────────────\t────────────────────\t───────────────\t───────────────\t──────────\t─────────")
	for _, o := range olts {
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\t%s\n", o.ID, o.Title, o.IP, o.Vendor, o.Model, o.Protocol)
	}
	w.Flush()
	fmt.Println()
	return nil
}

func actionGet() error {
	olt, err := pickOLT("Selecione a OLT para ver detalhes")
	if err != nil {
		return err
	}
	fmt.Printf("\n  ID:        %s\n  Título:    %s\n  IP:        %s\n  Tipo:      %s\n  Modelo:    %s\n  Protocolo: %s\n\n",
		olt.ID, olt.Title, olt.IP, olt.Vendor, olt.Model, olt.Protocol)
	return nil
}

func actionCreate() error {
	title, err := prompt("Título da OLT", "", validateNotEmpty)
	if err != nil {
		return err
	}

	ip, err := prompt("IP da OLT", "", validateNotEmpty)
	if err != nil {
		return err
	}
	oltType, model, err := pickTypeAndModel("", "")
	if err != nil {
		return err
	}
	protocol, err := promptSelect("Protocolo de conexão", []string{"ssh", "telnet"})
	if err != nil {
		return err
	}

	body, _ := json.Marshal(OLTDetails{Title: title, IP: ip, Vendor: oltType, Model: model, Protocol: protocol})
	resp, err := http.Post(serverURL+"/olts", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return apiError(resp)
	}

	var created OLTDetails
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return err
	}
	fmt.Printf("\n  ✔ OLT criada com sucesso! ID: %s\n\n", created.ID)
	return nil
}

func actionUpdate() error {
	olt, err := pickOLT("Selecione a OLT para atualizar")
	if err != nil {
		return err
	}

	fmt.Printf("\n  Editando OLT %s\n\n", olt.ID)

	title, err := promptDefault("Novo título", olt.Title, validateNotEmpty)
	if err != nil {
		return err
	}

	ip, err := promptDefault("Novo IP", olt.IP, validateNotEmpty)
	if err != nil {
		return err
	}
	oltType, model, err := pickTypeAndModel(olt.Vendor, olt.Model)
	if err != nil {
		return err
	}
	protocol, err := promptSelect("Protocolo de conexão", []string{"ssh", "telnet"})
	if err != nil {
		return err
	}

	body, _ := json.Marshal(OLTDetails{Title: title, IP: ip, Vendor: oltType, Model: model, Protocol: protocol})
	req, _ := http.NewRequest(http.MethodPut, serverURL+"/olts/"+olt.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	fmt.Printf("\n  ✔ OLT %s atualizada com sucesso!\n\n", olt.ID)
	return nil
}

func actionDelete() error {
	olt, err := pickOLT("Selecione a OLT para deletar")
	if err != nil {
		return err
	}

	confirm, err := promptSelect(
		fmt.Sprintf("Confirmar deleção de %s (%s)?", olt.ID, olt.IP),
		[]string{"Não", "Sim"},
	)
	if err != nil || confirm == "Não" {
		fmt.Println("\n  Operação cancelada.")
		return nil
	}

	req, _ := http.NewRequest(http.MethodDelete, serverURL+"/olts/"+olt.ID, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	fmt.Printf("\n  ✔ OLT %s removida com sucesso!\n\n", olt.ID)
	return nil
}

/* Prompts and helpers */

// pickOLT fetches the OLT list and lets the user choose with arrow keys.
func pickOLT(label string) (OLTDetails, error) {
	olts, err := fetchOLTs()
	if err != nil {
		return OLTDetails{}, err
	}
	if len(olts) == 0 {
		return OLTDetails{}, fmt.Errorf("nenhuma OLT cadastrada")
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Title | cyan }} — {{ .IP }} ({{ .Vendor }}/{{ .Model }}) — {{ .ID | faint }}",
		Inactive: "  {{ .Title }} — {{ .IP }} ({{ .Vendor }}/{{ .Model }}) — {{ .ID | faint }}",
		Selected: "  {{ .Title | green }} — {{ .IP }} ({{ .Vendor }}/{{ .Model }})",
	}

	p := promptui.Select{
		Label:     label,
		Items:     olts,
		Templates: templates,
		Size:      10,
	}

	i, _, err := p.Run()
	if err != nil {
		return OLTDetails{}, err
	}
	return olts[i], nil
}

func prompt(label, def string, validate promptui.ValidateFunc) (string, error) {
	p := promptui.Prompt{Label: label, Default: def, Validate: validate}
	return p.Run()
}

func promptDefault(label, def string, validate promptui.ValidateFunc) (string, error) {
	p := promptui.Prompt{Label: label, Default: def, Validate: validate, AllowEdit: true}
	return p.Run()
}

func promptSelect(label string, items []string) (string, error) {
	p := promptui.Select{Label: label, Items: items}
	_, result, err := p.Run()
	return result, err
}

func validateNotEmpty(s string) error {
	if s == "" {
		return fmt.Errorf("o campo não pode ser vazio")
	}
	return nil
}

// ─── helpers: HTTP ────────────────────────────────────────────────────────────

// OLTTypeEntry mirrors the response from GET /olts/types.
type OLTTypeEntry struct {
	Vendor string   `json:"vendor"`
	Models []string `json:"models"`
}

// fetchOLTTypes returns the supported type/model combinations from the server.
func fetchOLTTypes() ([]OLTTypeEntry, error) {
	resp, err := http.Get(serverURL + "/olts/types")
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var entries []OLTTypeEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("erro ao decodificar tipos: %w", err)
	}
	return entries, nil
}

// pickTypeAndModel shows select prompts for type then model.
// currentType/currentModel are pre-selected when editing.
func pickTypeAndModel(currentType, currentModel string) (string, string, error) {
	entries, err := fetchOLTTypes()
	if err != nil {
		return "", "", err
	}

	types := make([]string, len(entries))
	for i, e := range entries {
		types[i] = e.Vendor
	}

	// default index for type
	typeIdx := 0
	for i, t := range types {
		if t == currentType {
			typeIdx = i
			break
		}
	}

	p := promptui.Select{Label: "Tipo de OLT", Items: types, CursorPos: typeIdx}
	selectedTypeIdx, _, err := p.Run()
	if err != nil {
		return "", "", err
	}
	selectedType := types[selectedTypeIdx]
	models := entries[selectedTypeIdx].Models

	// default index for model
	modelIdx := 0
	for i, m := range models {
		if m == currentModel {
			modelIdx = i
			break
		}
	}

	m := promptui.Select{Label: "Modelo", Items: models, CursorPos: modelIdx}
	selectedModelIdx, _, err := m.Run()
	if err != nil {
		return "", "", err
	}

	return selectedType, models[selectedModelIdx], nil
}

func fetchOLTs() ([]OLTDetails, error) {
	resp, err := http.Get(serverURL + "/olts")
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}

	var olts []OLTDetails
	if err := json.NewDecoder(resp.Body).Decode(&olts); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}
	return olts, nil
}

func apiError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("erro do servidor (%d): %s", resp.StatusCode, string(body))
}
