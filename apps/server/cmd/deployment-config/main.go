package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/deploymentconfig"
)

func main() {
	publicURL := flag.String("public-url", "", "public HTTPS origin")
	version := flag.String("version", "", "release version without v prefix")
	output := flag.String("output", ".env.production", "output environment file")
	flag.Parse()

	if err := deploymentconfig.Write(deploymentconfig.Options{
		PublicURL: *publicURL,
		Version:   *version,
		Output:    *output,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "deployment config:", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s with mode 0600. No secret values were printed.\n", *output)
}
