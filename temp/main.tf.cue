package main


import "augur"

// Common configuration values
common: {
    cluster_name: "eks-workshop"
	vpc_cidr: "10.42.0.0/16"
}

// Define generic structures for Terraform modules
#ModuleEKSAddons: {
    source:  string
    version: string
    enable_aws_load_balancer_controller: bool
    aws_load_balancer_controller: {
        wait: bool
    }
    cluster_name:      string
    cluster_endpoint:  string
    cluster_version:   string
    oidc_provider_arn: string
}

#ModuleEKS: {
    source:  string
    version: string
    cluster_name:                   string
    cluster_version:                string
    cluster_endpoint_public_access: bool
    cluster_addons: {
        "vpc-cni": {
            before_compute:       bool
            most_recent:          bool
            configuration_values: string
        }
    }
    vpc_id:     string
    subnet_ids: [...string]
    create_cluster_security_group: bool
    create_node_security_group:    bool
    eks_managed_node_groups: {
        "default": {
            instance_types:       [string]
            force_update_version: bool
            release_version:      string
            min_size:             int
            max_size:             int
            desired_size:         int
            labels: {
                "workshop-default": string
            }
        }
    }
    tags: [string]:string
}

#ModuleVPC: {
    source:  string
    version: string
    name:                    string
    cidr:                    string
    azs:                     [...string]
    public_subnets:          [...string]
    private_subnets:         [...string]
    public_subnet_suffix:    string
    private_subnet_suffix:   string
    enable_nat_gateway:      bool
    create_igw:              bool
    enable_dns_hostnames:    bool
    single_nat_gateway:      bool
    manage_default_network_acl:    bool
    default_network_acl_tags:      {Name: string}
    manage_default_route_table:    bool
    default_route_table_tags:      {Name: string}
    manage_default_security_group: bool
    default_security_group_tags:   {Name: string}
    public_subnet_tags:      [string]:string
    private_subnet_tags:     [string]:string
    tags:                    [string]:string
}

// Create instances of the modules with actual values
eksAddons: #ModuleEKSAddons & {
    source:  "aws-ia/eks-blueprints-addons/aws"
    version: "1.9.2"
    enable_aws_load_balancer_controller: true
    aws_load_balancer_controller: {
        wait: true
    }
    cluster_name:      "PLACEHOLDER"  // Dynamically linked to EKS module
    cluster_endpoint:  "PLACEHOLDER"
    cluster_version:   "PLACEHOLDER"
    oidc_provider_arn: "PLACEHOLDER"
}

eks: #ModuleEKS & {
    source:  "terraform-aws-modules/eks/aws"
    version: "~> 19.16"
    cluster_name:                   "eks-workshop"
    cluster_version:                "1.29"
    cluster_endpoint_public_access: true
    cluster_addons: {
        "vpc-cni": {
            before_compute: true
            most_recent:    true
			configuration_values: """
				{
					"env": {
						"ENABLE_POD_ENI": "true",
						"ENABLE_PREFIX_DELEGATION": "true",
						"POD_SECURITY_GROUP_ENFORCING_MODE": "standard"
					},
					"enableNetworkPolicy": "true"
				}
				"""
        }
    }
    vpc_id:     "vpc-123"
    subnet_ids: ["subnet-456", "subnet-789"]
    create_cluster_security_group: false
    create_node_security_group:    false
    eks_managed_node_groups: {
        "default": {
            instance_types:       ["t3.large"]
            force_update_version: true
            release_version:      "1.29.0-20240129"
            min_size:             1
            max_size:             1
            desired_size:         1
            labels: {
                "workshop-default": "yes"
            }
        }
    }
    tags: {
        "karpenter.sh/discovery": "eks-workshop"
    }
}

locals: {
    // Slice the first two availability zones
    azs: ["us-east-1a", "us-east-1b"] 

    // Generate private subnets CIDRs
    private_subnets: [for index, _ in locals.azs { augur.CidrSubnet("10.42.0.0/16", 3, index + 3) }]

    // Generate public subnets CIDRs
    public_subnets: [for index, _ in locals.azs { augur.CidrSubnet("10.42.0.0/16", 3, index) }]
}

vpc: #ModuleVPC & {
    source:  "terraform-aws-modules/vpc/aws"
    version: "~> 5.1"
    name:    "eks-workshop"
    cidr:    "10.42.0.0/16"
    azs:     locals.azs
    public_subnets: locals.public_subnets
    private_subnets: locals.private_subnets
    public_subnet_suffix: "SubnetPublic"
    private_subnet_suffix: "SubnetPrivate"
    enable_nat_gateway: true
    create_igw: true
    enable_dns_hostnames: true
    single_nat_gateway: true
    manage_default_network_acl: true
    default_network_acl_tags: {Name: "\(common.cluster_name)-default"}
    manage_default_route_table: true
    default_route_table_tags: {Name: "\(common.cluster_name)-default"}
    manage_default_security_group: true
    default_security_group_tags: {Name: "\(common.cluster_name)-default"}
    public_subnet_tags: {"kubernetes.io/role/elb": "1"}
    private_subnet_tags: {"karpenter.sh/discovery": "eks-workshop"}
    tags: {
        "created-by": "eks-workshop-v2",
        "env": "eks-workshop"
    }
}


// Final blueprint combining all resources
cueform: {
	"data": {
        "aws_availability_zones": {
            "available": [
                {
                    "state": "available"
                }
            ]
        }
    },
    module: {
        "eks_blueprints_addons": eksAddons
        "eks": eks
        "vpc": vpc
    }
}
