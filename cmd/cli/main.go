package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hedge-fund",
	Short: "AI-powered hedge fund trading system",
	Long: `A modern hedge fund system with AI-powered investment analysis.

Features:
- Multi-agent AI investment analysis
- Real-time portfolio management
- Risk assessment and position sizing
- Market data integration`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add commands will be implemented later
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hedge-fund v0.1.0")
	},
}