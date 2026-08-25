package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/joho/godotenv"
	"github.com/mmtaee/ocserv-dashboard/backend/config"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	logging "github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	agentsettings "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/agent_settings"
	"github.com/spf13/cobra"
)

type agentTokenAction func(context.Context, *agentsettings.Usecase, io.Writer) error

func newAgentTokenCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "agent-token",
		Short: "Manage the local agent-node authentication token",
	}
	command.AddCommand(
		newAgentTokenActionCommand("get", "Get the current agent token", getAgentToken),
		newAgentTokenActionCommand("create", "Create the agent token", createAgentToken),
		newAgentTokenActionCommand("renew", "Replace the current agent token", renewAgentToken),
		newAgentTokenActionCommand("remove", "Remove the current agent token", removeAgentToken),
	)
	return command
}

func newAgentTokenActionCommand(use, short string, action agentTokenAction) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runAgentTokenAction(command, action)
		},
	}
}

func runAgentTokenAction(command *cobra.Command, action agentTokenAction) error {
	_ = godotenv.Load()
	config.Init(false, "", 0)
	if err := agentsettings.RequireAgentNode(config.Get().AgentNode); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(command.Context())
	logging.Init(ctx, 128)
	defer func() {
		cancel()
		logging.Wait()
	}()
	if err := database.Connect(); err != nil {
		return err
	}
	defer database.Close()

	usecase := agentsettings.New(repository.NewAgentTokenRepository())
	return action(ctx, usecase, command.OutOrStdout())
}

func getAgentToken(ctx context.Context, usecase *agentsettings.Usecase, output io.Writer) error {
	token, err := usecase.Get(ctx)
	return writeAgentToken(output, token, err)
}

func createAgentToken(ctx context.Context, usecase *agentsettings.Usecase, output io.Writer) error {
	token, err := usecase.Create(ctx)
	return writeAgentToken(output, token, err)
}

func renewAgentToken(ctx context.Context, usecase *agentsettings.Usecase, output io.Writer) error {
	token, err := usecase.Renew(ctx)
	return writeAgentToken(output, token, err)
}

func removeAgentToken(ctx context.Context, usecase *agentsettings.Usecase, output io.Writer) error {
	if err := usecase.Remove(ctx); err != nil {
		return err
	}
	_, err := fmt.Fprintln(output, "agent token removed")
	return err
}

func writeAgentToken(output io.Writer, token *models.AgentToken, err error) error {
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(token)
}

func init() {
	rootCmd.AddCommand(newAgentTokenCommand())
}
