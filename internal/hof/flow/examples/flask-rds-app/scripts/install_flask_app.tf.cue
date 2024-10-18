package main

import (
	"augur.ai/rds-flask-app/defs"
)

deploy_flask_rds: {
	@flow(deploy_flask_rds)

	setup_tf_providers: {
		@task(mantis.core.TF)
		config: defs.#providers
	}

	get_default_vpc: {
		@task(mantis.core.TF)
		dep: [setup_tf_providers]
		config: data: aws_vpc: default: {
			default: true
		}
		exports: [{
			jqpath: ".data.aws_vpc.default.id"
			var:  "vpc_id"
		}, {
			jqpath: ".data.aws_vpc.default.cidr_block"
			var:  "vpc_cidr_block"
		}, {
			jqpath: ".data.aws_vpc.default.id"
			var:  "vpc_id_set"
			as: [string]
        },{
            jqpath: ".data.aws_vpc.default.cidr_block"
			var:  "vpc_cidr_block_set"
			as: [string]
        }]
	}

	get_subnets: {
		@task(mantis.core.TF)
		dep: [get_default_vpc]
		config: data: aws_subnets: default: {
			filter: [{
				name: "vpc-id"
				values: [...string] | *null @var(vpc_id_set)
			}]
		}
		exports: [{
			jqpath: ".data.aws_subnets.default.ids"
			var:  "subnet_ids"
		}]
	}

	select_subnets: {
		@task(mantis.core.Eval)
		dep: [get_subnets]
		cueexpr: """
			subnet_ids: @var(subnet_ids)
			
			result: {
			    subnet_az1_id: subnet_ids[0]
			    subnet_az2_id: subnet_ids[1]
			}
			"""
		exports: [{
			var: "selected_subnets"
			jqpath: "."
		}]
	}

	create_subnets: {
		@task(mantis.core.TF)
		dep: [select_subnets]
		config: data: aws_subnet: {
			subnet_az1: {
				id: string | *null @var(selected_subnets.subnet_az1_id)
			}
			subnet_az2: {
				id: string | *null @var(selected_subnets.subnet_az2_id)
			}
		}
		exports: [{
			jqpath: ".data.aws_subnet.subnet_az1.id"
			var:  "subnet_az1_id"
		}, {
			jqpath: ".data.aws_subnet.subnet_az2.id"
			var:  "subnet_az2_id"
		},{
            jqpath: ".data.aws_subnet[].id",
            var:"subnet_ids"
        }]
	}

	setup_db_subnet_group: {
		@task(mantis.core.TF)
		dep: [create_subnets]
		config: resource: aws_db_subnet_group: default: {
			name: "flask-rds-subnet-group-2"
			subnet_ids: [...string] | *null @var(subnet_ids)
			tags: Name: "Flask RDS subnet group"
		}
	}

	create_rds_security_group: {
		@task(mantis.core.TF)
		dep: [setup_db_subnet_group]
		config: resource: aws_security_group: rds_sg: {
			name:        "rds-security-group-agr-1237"
			description: "Security group for RDS instance"
			vpc_id:      string | *null @var(vpc_id)
			ingress: [{
				description: "MySQL access"
				from_port:   3306
				to_port:     3306
				protocol:    "tcp"
				cidr_blocks: [string] | *null @var(vpc_cidr_block_set)
				ipv6_cidr_blocks: []
				prefix_list_ids: []
				security_groups: []
				self: false
			}]
			egress: [{
				description: "Allow all outbound traffic"
				from_port:   0
				to_port:     0
				protocol:    "-1"
				cidr_blocks: ["0.0.0.0/0"]
				ipv6_cidr_blocks: ["::/0"]
				prefix_list_ids: []
				security_groups: []
				self: false
			}]
			tags: Name: "RDS Security Group"
		}
		exports: [{
			jqpath: ".aws_security_group.rds_sg.id"
			var:  "rds_sg_id"
            as: [string]
		}]
	}

	setup_rds: {
		@task(mantis.core.TF)
		dep: [setup_db_subnet_group, create_rds_security_group]
		config: resource: aws_db_instance: this: {
			identifier:           "flask-rds-instance"
			engine:               "mysql"
			engine_version:       "5.7"
			instance_class:       "db.t3.micro"
			allocated_storage:    20
			db_name:              "mydb"
			username:             "admin"
			password:             "change_this_password"
			db_subnet_group_name: string | *null @var(aws_db_subnet_group.default.name)
			multi_az:             true
			vpc_security_group_ids: [...string] | *null @var(rds_sg_id)
			deletion_protection: false
			skip_final_snapshot: true
			tags: Name: "Flask RDS Instance"
		}
		exports: [{
			jqpath: ".aws_db_instance.this.endpoint"
			var:  "rds_endpoint"
		}, {
			jqpath: ".aws_db_instance.this.port"
			var:  "rds_port"
		}]
	}

	deploy_flask_app: {
		@task(mantis.core.K8s)
		dep: [setup_rds]
		config: defs.flaskRdsDeployment
	}
}
