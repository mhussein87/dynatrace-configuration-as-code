// @license
// Copyright 2022 Dynatrace LLC
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build unit

package download

import (
	config "github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/config/v2"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/config/v2/coordinate"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/config/v2/parameter/value"
	"github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/config/v2/template"
	v2 "github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/project/v2"
	"github.com/spf13/afero"
	"path/filepath"
	"testing"
)

func TestWriteToDisk(t *testing.T) {

	type args struct {
		fs                afero.Fs
		downloadedConfigs v2.ConfigsPerType
		projectName       string
		tokenEnvVarName   string
		environmentUrl    string
		outputFolder      string
		timestampString   string
	}
	tests := []struct {
		name             string
		args             args
		wantOutputFolder string
		wantManifestFile string
		wantErr          bool
	}{
		{
			"creates expected files",
			args{
				fs: emptyTestFs(),
				downloadedConfigs: v2.ConfigsPerType{
					"test-api": []config.Config{
						{
							Template:    template.CreateTemplateFromString("template.json", "{}"),
							Coordinate:  coordinate.Coordinate{},
							Group:       "",
							Environment: "",
							Parameters: config.Parameters{
								"name": value.New("test-config"),
							},
							References: nil,
							Skip:       false,
						},
					},
				},
				projectName:     "test-project",
				tokenEnvVarName: "TEST_ENV_TOKEN",
				environmentUrl:  "env.url.com",
				outputFolder:    "test-output",
				timestampString: "TESTING_TIME",
			},
			"test-output",
			"manifest.yaml",
			false,
		},
		{
			"creates 'download_{TIMESTAMP}' output if no output folder is defined",
			args{
				fs: emptyTestFs(),
				downloadedConfigs: v2.ConfigsPerType{
					"test-api": []config.Config{
						{
							Template:    template.CreateTemplateFromString("template.json", "{}"),
							Coordinate:  coordinate.Coordinate{},
							Group:       "",
							Environment: "",
							Parameters: config.Parameters{
								"name": value.New("test-config"),
							},
							References: nil,
							Skip:       false,
						},
					},
				},
				projectName:     "test-project",
				tokenEnvVarName: "TEST_ENV_TOKEN",
				environmentUrl:  "env.url.com",
				outputFolder:    "",
				timestampString: "TESTING_TIME",
			},
			"download_TESTING_TIME",
			"manifest.yaml",
			false,
		},
		{
			"creates 'manifest_{TIMESTAMP}' if a manifest.yaml already exists",
			args{
				fs: testFsWithWithExistingManifest("test-output"),
				downloadedConfigs: v2.ConfigsPerType{
					"test-api": []config.Config{
						{
							Template:    template.CreateTemplateFromString("template.json", "{}"),
							Coordinate:  coordinate.Coordinate{},
							Group:       "",
							Environment: "",
							Parameters: config.Parameters{
								"name": value.New("test-config"),
							},
							References: nil,
							Skip:       false,
						},
					},
				},
				projectName:     "test-project",
				tokenEnvVarName: "TEST_ENV_TOKEN",
				environmentUrl:  "env.url.com",
				outputFolder:    "test-output",
				timestampString: "TESTING_TIME",
			},
			"test-output",
			"manifest_TESTING_TIME.yaml",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := CreateProjectData(tt.args.downloadedConfigs, tt.args.projectName) //using CreateProject data to simplify test struct setup

			if err := writeToDisk(tt.args.fs, proj, tt.args.tokenEnvVarName, tt.args.environmentUrl, tt.args.outputFolder, tt.args.timestampString); (err != nil) != tt.wantErr {
				t.Errorf("WriteToDisk() error = %v, wantErr %v", err, tt.wantErr)
			}

			if exists, err := afero.Exists(tt.args.fs, tt.args.outputFolder); err != nil || !exists {
				t.Errorf("WriteToDisk(): expected outputfolder %v was not created", tt.args.outputFolder)
			}

			if exists, err := afero.Exists(tt.args.fs, tt.wantOutputFolder); err != nil || !exists {
				t.Errorf("WriteToDisk(): expected outputfolder %v was not created", tt.wantOutputFolder)
			}

			expectedProjectFolder := filepath.Join(tt.wantOutputFolder, tt.args.projectName)
			if exists, err := afero.Exists(tt.args.fs, expectedProjectFolder); err != nil || !exists {
				t.Errorf("WriteToDisk(): expected project %v was not created", expectedProjectFolder)
			}

			expectedManifest := filepath.Join(tt.wantOutputFolder, tt.wantManifestFile)
			if exists, err := afero.Exists(tt.args.fs, expectedManifest); err != nil || !exists {
				t.Errorf("WriteToDisk(): expected manifest %v was not created", expectedManifest)
			}

		})
	}
}

func emptyTestFs() afero.Fs {
	return afero.NewMemMapFs()
}

func testFsWithWithExistingManifest(folder string) afero.Fs {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll(folder, 0777)
	_ = afero.WriteFile(fs, filepath.Join(folder, "manifest.yaml"), []byte{}, 0777)
	return fs
}
