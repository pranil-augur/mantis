package main

import (
	"augur.ai/myapp/defs"
)

// Tagging a file with @flow(name) creates a flow with the given name.
// 1. It's the entry point for the program. Mantis looks for this tag across all files, and executes all top level flows.
// 2. Mantis can execute a flow by name using -F flow_name. This provides multiple entry and exit points 

install_nginx_ingress: {
	@flow(install_nginx_ingress)
	setup_namespace: {
		@task(mantis.core.K8s)
		config: defs.namespace
	}

	setup_service_account: {
		dep: setup_namespace
		@task(mantis.core.K8s)
		config: defs.serviceAccount
	}

	setup_ingress: {
		dep: setup_service_account
		@task(mantis.core.K8s)
		config: defs.deployment
	}

	setup_service: {
		dep: [setup_ingress, setup_service_account]
		@task(mantis.core.K8s)
		config: defs.service
	}
}
