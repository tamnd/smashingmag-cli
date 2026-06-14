package cli

import (
	"github.com/spf13/cobra"
)

// articlesCmd returns the articles command.
func (a *App) articlesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "articles",
		Short: "List latest Smashing Magazine articles",
		Long:  `Fetch the Smashing Magazine RSS feed and print articles as table, JSON, or other formats.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching articles from Smashing Magazine...")
			articles, err := a.client.Articles(cmd.Context(), n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(articles, len(articles))
		},
	}
}
