package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/theakshaypant/regret"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version information for the regret CLI tool.`,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("regret version %s\n", regret.Version)
	fmt.Printf("Regex threat detection and analysis tool\n")
	fmt.Printf("\nFeatures:\n")
	fmt.Printf("  • Fast heuristics detection\n")
	fmt.Printf("  • NFA-based formal analysis (EDA/IDA)\n")
	fmt.Printf("  • Complexity scoring (0-100)\n")
	fmt.Printf("  • Adversarial input generation\n")
	fmt.Printf("\nFor more information: https://github.com/theakshaypant/regret\n")
}
