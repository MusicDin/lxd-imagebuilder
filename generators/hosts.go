package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	lxdShared "github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"

	"github.com/canonical/lxd-imagebuilder/image"
	"github.com/canonical/lxd-imagebuilder/shared"
)

type hosts struct {
	common
}

// RunLXC creates a LXC specific entry in the hosts file.
func (g *hosts) RunLXC(img *image.LXCImage, target shared.DefinitionTargetLXC) error {
	// Skip if the file doesn't exist
	if !lxdShared.PathExists(filepath.Join(g.sourceDir, g.defFile.Path)) {
		return nil
	}

	// Read the current content
	content, err := os.ReadFile(filepath.Join(g.sourceDir, g.defFile.Path))
	if err != nil {
		return fmt.Errorf("Failed to read file %q: %w", filepath.Join(g.sourceDir, g.defFile.Path), err)
	}

	// Replace hostname with placeholder
	content = []byte(strings.ReplaceAll(string(content), "lxd-imagebuilder", "LXC_NAME"))

	// Add a new line if needed
	if !strings.Contains(string(content), "LXC_NAME") {
		content = append([]byte("127.0.1.1\tLXC_NAME\n"), content...)
	}

	f, err := os.Create(filepath.Join(g.sourceDir, g.defFile.Path))
	if err != nil {
		return fmt.Errorf("Failed to create file %q: %w", filepath.Join(g.sourceDir, g.defFile.Path), err)
	}

	defer f.Close()

	// Overwrite the file
	_, err = f.Write(content)
	if err != nil {
		return fmt.Errorf("Failed to write to file %q: %w", filepath.Join(g.sourceDir, g.defFile.Path), err)
	}

	// Add hostname path to LXC's templates file
	err = img.AddTemplate(g.defFile.Path)
	if err != nil {
		return fmt.Errorf("Failed to add template: %w", err)
	}

	return nil
}

// RunLXD creates a hosts template.
func (g *hosts) RunLXD(img *image.LXDImage, target shared.DefinitionTargetLXD) error {
	// Skip if the file doesn't exist
	if !lxdShared.PathExists(filepath.Join(g.sourceDir, g.defFile.Path)) {
		return nil
	}

	templateDir := filepath.Join(g.cacheDir, "templates")

	// Create templates path
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create directory %q: %w", templateDir, err)
	}

	// Read the current content
	content, err := os.ReadFile(filepath.Join(g.sourceDir, g.defFile.Path))
	if err != nil {
		return fmt.Errorf("Failed to read file %q: %w", filepath.Join(g.sourceDir, g.defFile.Path), err)
	}

	// Replace hostname with placeholder
	content = []byte(strings.ReplaceAll(string(content), "lxd-imagebuilder", "{{ container.name }}"))

	// Add a new line if needed
	if !strings.Contains(string(content), "{{ container.name }}") {
		content = append([]byte("127.0.1.1\t{{ container.name }}\n"), content...)
	}

	// Write the template
	err = os.WriteFile(filepath.Join(templateDir, "hosts.tpl"), content, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write file %q: %w", filepath.Join(templateDir, "hosts.tpl"), err)
	}

	img.Metadata.Templates[g.defFile.Path] = &api.ImageMetadataTemplate{
		Template:   "hosts.tpl",
		Properties: g.defFile.Template.Properties,
		When:       g.defFile.Template.When,
	}

	if len(g.defFile.Template.When) == 0 {
		img.Metadata.Templates[g.defFile.Path].When = []string{
			"create",
			"copy",
		}
	}

	return nil
}

// Run does nothing.
func (g *hosts) Run() error {
	return nil
}
