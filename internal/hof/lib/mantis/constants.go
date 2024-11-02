/*
/*
 * Copyright (c) 2024 Augur AI, Inc.
 * This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. 
 * If a copy of the MPL was not distributed with this file, you can obtain one at https://mozilla.org/MPL/2.0/.
 *
 
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package mantis

const (
	// ExportsAlias is the default alias for Exports language files
	MantisTaskExports = "exports"

	// ExportsExtension is the file extension for Exports language files
	MantisVar = "var"

	// MantisDataSourcePath is the jqpath from which the output is extracted
	MantisDataSourcePath = "jqpath"

	// MantisExportAs is the default alias for task outputs
	MantisExportAs = "as"

	// MantisBackendConfigPath is the default path for backend configuration
	MantisBackendConfigPath = "mantis_state/"

	// MantisStateFilePath is the default path for the state file
	MantisStateFilePath = "mantis_state/mantis_%s.tfstate"

	// MantisTaskOuts is the default path for the task outputs
	MantisTaskOuts = "out"

	// MantisJsonConfig is the in-memory json file used to push config to OpenTF engine
	MantisJsonConfig = "mantis.json"
)

type KubernetesResource int

const (
	Deployment KubernetesResource = iota
	Service
	Ingress
	Secret
	ConfigMap
	StatefulSet
	DaemonSet
	Job
	Pod
	Namespace
	PersistentVolumeClaim
	PersistentVolume
	ServiceAccount
	Role
	RoleBinding
	ClusterRole
	ClusterRoleBinding
)

// MantisKubernetesResources is the list of kubernetes resources
var MantisKubernetesResourceNames = map[KubernetesResource]string{
	ConfigMap:             "configmap",
	Deployment:            "deployment",
	Service:               "service",
	Ingress:               "ingress",
	Secret:                "secret",
	StatefulSet:           "statefulset",
	DaemonSet:             "daemonset",
	Job:                   "job",
	Pod:                   "pod",
	Namespace:             "namespace",
	PersistentVolumeClaim: "persistentvolumeclaim",
	PersistentVolume:      "persistentvolume",
	ServiceAccount:        "serviceaccount",
	Role:                  "role",
	RoleBinding:           "rolebinding",
	ClusterRole:           "clusterrole",
	ClusterRoleBinding:    "clusterrolebinding",
}
