package routing

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

var allowMethods = []string{"GET", "POST", "DELETE", "PATCH"}

func VerboseLog(e *echo.Echo, service string) {
	paths := e.Router().Routes()
	sort.SliceStable(paths, func(i, j int) bool {
		return paths[i].Path < paths[j].Path
	})

	table := tablewriter.NewTable(
		os.Stdout,
		tablewriter.WithRenderer(
			renderer.NewBlueprint(
				tw.Rendition{
					Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
				},
			),
		),
	)
	table.Header([]string{"Method", "Url"})

	for _, route := range paths {
		if slices.Contains(allowMethods, route.Method) {
			if strings.Contains(route.Path, "/*") {
				indexPath := strings.Replace(route.Path, "/*", "/index.html", -1)
				docsPath := strings.Replace(route.Path, "/*", "/doc.json", -1)
				_ = table.Append([]string{
					route.Method,
					fmt.Sprintf("http://%s%s", service, indexPath),
				})
				_ = table.Append([]string{
					route.Method,
					fmt.Sprintf("http://%s%s", service, docsPath),
				})
				continue
			}
			err := table.Append([]string{
				route.Method,
				fmt.Sprintf("http://%s%s", service, route.Path),
			})
			if err != nil {
				return
			}
		}
	}

	_ = table.Render()
}
