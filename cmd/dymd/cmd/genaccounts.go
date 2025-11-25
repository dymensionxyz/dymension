package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strings" // Added for cleaner string operations

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	address "cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/evmos/ethermint/crypto/hd"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Constants for vesting flags
const (
	flagVestingStart = "vesting-start-time"
	flagVestingEnd   = "vesting-end-time"
	flagVestingAmt   = "vesting-amount"
)

// AddGenesisAccountCmd returns the add-genesis-account cobra command.
// This command allows adding an account with a specific balance and an optional vesting schedule
// to the network's genesis file.
func AddGenesisAccountCmd(defaultNodeHome string, addressCodec address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account [address_or_key_name] [coin][,[coin]]",
		Short: "Add a genesis account with an optional vesting schedule to genesis.json",
		Long: strings.TrimSpace(`
Add a genesis account to genesis.json. The provided account must specify
the account address or a key name available in the local Keyring. 
A list of initial coins (e.g., 1000stake,500uatom) must also be provided. 
Accounts may optionally be supplied with vesting parameters using the --vesting flags.
`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize ClientContext and set custom Keyring options for Ethermint (EthSecp256k1)
			clientCtx := client.GetClientContextFromCmd(cmd).WithKeyringOptions(hd.EthSecp256k1Option())
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			// --- 1. Address Resolution ---

			var addrBytes []byte
			var err error
			
			// Attempt to parse the argument as a raw address (AccAddress)
			addrBytes, err = addressCodec.StringToBytes(args[0])
			
			if err != nil {
				// Address parsing failed, assume it is a Keyring key name.
				
				// Initialize or retrieve keyring
				var kr keyring.Keyring
				inBuf := bufio.NewReader(cmd.InOrStdin())
				keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
				
				if clientCtx.Keyring == nil {
					// NOTE: We rely on the clientCtx to eventually be populated, but
					// if keyringBackend is explicitly provided and Keyring is nil, create it.
					kr, err = keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf, clientCtx.Codec, hd.EthSecp256k1Option())
					if err != nil {
						return fmt.Errorf("failed to create keyring: %w", err)
					}
				} else {
					kr = clientCtx.Keyring
				}

				info, err := kr.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keyring using name '%s': %w", args[0], err)
				}

				addrBytes, err = info.GetAddress()
				if err != nil {
					return fmt.Errorf("failed to get bytes address from key info: %w", err)
				}
			}

			// Convert resolved address bytes to AccAddress type
			accAddr := sdk.AccAddress(addrBytes)
			
			// --- 2. Flag and Coin Parsing ---

			// Parse vesting flags
			vestingStart, err := cmd.Flags().GetInt64(flagVestingStart)
			if err != nil {
				return fmt.Errorf("failed to parse vesting start time: %w", err)
			}
			vestingEnd, err := cmd.Flags().GetInt64(flagVestingEnd)
			if err != nil {
				return fmt.Errorf("failed to parse vesting end time: %w", err)
			}
			vestingAmtStr, err := cmd.Flags().GetString(flagVestingAmt)
			if err != nil {
				return fmt.Errorf("failed to parse vesting amount string: %w", err)
			}

			// Parse total coins to be assigned
			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse total coins: %w", err)
			}

			// Parse vesting amount
			vestingAmt, err := sdk.ParseCoinsNormalized(vestingAmtStr)
			if err != nil {
				return fmt.Errorf("failed to parse vesting amount coins: %w", err)
			}
			
			// --- 3. Account Creation ---

			var genAccount authtypes.GenesisAccount
			balances := banktypes.Balance{Address: accAddr.String(), Coins: coins.Sort()}
			baseAccount := authtypes.NewBaseAccount(accAddr, nil, 0, 0)

			if !vestingAmt.IsZero() {
				// --- Vesting Account Logic ---
				
				baseVestingAccount, err := authvesting.NewBaseVestingAccount(baseAccount, vestingAmt.Sort(), vestingEnd)
				if err != nil {
					return fmt.Errorf("failed to create base vesting account: %w", err)
				}

				// Critical check: vesting amount cannot exceed total account balance
				if (balances.Coins.IsZero() && !baseVestingAccount.OriginalVesting.IsZero()) ||
					baseVestingAccount.OriginalVesting.IsAnyGT(balances.Coins) {
					return errors.New("vesting amount cannot be greater than total amount")
				}

				// Determine concrete vesting type
				switch {
				case vestingStart != 0 && vestingEnd != 0:
					// Continuous vesting requires both start and end time
					genAccount = authvesting.NewContinuousVestingAccountRaw(baseVestingAccount, vestingStart)

				case vestingEnd != 0:
					// Delayed vesting requires only the end time
					genAccount = authvesting.NewDelayedVestingAccountRaw(baseVestingAccount)

				default:
					return errors.New("invalid vesting parameters: must supply start and end time (continuous) or end time (delayed)")
				}
			} else {
				// --- Regular Ethermint EVM Account ---
				// Default to EthAccount with an empty code hash
				genAccount = &ethermint.EthAccount{
					BaseAccount: baseAccount,
					CodeHash:    common.BytesToHash(evmtypes.EmptyCodeHash).Hex(),
				}
			}

			if err := genAccount.Validate(); err != nil {
				return fmt.Errorf("failed to validate new genesis account: %w", err)
			}
			
			// --- 4. Genesis File Update ---

			genFilePath := config.GenesisFile()
			appState, appGenesis, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// --- Update Auth Module State ---
			authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to unpack existing accounts: %w", err)
			}

			if accs.Contains(accAddr) {
				return fmt.Errorf("cannot add account at existing address %s", accAddr.String())
			}

			// Add the new account and sanitize the accounts list
			accs = append(accs, genAccount)
			accs = authtypes.SanitizeGenesisAccounts(accs)

			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				return fmt.Errorf("failed to pack accounts into Any type: %w", err)
			}
			authGenState.Accounts = genAccs

			authGenStateBz, err := cdc.MarshalJSON(&authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}

			appState[authtypes.ModuleName] = authGenStateBz

			// --- Update Bank Module State ---
			bankGenState := banktypes.GetGenesisStateFromAppState(cdc, appState)
			
			// Add balances and update supply
			bankGenState.Balances = append(bankGenState.Balances, balances)
			bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)
			bankGenState.Supply = bankGenState.Supply.Add(balances.Coins...)

			bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}

			appState[banktypes.ModuleName] = bankGenStateBz

			// --- Finalize and Export ---
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			appGenesis.AppState = appStateJSON
			
			// Write the updated genesis to file
			return genutil.ExportGenesisFile(appGenesis, genFilePath)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flagVestingAmt, "", "Amount of coins for vesting accounts (e.g., 1000stake)")
	cmd.Flags().Int64(flagVestingStart, 0, "Schedule start time (unix epoch) for continuous vesting accounts")
	cmd.Flags().Int64(flagVestingEnd, 0, "Schedule end time (unix epoch) for vesting accounts")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
