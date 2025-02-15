//go:build unit

/*
 * @license
 * Copyright 2023 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package delete

import (
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/delete/persistence"
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/delete/pointer"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestParseDeleteEntry(t *testing.T) {
	api := "auto-tag"
	name := "test entity"

	ctx := loaderContext{
		knownApis: toSetMap([]string{
			"management-zone",
			"auto-tag",
		}),
	}

	entry, err := parseDeleteEntry(&ctx, 0, api+deleteDelimiter+name)

	assert.NoError(t, err)
	assert.Equal(t, api, entry.Type)
	assert.Equal(t, name, entry.Identifier)
}

func TestParseSettingsDeleteEntry(t *testing.T) {
	cfgType := "builtin:tagging.auto"
	name := "test entity"

	ctx := loaderContext{
		knownApis: toSetMap([]string{
			"management-zone",
			"auto-tag",
		}),
	}

	entry, err := parseDeleteEntry(&ctx, 0, cfgType+deleteDelimiter+name)

	assert.NoError(t, err)
	assert.Equal(t, cfgType, entry.Type)
	assert.Equal(t, name, entry.Identifier)
}

func TestParseDeleteEntryWithMultipleSlashesShouldWork(t *testing.T) {
	api := "auto-tag"
	name := "test entity/entry"

	ctx := loaderContext{
		knownApis: toSetMap([]string{
			"management-zone",
			"auto-tag",
		}),
	}

	entry, err := parseDeleteEntry(&ctx, 0, api+deleteDelimiter+name)

	assert.NoError(t, err)
	assert.Equal(t, api, entry.Type)
	assert.Equal(t, name, entry.Identifier)
}

func TestParseDeleteEntryInvalidEntryWithoutDelimiterShouldFail(t *testing.T) {
	value := "auto-tag"

	ctx := loaderContext{
		knownApis: toSetMap([]string{
			"management-zone",
			"auto-tag",
		}),
	}

	_, err := parseDeleteEntry(&ctx, 0, value)

	assert.NotNil(t, err, "value `%s` should return error", value)
}

func TestParseDeleteFileDefinitions(t *testing.T) {
	api := "auto-tag"
	name := "test entity/entry"
	entity := api + deleteDelimiter + name

	api2 := "management-zone"
	name2 := "test entity/entry"
	entity2 := api2 + deleteDelimiter + name2

	ctx := loaderContext{
		knownApis: toSetMap([]string{
			"management-zone",
			"auto-tag",
		}),
	}

	result, errors := parseDeleteFileDefinition(&ctx, persistence.FileDefinition{
		DeleteEntries: []interface{}{
			entity,
			entity2,
		},
	})

	assert.Equal(t, 0, len(errors))
	assert.Equal(t, 2, len(result))

	apiEntities := result[api]

	assert.Equal(t, 1, len(apiEntities))
	assert.Equal(t, pointer.DeletePointer{
		Type:       api,
		Identifier: name,
	}, apiEntities[0])

	api2Entities := result[api2]

	assert.Equal(t, 1, len(api2Entities))
	assert.Equal(t, pointer.DeletePointer{
		Type:       api2,
		Identifier: name2,
	}, api2Entities[0])
}

func TestParseDeleteFileDefinitionsWithInvalidDefinition(t *testing.T) {
	api := "auto-tag"
	name := "test entity/entry"
	entity := api + deleteDelimiter + name

	api2 := "management-zone"
	name2 := "test entity/entry"
	entity2 := api2 + deleteDelimiter + name2

	ctx := loaderContext{
		knownApis: toSetMap([]string{
			"management-zone",
			"auto-tag",
		}),
	}

	result, errors := parseDeleteFileDefinition(&ctx, persistence.FileDefinition{
		DeleteEntries: []interface{}{
			entity,
			entity2,
			"invalid-definition",
		},
	})

	assert.Equal(t, 1, len(errors))
	assert.Equal(t, 0, len(result))
}

func TestLoadEntriesToDelete(t *testing.T) {

	tests := []struct {
		name             string
		givenFileContent string
		want             DeleteEntries
	}{
		{
			"Loads simple file",
			`delete:
- management-zone/test entity/entities
- auto-tag/random tag
`,
			DeleteEntries{
				"auto-tag": {
					{
						Type:       "auto-tag",
						Identifier: "random tag",
					},
				},
				"management-zone": {
					{
						Type:       "management-zone",
						Identifier: "test entity/entities",
					},
				},
			},
		},
		{
			"Loads Settings",
			`delete:
- management-zone/test entity/entities
- builtin:auto.tagging/random tag
`,
			DeleteEntries{
				"builtin:auto.tagging": {
					{
						Type:       "builtin:auto.tagging",
						Identifier: "random tag",
					},
				},
				"management-zone": {
					{
						Type:       "management-zone",
						Identifier: "test entity/entities",
					},
				},
			},
		},
		{
			"Loads Full Format",
			`delete:
- project: "myProject"
  type: management-zone
  name: test entity/entities
- project: some-project
  type: builtin:auto.tagging
  id: my-tag
`,
			DeleteEntries{
				"builtin:auto.tagging": {
					{
						Project:    "some-project",
						Type:       "builtin:auto.tagging",
						Identifier: "my-tag",
					},
				},
				"management-zone": {
					{
						Type:       "management-zone",
						Identifier: "test entity/entities",
					},
				},
			},
		},
		{
			"Loads Mixed Format",
			`delete:
- "management-zone/test entity/entities"
- project: some-project
  type: builtin:auto.tagging
  id: my-tag
`,
			DeleteEntries{
				"builtin:auto.tagging": {
					{
						Project:    "some-project",
						Type:       "builtin:auto.tagging",
						Identifier: "my-tag",
					},
				},
				"management-zone": {
					{
						Type:       "management-zone",
						Identifier: "test entity/entities",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			deleteFile, err := filepath.Abs("delete.yaml")
			assert.NoError(t, err)

			fs := afero.NewMemMapFs()

			err = afero.WriteFile(fs, deleteFile, []byte(tt.givenFileContent), 0666)
			assert.NoError(t, err)

			result, errors := LoadEntriesToDelete(fs, deleteFile)

			assert.Empty(t, errors)
			assert.Equal(t, 2, len(result))
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestLoadEntriesToDeleteWithInvalidEntry(t *testing.T) {
	fileContent := `delete:
- management-zone/test entity/entities
- auto-invalid
`

	workingDir := filepath.FromSlash("/home/test/monaco")
	deleteFileName := "delete.yaml"
	deleteFilePath := filepath.Join(workingDir, deleteFileName)

	fs := afero.NewMemMapFs()
	err := fs.MkdirAll(workingDir, 0777)

	assert.NoError(t, err)

	err = afero.WriteFile(fs, deleteFilePath, []byte(fileContent), 0666)
	assert.NoError(t, err)

	result, errors := LoadEntriesToDelete(fs, deleteFilePath)

	assert.Equal(t, 1, len(errors))
	assert.Equal(t, 0, len(result))
}

func TestLoadEntriesToDeleteNonExistingFile(t *testing.T) {
	workingDir := filepath.FromSlash("/home/test/monaco")

	fs := afero.NewMemMapFs()
	err := fs.MkdirAll(workingDir, 0777)

	assert.NoError(t, err)

	result, errors := LoadEntriesToDelete(fs, "/home/test/monaco/non-existing-delete.yaml")

	assert.Equal(t, 1, len(errors))
	assert.Equal(t, 0, len(result))
}

func TestLoadEntriesToDeleteWithMalformedFile(t *testing.T) {
	fileContent := `deleting:
- auto-invalid
`

	workingDir := filepath.FromSlash("/home/test/monaco")
	deleteFileName := "delete.yaml"
	deleteFilePath := filepath.Join(workingDir, deleteFileName)

	fs := afero.NewMemMapFs()
	err := fs.MkdirAll(workingDir, 0777)

	assert.NoError(t, err)

	err = afero.WriteFile(fs, deleteFilePath, []byte(fileContent), 0666)
	assert.NoError(t, err)

	result, errors := LoadEntriesToDelete(fs, deleteFilePath)

	assert.Equal(t, 1, len(errors))
	assert.Equal(t, 0, len(result))
}

func TestLoadEntriesToDeleteWithEmptyFile(t *testing.T) {
	workingDir := filepath.FromSlash("/home/test/monaco")
	deleteFileName := "empty_delete_file.yaml"
	deleteFilePath := filepath.Join(workingDir, deleteFileName)

	fs := afero.NewMemMapFs()
	err := fs.MkdirAll(workingDir, 0777)

	assert.NoError(t, err)

	err = afero.WriteFile(fs, deleteFilePath, []byte{}, 0666)
	assert.NoError(t, err)

	result, errors := LoadEntriesToDelete(fs, deleteFilePath)

	assert.Equal(t, 1, len(errors))
	assert.Equal(t, 0, len(result))
}
