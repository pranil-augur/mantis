package defs

rds: { 
    resource: {
        aws_db_instance: {
            "education_app": {
                source:                  "terraform-aws-modules/rds/aws"
                version:                 "~> 5.1"
                identifier:              "\(common.project_name)-db"
                engine:                  "postgres"
                engine_version:          "14.1"
                instance_class:          "db.t3.micro"
                allocated_storage:       5
                username:                "\(common.db_username)"
                password:                "\(common.db_password)"
                publicly_accessible:     true
                skip_final_snapshot:     true
                db_subnet_group_name: [string] | *null @var(subnet_ids)
            }
        }
    }   
}

