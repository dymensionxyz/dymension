package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"

	"github.com/dymensionxyz/dymension/app"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

const (
	FlagHeight     = "height"
	FlagTraceStore = "trace-store"
)

// InspectCmd dumps app state to JSON.

func InspectCmd(appExporter types.AppExporter, appCreator types.AppCreator, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect db state",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.OutOrStderr())

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			cdc := client.GetClientContextFromCmd(cmd).Codec

			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
			config.SetRoot(homeDir)

			if _, err := os.Stat(config.GenesisFile()); os.IsNotExist(err) {
				return err
			}

			/* --------------------------- read rollapps state from db --------------------------- */
			fmt.Println("Getting rollapps data...")
			db, err := openDB(config.RootDir, server.GetAppDBBackend(serverCtx.Viper))
			if err != nil {
				return err
			}

			traceWriterFile := serverCtx.Viper.GetString(FlagTraceStore)
			traceWriter, err := openTraceWriter(filepath.Clean(traceWriterFile))
			if err != nil {
				return err
			}

			height, _ := cmd.Flags().GetInt64(FlagHeight)
			exported, err := appExporter(serverCtx.Logger, db, traceWriter, height, false, []string{}, serverCtx.Viper)
			if err != nil {
				return fmt.Errorf("error exporting state: %v", err)
			}

			appState := exported.AppState

			// Temporary map to hold the unmarshalled content
			tmp := make(map[string]json.RawMessage)

			// Unmarshal AppState into the temporary map
			if err := json.Unmarshal(appState, &tmp); err != nil {
				fmt.Println("Error:", err)
				return err
			}

			// let's get the rollapp part in the json
			rollappStateTmp := app.GenesisState(tmp)[rollapptypes.ModuleName]

			// Unmarshal to rollapp state proto
			var rollappState rollapptypes.GenesisState
			if err := cdc.UnmarshalJSON(rollappStateTmp, &rollappState); err != nil {
				fmt.Println("Error:", err)
				return err
			}

			fmt.Println("Num of rollapps: ", len(rollappState.RollappList))
			fmt.Println("Num of state info: ", len(rollappState.StateInfoList))
			fmt.Println("Total size of states: ", humanize.Bytes(uint64(rollappState.Size())))

			sizes := make([]float64, len(rollappState.StateInfoList))
			for i, data := range rollappState.StateInfoList {
				sizes[i] = float64(data.Size())
			}

			mean := stat.Mean(sizes, nil)
			min := floats.Min(sizes)
			max := floats.Max(sizes)

			fmt.Printf("Mean: %v\n", humanize.Bytes(uint64(mean)))
			fmt.Printf("Min: %v\n", humanize.Bytes(uint64(min)))
			fmt.Printf("Max: %v\n", humanize.Bytes(uint64(max)))

			/* -------------------------- checking size on disk ------------------------- */
			// Get list of subdirectories
			fmt.Println("\n\nGetting storage on disk...")
			dataDir := filepath.Join(config.RootDir, "data")
			directories, err := os.ReadDir(dataDir)
			if err != nil {
				return fmt.Errorf("Error reading directory: %v", err)
			}

			for _, dir := range directories {
				if dir.IsDir() {
					size, err := getDirSize(filepath.Join(dataDir, dir.Name()))
					if err != nil {
						fmt.Printf("Error getting size for directory %s: %v", dir.Name(), err)
						continue
					}
					fmt.Printf("Size of %s: %s\n", dir.Name(), humanize.Bytes(uint64(size)))
				}
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().Int64(FlagHeight, -1, "Export state from a particular height (-1 means latest height)")
	cmd.Flags().String(FlagTraceStore, "", "Enable KVStore tracing to an output file")

	return cmd
}

// getDirSize returns the total size in bytes of a directory and its subdirectories.
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func openDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}

func openTraceWriter(traceWriterFile string) (w io.Writer, err error) {
	if traceWriterFile == "" {
		return
	}
	return os.OpenFile(
		traceWriterFile,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0o666,
	)
}
