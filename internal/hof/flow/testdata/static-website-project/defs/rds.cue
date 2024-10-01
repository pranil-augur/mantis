package defs

import (
    "augur.ai/static-website/schemas"
)

rds: schemas.#ModuleRDS & {
    source:                  "terraform-aws-modules/rds/aws"
    version:                 "~> 5.1"
    identifier:              "\(common.project_name)-db"
    engine:                  "postgres"
    engine_version:          "14.1"
    instance_class:          "db.t3.micro"
    allocated_storage:       5
    username:                "\(common.db_username)"
    password:                "\(common.db_password)" // Assuming `common.db_password` is defined
    db_subnet_group_name:    "\(common.db_subnet_group_name)"
    vpc_security_group_ids: [string] [@runinject(vpc_security_group_ids)]
    parameter_group_name:    "\(common.db_parameter_group_name)"
    publicly_accessible:     true
    skip_final_snapshot:     true
}
