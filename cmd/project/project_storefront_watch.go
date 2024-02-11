package project

import (
	"github.com/FriendsOfShopware/shopware-cli/extension"
	"github.com/FriendsOfShopware/shopware-cli/shop"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

var projectStorefrontWatchCmd = &cobra.Command{
	Use:   "storefront-watch [path]",
	Short: "Starts the Shopware Storefront Watcher",
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectRoot string
		var err error

		if len(args) == 1 {
			projectRoot = args[0]
		} else if projectRoot, err = findClosestShopwareProject(); err != nil {
			return err
		}

		if err := extension.LoadSymfonyEnvFile(projectRoot); err != nil {
			return err
		}

		shopCfg, err := shop.ReadConfig(projectConfigPath, true)
		if err != nil {
			return err
		}

		if err := filterAndWritePluginJson(cmd, projectRoot, shopCfg); err != nil {
			return err
		}

		if err := runTransparentCommand(commandWithRoot(exec.CommandContext(cmd.Context(), "php", "bin/console", "feature:dump"), projectRoot)); err != nil {
			return err
		}

		activeOnly := "--active-only"

		if !themeCompileSupportsActiveOnly(projectRoot) {
			activeOnly = "-v"
		}

		if err := runTransparentCommand(commandWithRoot(exec.CommandContext(cmd.Context(), "php", "bin/console", "theme:compile", activeOnly), projectRoot)); err != nil {
			return err
		}

		if err := runTransparentCommand(commandWithRoot(exec.CommandContext(cmd.Context(), "php", "bin/console", "theme:dump"), projectRoot)); err != nil {
			return err
		}

		if err := os.Setenv("PROJECT_ROOT", projectRoot); err != nil {
			return err
		}

		if err := os.Setenv("STOREFRONT_ROOT", extension.PlatformPath(projectRoot, "Storefront", "")); err != nil {
			return err
		}

		return runTransparentCommand(commandWithRoot(exec.CommandContext(cmd.Context(), "npm", "run-script", "hot-proxy"), extension.PlatformPath(projectRoot, "Storefront", "Resources/app/storefront")))
	},
}

func themeCompileSupportsActiveOnly(projectRoot string) bool {
	themeFile := extension.PlatformPath(projectRoot, "Storefront", "Theme/Command/ThemeCompileCommand.php")

	bytes, err := os.ReadFile(themeFile)
	if err != nil {
		return false
	}

	return !strings.Contains(string(bytes), "active-only")
}

func init() {
	projectRootCmd.AddCommand(projectStorefrontWatchCmd)
	projectStorefrontWatchCmd.PersistentFlags().String("only-extensions", "", "Only watch the given extensions")
	projectStorefrontWatchCmd.PersistentFlags().String("skip-extensions", "", "Skips the given extensions")
}
