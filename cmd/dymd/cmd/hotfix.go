package cmd

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cobra"
)

// InspectCmd dumps app state to JSON.

func HotfixCmd(appExporter types.AppExporter, appCreator types.AppCreator, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hotfix",
		Short: "hotfix db state",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.OutOrStderr())

			serverCtx := server.GetServerContextFromCmd(cmd)

			// use the parsed config to load the block and state store
			blockStore, stateStore, err := loadStateAndBlockStore(serverCtx.Config)
			if err != nil {
				return err
			}
			defer func() {
				_ = blockStore.Close()
				_ = stateStore.Close()
			}()

			// Read the data from the KVStore
			fmt.Println("LOADING STATE")

			state, err := stateStore.Load()
			if err != nil {
				return err
			}
			if state.IsEmpty() {
				return errors.New("no state found")
			}
			fmt.Printf("%+v\n", state)

			state.Validators = state.NextValidators.Copy()

			if err := stateStore.Save(state); err != nil {
				return fmt.Errorf("failed to save rolled back state: %w", err)
			}

			return nil
		},
	}

	return cmd
}
