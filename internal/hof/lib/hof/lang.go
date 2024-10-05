// Copyright 2024 pdasika
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hof

const (
	// ExportsAlias is the default alias for Exports language files
	MantisTaskExports = "exports"

	// ExportsExtension is the file extension for Exports language files
	MantisTaskAlias = "alias"

	// MantisTaskPath is the default path for task outputs
	MantisTaskPath = "path"

	// MantisBackendConfigPath is the default path for backend configuration
	MantisBackendConfigPath = "mantis_state/"

	// MantisStateFilePath is the default path for the state file
	MantisStateFilePath = "mantis_state/mantis_%s.tfstate"

	// MantisTaskOuts is the default path for the task outputs
	MantisTaskOuts = "out"

	// MantisJsonConfig is the in-memory json file used to push config to OpenTF engine
	MantisJsonConfig = "mantis.json"
)
