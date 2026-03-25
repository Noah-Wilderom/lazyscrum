package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Noah-Wilderom/lazyscrum/internal/tracker/jira"

	"github.com/Noah-Wilderom/lazyscrum/internal/config"

	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to an issue tracker",
}

var connectJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Connect to Jira via API token",
	RunE:  runConnectJira,
}

func init() {
	connectCmd.AddCommand(connectJiraCmd)
	rootCmd.AddCommand(connectCmd)
}

func runConnectJira(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your Jira Cloud domain (e.g. mycompany.atlassian.net): ")
	domain, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	domain = strings.TrimSpace(domain)

	fmt.Print("Enter your Jira email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	email = strings.TrimSpace(email)

	fmt.Println()
	fmt.Println("Create an API token at: https://id.atlassian.com/manage-profile/security/api-tokens")
	fmt.Println()
	fmt.Print("Enter your API token: ")
	token, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	token = strings.TrimSpace(token)

	// Verify the credentials work
	client := jira.NewClient(domain, email, token)
	_, err = client.SearchIssues("updated >= -30d ORDER BY updated DESC")
	if err != nil {
		return fmt.Errorf("failed to connect to Jira: %w", err)
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	cfg.Tracker = &config.TrackerConfig{
		Provider: "jira",
		Domain:   domain,
		Email:    email,
		APIToken: token,
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("Successfully connected to Jira (%s)!\n", domain)
	fmt.Println("You can now use the ticket selector in the TUI (press J).")

	return nil
}
