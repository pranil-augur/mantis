package defs

import (
	"augur.ai/static-website/schemas"
)

vpc: {
	module: {
		vpc: schemas.#ModuleVPC & {
			source:               "terraform-aws-modules/vpc/aws"
			version:              "~> 5.1"
			name:                 common.project_name
			cidr:                 common.vpc_cidr
			azs:                  locals.azs
			private_subnets:      locals.private_subnets
			public_subnets:       locals.public_subnets
			enable_nat_gateway:   true
			single_nat_gateway:   true
			enable_dns_hostnames: true
			tags: {
				Project:     common.project_name
				Environment: "dev"
			}
		}
	}
}

subnet_group: {
	resource: {
		aws_db_subnet_group: education: {
			name: "education"
			subnet_ids: [...string] | *null @var(public_subnet_ids)
			tags: {
				Name: "Education"
			}
		}
	}
}
