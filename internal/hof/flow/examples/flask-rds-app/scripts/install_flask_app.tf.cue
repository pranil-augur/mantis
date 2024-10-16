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
            path: ".data.aws_vpc.default.id"
            var:  "vpc_id"
        }]
    }

    get_availability_zones: {
        @task(mantis.core.TF)
        dep: [setup_tf_providers]
        config: data: aws_availability_zones: available: {
            state: "available"
        }
        exports: [{
            path: ".data.aws_availability_zones.available.names"
            var:  "az_names"
            exportAs: [string]
        }]
    }

     create_subnet_configs: {
        @task(mantis.core.Evaluate)
        dep: [get_default_vpc, get_availability_zones]
        exports: [{
            cueexpr: """
            import "net"
            subnet_configs: {
                subnet_az1: {
                    vpc_id: string @var(vpc_id)
                    cidr_block: "10.0.1.0/24"
                    availability_zone: string @arr(az_names, 0)
                }
                subnet_az2: {
                    vpc_id: string @var(vpc_id)
                    cidr_block: "10.0.2.0/24"
                    availability_zone: string @arr(az_names, 1)
                }
            }

            // Validate CIDR blocks
            _validateCIDR: {
                subnet_az1: net.IPCIDR & subnet_configs.subnet_az1.cidr_block
                subnet_az2: net.IPCIDR & subnet_configs.subnet_az2.cidr_block
            }

            // Output the validated subnet configurations
            result: subnet_configs
            """
            var: "subnet_configs"
        }]
    }

    create_subnets: {
        @task(mantis.core.TF)
        dep: [create_subnet_configs]
        config: resource: aws_subnet: _ | *null @var(subnet_configs)
        exports: [{
            path: ".aws_subnet.subnet_az1.id"
            var:  "subnet_az1_id"
        }, {
            path: ".aws_subnet.subnet_az2.id"
            var:  "subnet_az2_id"
        }]
    }

    setup_db_subnet_group: {
        @task(mantis.core.TF)
        dep: [create_subnets]
        config: resource: aws_db_subnet_group: default: {
            name:       "flask-rds-subnet-group"
            subnet_ids: [string] | *null @var(subnet_az1_id, subnet_az2_id)
            tags: Name: "Flask RDS subnet group"
        }
    }

    create_rds_security_group: {
        @task(mantis.core.TF)
        dep: [setup_db_subnet_group]
        config: resource: aws_security_group: rds_sg: {
            name:        "rds-security-group"
            description: "Security group for RDS instance"
            vpc_id:      string | *null @var(vpc_id)
            ingress: [{
                description:      "MySQL access"
                from_port:        3306
                to_port:          3306
                protocol:         "tcp"
                cidr_blocks:      ["0.0.0.0/0"]
                ipv6_cidr_blocks: []
                prefix_list_ids:  []
                security_groups:  []
                self:             false
            }]
            egress: [{
                description:      "Allow all outbound traffic"
                from_port:        0
                to_port:          0
                protocol:         "-1"
                cidr_blocks:      ["0.0.0.0/0"]
                ipv6_cidr_blocks: ["::/0"]
                prefix_list_ids:  []
                security_groups:  []
                self:             false
            }]
            tags: Name: "RDS Security Group"
        }
        exports: [{
            path: ".aws_security_group.rds_sg.id"
            var:  "rds_sg_id"
        }]
    }

    setup_rds: {
        @task(mantis.core.TF)
        dep: [setup_db_subnet_group, create_rds_security_group]
        config: resource: aws_db_instance: this: {
            identifier:        "flask-rds-instance"
            engine:            "mysql"
            engine_version:    "5.7"
            instance_class:    "db.t3.micro"
            allocated_storage: 20
            db_name:           "mydb"
            username:          "admin"
            password:          "change_this_password"
            db_subnet_group_name:   string | *null @var(aws_db_subnet_group.default.name)
            multi_az:               true
            vpc_security_group_ids: [string] | *null @var(rds_sg_id)
            deletion_protection:    false
            skip_final_snapshot:    true
            tags: Name: "Flask RDS Instance"
        }
        exports: [{
            path: ".aws_db_instance.this.endpoint"
            var:  "rds_endpoint"
        }, {
            path: ".aws_db_instance.this.port"
            var:  "rds_port"
        }]
    }

    deploy_flask_app: {
        @task(mantis.core.K8s)
        dep: [setup_rds]
        config: defs.flaskRdsDeployment 
    }
}