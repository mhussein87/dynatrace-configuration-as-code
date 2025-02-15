//go:build unit

/**
 * @license
 * Copyright 2020 Dynatrace LLC
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

package template_test

import (
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/config/template"
	"github.com/spf13/afero"
	"gotest.tools/assert"
	"path/filepath"
	"testing"
)

func TestLoadTemplate(t *testing.T) {
	testFilepath := filepath.FromSlash("proj/api/template.json")

	testFs := afero.NewMemMapFs()
	_ = testFs.MkdirAll("proj/api/", 0755)
	_ = afero.WriteFile(testFs, testFilepath, []byte("{ template: from_file }"), 0644)

	got, gotErr := template.NewFileTemplate(testFs, testFilepath)
	assert.NilError(t, gotErr)
	assert.Equal(t, got.ID(), testFilepath)
	assert.Equal(t, got.(*template.FileBasedTemplate).FilePath(), testFilepath)
	gotContent, err := got.Content()
	assert.NilError(t, err)
	assert.Equal(t, gotContent, "{ template: from_file }")
}

func TestLoadTemplate_ReturnsErrorIfFileDoesNotExist(t *testing.T) {
	testFilepath := filepath.FromSlash("proj/api/template.json")

	testFs := afero.NewMemMapFs()

	_, gotErr := template.NewFileTemplate(testFs, testFilepath)
	assert.ErrorContains(t, gotErr, testFilepath)
}

func TestLoadTemplate_WorksWithAnyPathSeparator(t *testing.T) {

	testFs := afero.NewReadOnlyFs(afero.NewOsFs())
	tests := []struct {
		name          string
		givenFilepath string
	}{
		{
			"windows path",
			`test-resources\template.json`,
		},
		{
			"relative windows path",
			`..\template\test-resources\template.json`,
		},
		{
			"unix path",
			`test-resources/template.json`,
		},
		{
			"relative unix path",
			`../template/test-resources/template.json`,
		},
		{
			"mixed path",
			`..\template\test-resources/template.json`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := template.NewFileTemplate(testFs, tt.givenFilepath)
			assert.NilError(t, gotErr)
		})
	}
}
