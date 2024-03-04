package generators

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/canonical/lxd-imagebuilder/image"
	"github.com/canonical/lxd-imagebuilder/shared"
)

func TestTemplateGeneratorRunLXD(t *testing.T) {
	cacheDir, err := os.MkdirTemp(os.TempDir(), "lxd-imagebuilder-test-")
	require.NoError(t, err)

	rootfsDir := filepath.Join(cacheDir, "rootfs")

	setup(t, cacheDir)
	defer teardown(cacheDir)

	definition := shared.Definition{
		Image: shared.DefinitionImage{
			Distribution: "ubuntu",
			Release:      "artful",
		},
	}

	generator, err := Load("template", nil, cacheDir, rootfsDir, shared.DefinitionFile{
		Generator: "template",
		Name:      "template",
		Content:   "==test==",
		Path:      "/root/template",
	}, definition)
	require.IsType(t, &template{}, generator)
	require.NoError(t, err)

	image := image.NewLXDImage(context.TODO(), cacheDir, "", cacheDir, definition)

	err = os.MkdirAll(filepath.Join(cacheDir, "rootfs", "root"), 0755)
	require.NoError(t, err)

	createTestFile(t, filepath.Join(cacheDir, "rootfs", "root", "template"), "--test--")

	err = generator.RunLXD(image, shared.DefinitionTargetLXD{})
	require.NoError(t, err)

	validateTestFile(t, filepath.Join(cacheDir, "templates", "template.tpl"), "==test==\n")
	validateTestFile(t, filepath.Join(cacheDir, "rootfs", "root", "template"), "--test--")
}

func TestTemplateGeneratorRunLXDDefaultWhen(t *testing.T) {
	cacheDir, err := os.MkdirTemp(os.TempDir(), "lxd-imagebuilder-test-")
	require.NoError(t, err)

	rootfsDir := filepath.Join(cacheDir, "rootfs")

	setup(t, cacheDir)
	defer teardown(cacheDir)

	definition := shared.Definition{
		Image: shared.DefinitionImage{
			Distribution: "ubuntu",
			Release:      "artful",
		},
	}

	generator, err := Load("template", nil, cacheDir, rootfsDir, shared.DefinitionFile{
		Generator: "template",
		Name:      "test-default-when",
		Content:   "==test==",
		Path:      "test-default-when",
	}, definition)
	require.IsType(t, &template{}, generator)
	require.NoError(t, err)

	image := image.NewLXDImage(context.TODO(), cacheDir, "", cacheDir, definition)

	err = generator.RunLXD(image, shared.DefinitionTargetLXD{})
	require.NoError(t, err)

	generator, err = Load("template", nil, cacheDir, rootfsDir, shared.DefinitionFile{
		Generator: "template",
		Name:      "test-when",
		Content:   "==test==",
		Path:      "test-when",
		Template: shared.DefinitionFileTemplate{
			When: []string{"create"},
		},
	}, definition)
	require.IsType(t, &template{}, generator)
	require.NoError(t, err)

	err = generator.RunLXD(image, shared.DefinitionTargetLXD{})
	require.NoError(t, err)

	testvalue := []string{"create", "copy"}
	require.Equal(t, image.Metadata.Templates["test-default-when"].When, testvalue)

	testvalue = []string{"create"}
	require.Equal(t, image.Metadata.Templates["test-when"].When, testvalue)
}
