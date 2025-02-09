package render

import "github.com/charmbracelet/glamour"

func RenderMD(content string) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
	)
	if err != nil {
		return "", err
	}

	str, err := renderer.Render(content)
	if err != nil {
		return "", err
	}

	return str, nil
}
