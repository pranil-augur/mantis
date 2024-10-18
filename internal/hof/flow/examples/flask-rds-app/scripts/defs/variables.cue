package defs


#providers: {
    provider: {
		"aws": {}
	}
	terraform: {
		required_providers: {
			aws: {
				source:  "hashicorp/aws"
				version: ">= 4.67.0"
			}
		}
	}
}

project_name: "flask-rds-app"

common: {   
    // Common configurations for the RDS setup
    project_name: "flask-rds-app"

    // Database credentials (should be secured in practice)
    db_username: "admin"
    db_password: "supersecretpassword"

    db_name: "rds_app"

    container_repo: "registry.gitlab.com/flashresolve1/augur:v1"
}

