package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

// Backward compatibility
func CmdMemoIBCtoHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-ibc-to-hl [token-id] [destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:       cobra.ExactArgs(5),
		Short:      "DEPRECATED: Use 'create-memo --source=ibc --dest=hl' instead",
		Deprecated: "use 'create-memo --source=ibc --dest=hl' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcIBC)
			_ = newCmd.Flags().Set(FlagDst, DstHL)
			_ = newCmd.Flags().Set(FlagTokenID, args[0])
			_ = newCmd.Flags().Set(FlagDstDomain, args[1])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[2])
			_ = newCmd.Flags().Set(FlagAmount, args[3])
			_ = newCmd.Flags().Set(FlagMaxFee, args[4])

			return runCreateMemo(newCmd, []string{})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoEIBCtoHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-eibc-to-hl [eibc-fee] [token-id] [destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:       cobra.ExactArgs(6),
		Short:      "DEPRECATED: Use 'create-memo --source=eibc --dest=hl' instead",
		Deprecated: "use 'create-memo --source=eibc --dest=hl' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcEIBC)
			_ = newCmd.Flags().Set(FlagDst, DstHL)
			_ = newCmd.Flags().Set(FlagEIBCFee, args[0])
			_ = newCmd.Flags().Set(FlagTokenID, args[1])
			_ = newCmd.Flags().Set(FlagDstDomain, args[2])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[3])
			_ = newCmd.Flags().Set(FlagAmount, args[4])
			_ = newCmd.Flags().Set(FlagMaxFee, args[5])

			return runCreateMemo(newCmd, []string{})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoIBCtoIBC() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-ibc-to-ibc [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:       cobra.ExactArgs(3),
		Short:      "DEPRECATED: Use 'create-memo --source=ibc --dest=ibc' instead",
		Deprecated: "use 'create-memo --source=ibc --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcIBC)
			_ = newCmd.Flags().Set(FlagDst, DstIBC)
			_ = newCmd.Flags().Set(FlagChannel, args[0])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[1])
			_ = newCmd.Flags().Set(FlagTimeout, args[2])

			return runCreateMemo(newCmd, []string{})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoEIBCtoIBC() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-eibc-to-ibc [eibc-fee] [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:       cobra.ExactArgs(4),
		Short:      "DEPRECATED: Use 'create-memo --source=eibc --dest=ibc' instead",
		Deprecated: "use 'create-memo --source=eibc --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcEIBC)
			_ = newCmd.Flags().Set(FlagDst, DstIBC)
			_ = newCmd.Flags().Set(FlagEIBCFee, args[0])
			_ = newCmd.Flags().Set(FlagChannel, args[1])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[2])
			_ = newCmd.Flags().Set(FlagTimeout, args[3])

			return runCreateMemo(newCmd, []string{})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoHLtoIBCRaw() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-hl-to-ibc [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:       cobra.ExactArgs(3),
		Short:      "DEPRECATED: Use 'create-memo --source=hl --dest=ibc' instead",
		Deprecated: "use 'create-memo --source=hl --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcHL)
			_ = newCmd.Flags().Set(FlagDst, DstIBC)
			_ = newCmd.Flags().Set(FlagChannel, args[0])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[1])
			_ = newCmd.Flags().Set(FlagTimeout, args[2])
			_ = newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateMemo(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a human readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoHLtoHLRaw() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-hl-to-hl [token-id] [destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:       cobra.ExactArgs(5),
		Short:      "DEPRECATED: Use 'create-memo --source=hl --dest=hl' instead",
		Deprecated: "use 'create-memo --source=hl --dest=hl' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcHL)
			_ = newCmd.Flags().Set(FlagDst, DstHL)
			_ = newCmd.Flags().Set(FlagTokenID, args[0])
			_ = newCmd.Flags().Set(FlagDstDomain, args[1])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[2])
			_ = newCmd.Flags().Set(FlagAmount, args[3])
			_ = newCmd.Flags().Set(FlagMaxFee, args[4])
			_ = newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateMemo(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a human readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdTestHLtoIBCMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "hl-message [nonce] [src-domain] [src-contract] [dst-domain] [token-id] [hyperlane recipient] [amount] [ibc-source-chan] [ibc-recipient] [ibc timeout duration] [recovery-address]",
		Args:       cobra.ExactArgs(11),
		Short:      "DEPRECATED: Use 'create-hl-message --source=hl --dest=ibc' instead",
		Deprecated: "use 'create-hl-message --source=hl --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateHLMessage()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcHL)
			_ = newCmd.Flags().Set(FlagDst, DstIBC)
			_ = newCmd.Flags().Set(FlagNonce, args[0])
			_ = newCmd.Flags().Set(FlagSrcDomain, args[1])
			_ = newCmd.Flags().Set(FlagSrcContract, args[2])
			_ = newCmd.Flags().Set(FlagDstDomain, args[3])
			_ = newCmd.Flags().Set(FlagTokenID, args[4])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[5])
			_ = newCmd.Flags().Set(FlagAmount, args[6])
			_ = newCmd.Flags().Set(FlagChannel, args[7])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[8])
			_ = newCmd.Flags().Set(FlagTimeout, args[9])
			_ = newCmd.Flags().Set(FlagRecoveryAddr, args[10])
			_ = newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateHLMessage(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdHLMessageKaspaToHub() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "hl-message-kaspa [token-id] [hub recipient] [amount] [kas token placeholder] [kas-domain] [hub-domain]",
		Args:       cobra.ExactArgs(6),
		Short:      "DEPRECATED: Use 'create-hl-message --source=kaspa --dest=hub' instead",
		Deprecated: "use 'create-hl-message --source=kaspa --dest=hub' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateHLMessage()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcKaspa)
			_ = newCmd.Flags().Set(FlagDst, DstHub)
			_ = newCmd.Flags().Set(FlagTokenID, args[0])
			_ = newCmd.Flags().Set(FlagRecipientHub, args[1])
			_ = newCmd.Flags().Set(FlagAmount, args[2])
			_ = newCmd.Flags().Set(FlagKasToken, args[3])
			_ = newCmd.Flags().Set(FlagKasDomain, args[4])
			_ = newCmd.Flags().Set(FlagHubDomain, args[5])
			_ = newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateHLMessage(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdHLMessageKaspaToIBC() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "hl-message-kaspa-to-ibc [token-id] [hub recipient] [amount] [kas token placeholder] [kas-domain] [hub-domain] [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:       cobra.ExactArgs(9),
		Short:      "DEPRECATED: Use 'create-hl-message --source=kaspa --dest=ibc' instead",
		Deprecated: "use 'create-hl-message --source=kaspa --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateHLMessage()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcKaspa)
			_ = newCmd.Flags().Set(FlagDst, DstIBC)
			_ = newCmd.Flags().Set(FlagTokenID, args[0])
			_ = newCmd.Flags().Set(FlagRecipientHub, args[1])
			_ = newCmd.Flags().Set(FlagAmount, args[2])
			_ = newCmd.Flags().Set(FlagKasToken, args[3])
			_ = newCmd.Flags().Set(FlagKasDomain, args[4])
			_ = newCmd.Flags().Set(FlagHubDomain, args[5])
			_ = newCmd.Flags().Set(FlagChannel, args[6])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[7])
			_ = newCmd.Flags().Set(FlagTimeout, args[8])
			_ = newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateHLMessage(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdHLMessageKaspaToHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "hl-message-kaspa-to-hl [token-id] [hub recipient] [amount] [kas token placeholder] [kas-domain] [hub-domain] [hl-token-id] [hl-destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:       cobra.ExactArgs(11),
		Short:      "DEPRECATED: Use 'create-hl-message --source=kaspa --dest=hl' instead",
		Deprecated: "use 'create-hl-message --source=kaspa --dest=hl' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateHLMessage()
			newCmd.SetArgs([]string{})
			_ = newCmd.Flags().Set(FlagSrc, SrcKaspa)
			_ = newCmd.Flags().Set(FlagDst, DstHL)
			_ = newCmd.Flags().Set(FlagTokenID, args[0])
			_ = newCmd.Flags().Set(FlagRecipientHub, args[1])
			_ = newCmd.Flags().Set(FlagAmount, args[2])
			_ = newCmd.Flags().Set(FlagKasToken, args[3])
			_ = newCmd.Flags().Set(FlagKasDomain, args[4])
			_ = newCmd.Flags().Set(FlagHubDomain, args[5])
			_ = newCmd.Flags().Set(FlagDstTokenID, args[6])
			_ = newCmd.Flags().Set(FlagDstDomain, args[7])
			_ = newCmd.Flags().Set(FlagRecipientDst, args[8])
			_ = newCmd.Flags().Set(FlagDstAmount, args[9])
			_ = newCmd.Flags().Set(FlagMaxFee, args[10])
			_ = newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateHLMessage(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
