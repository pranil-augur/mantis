package staticwebsite

import (
	"augur.ai/static-website/defs"
)

// Tagging a file with @flow(name) creates a flow with the given name.
// 1. It's the entry point for the program. Mantis looks for this tag across all files, and executes all top level flows.
// 2. Mantis can execute a flow by name using -F flow_name. This provides multiple entry and exit points 

@flow(static_website_setup)
install_static_website: {

	setup_providers: {
		@task(mantis.core.TF)
		config: defs.#providers
	}

	setup_vpc: {
		@task(mantis.core.TF)
		dep:    setup_providers
		config: defs.vpc
		exports: [{
			jqpath: ".module.vpc.aws_vpc.this[0].id"
			var:  "vpc_id"
		}, {
			jqpath: ".module.vpc.aws_subnet.public[].id"
			var:  "public_subnet_ids"
		}, {
			jqpath: ".module.vpc.aws_subnet.public[0].id"
			var:  "public_subnet_id"
		}, {
			jqpath: ".module.vpc.aws_subnet.private[].id"
			var:  "private_subnet_ids"
		}, {
			jqpath: ".module.vpc.aws_vpc.this[0].cidr_block"
			var:  "vpc_cidr_block"
		}, {
			jqpath: ".module.vpc.aws_eip.nat[].public_ip"
			var:  "nat_public_ips"
		}, {
			jqpath: ".module.vpc.aws_default_security_group.this[].id"
			var:  "default_security_group_id"
		}]
	}

	db_subnet_group: {
		@task(mantis.core.TF)
		dep:    setup_vpc
		config: defs.subnet_group
		exports: [{
			jqpath: ".aws_db_subnet_group.this[0].id"
			var:  "subnet_group_ids"
		}]
	}

	setup_rds: {
		@task(mantis.core.TF)
		dep: [setup_vpc, db_subnet_group]
		config: defs.rds
		exports: [{
			jqpath: ".aws_db_instance.this[0].endpoint"
			var:  "rds_endpoint"
		}]
	}

	setup_ec2: {
		@task(mantis.core.TF)
		dep:    setup_rds
		config: defs.ec2
	}
}
