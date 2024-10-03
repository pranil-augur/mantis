package defs

import (
	"augur.ai/static-website/schemas"
)

ec2: module: schemas.#ModuleEC2 & {
	source:        "terraform-aws-modules/ec2-instance/aws"
	version:       "~> 4.3"
	name:          "\(common.project_name)-instance"
	ami:           "ami-0c55b159cbfafe1f0" // Amazon Linux 2 AMI (HVM), SSD Volume Type
	instance_type: "t2.micro"
    subnet_id: string | *null @runinject(public_subnet_id)
	user_data: """
		#!/bin/bash
		amazon-linux-extras install docker
		service docker start
		usermod -a -G docker ec2-user
		
		# Save RDS endpoint to a file
		echo "@runinject(rds_endpoint)" > /home/ec2-user/rds_endpoint.txt
		
		# Run a container with the RDS endpoint as an environment variable
		docker run -d -p 80:80 -e RDS_ENDPOINT=$(cat /home/ec2-user/rds_endpoint.txt) nginx:latest
		"""
	tags: {
		Project:     common.project_name
		Environment: "dev"
	}
}
