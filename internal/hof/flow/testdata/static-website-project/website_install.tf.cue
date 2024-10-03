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
		@task(opentf.TF)
		config: defs.#providers
	}

	setup_vpc: {
		@task(opentf.TF)
		dep:    setup_providers
		config: defs.vpc
		outputs: [{
			alias: "vpc_id"
			path:  ".module.vpc.aws_vpc.this[0].id"
		}, {
			alias: "public_subnet_ids"
			path:  ".module.vpc.aws_subnet.public[].id"
		}, {
			alias: "public_subnet_id"
			path:  ".module.vpc.aws_subnet.public[0].id"
		}, {
			alias: "private_subnet_ids"
			path:  ".module.vpc.aws_subnet.private[].id"
		}, {
			alias: "vpc_cidr_block"
			path:  ".module.vpc.aws_vpc.this[0].cidr_block"
		}, {
			alias: "nat_public_ips"
			path:  ".module.vpc.aws_eip.nat[].public_ip"
		}, {
			alias: "default_security_group_id"
			path:  ".module.vpc.aws_default_security_group.this[].id"
		}]
	}

	db_subnet_group: {
		@task(opentf.TF)
		dep:    setup_vpc
		config: defs.subnet_group
		outputs: [{
			alias: "subnet_group_ids"
			path:  ".aws_db_subnet_group.this[0].id"
		}]
	}

	setup_rds: {
		@task(opentf.TF)
		dep: [setup_vpc, db_subnet_group]
		config: defs.rds
		outputs: [{
			alias: "rds_endpoint"
			path:  ".aws_db_instance.this[0].endpoint"
		}]
	}

	setup_ec2: {
		@task(opentf.TF)
		dep:    setup_rds
		config: defs.ec2
	}
}
