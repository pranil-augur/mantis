package main 

// Define EKS cluster configuration
#EKS_Cluster: {
    resource: {
        "aws_eks_cluster": {
            [string]: {
                cluster_name:                   string
                version:                        string
                enabled_cluster_log_types:      [...string]
                role_arn:                       string
                vpc_config: {
                    subnet_ids: [...string]
                }
            }
        }
    }
}

// Define VPC configuration
#VPC: {
    resource: {
        "aws_vpc": {
            [string]: {
                cidr_block: string
                tags: {
                    Name: string
                }
            }
        }
    }
}

// Define subnet configuration within the VPC
#Subnet: {
    resource: {
        "aws_subnet": {
            [string]: {
                vpc_id:     string
                cidr_block: string
                tags: {
                    Name: string
                }
            }
        }
    }
}

