package main

import (
    "augur.ai/static-website/defs"
)

tasks: {
    @flow(static_website_setup)

    setup_providers: {
        @task(opentf.TFProviders)
        config: defs.#providers
    }

    setup_vpc: {
        @task(opentf.TF)
        dep: setup_providers
        config: defs.vpc
        
        outputs: [{
            alias: "vpc_id"
            path: ".module.vpc.aws_vpc.this[0].id"

        }, {
            alias: "public_subnet_ids"
            path: ".module.vpc.aws_subnet.public[].id"
        }, {
            alias: "private_subnet_ids"
            path: ".module.vpc.aws_subnet.private[].id"
        }, {
            alias: "vpc_cidr_block"
            path: ".module.vpc.aws_vpc.this[0].cidr_block"
        }, {
            alias: "nat_public_ips"
            path: ".module.vpc.aws_eip.nat[].public_ip"
        },{
            alias: "default_security_group_id"
            path: ".module.vpc.aws_default_security_group.this[0].id"
        }]
    }

    setup_rds: {
        @task(opentf.TF)
        dep: setup_vpc
        inputs: {[
            {
                alias: "vpc_security_group_ids",
                value: ["tasks.setup_vpc.outputs.default_security_group_id"]
            },
            {
                alias: "private_subnet_ids",
                value: ["tasks.setup_vpc.outputs.private_subnet_ids"]
            }
        ]}
        config: defs.rds
        outputs: [{
            alias: "rds_endpoint"
            path: ["module", "rds", "db_instance_endpoint"]
        }]
    }

    setup_ec2: {
        @task(opentf.TF)
        dep: setup_rds
        inputs: {[
            {
                alias: "vpc_security_group_ids",
                value: ["tasks.setup_vpc.outputs.default_security_group_id"]
            },
            {
                alias: "public_subnet_id",
                value: ["tasks.setup_vpc.outputs.public_subnet_ids[0]"]
            },
            {
                alias: "rds_endpoint",
                value: ["tasks.setup_rds.outputs.rds_endpoint"]
            }
        ]}
        config: defs.ec2
    }
}