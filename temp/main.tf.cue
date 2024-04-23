package main

#VPCConfig: {
    cidr_block: string
}

#SubnetConfig: {
    vpc_id:     string
    cidr_block: string
    tags:       [string]: string | *{}
}

// Example VPC and subnet configurations
exampleVPC: {
    cidrBlock: "10.0.0.0/16"
    subnets: {
        "subnet-01": { cidr: "10.0.1.0/24" },
        "subnet-02": { cidr: "10.0.2.0/24" }
    }
}

vpcResources: {
    "aws_vpc": {
        "example": {
            cidr_block: exampleVPC.cidrBlock
        }
    },
    "aws_subnet": {
        for name, subnet in exampleVPC.subnets {
            "\(name)": {
                vpc_id:     "aws_vpc.example.id",
                cidr_block: subnet.cidr,
                tags:       subnet.tags | *{}
            }
        }
    }
}

#EKSClusterConfig: {
    cluster_name: string
    version:      string
    subnet_ids:   [...string]
    depends_on:   [...string]
}

// Adjusted Example EKS cluster configuration
exampleCluster: {
    name:    "example-eks-cluster", // Corrected to 'name' from 'cluster_name'
    version: "1.29",
    vpcConfig: {
        subnets: ["subnet-12345678", "subnet-87654321"]
    },
    role_arn: "arn:aws:iam::123456789012:role/EKSRole" // Add your actual IAM role ARN here
}

// Adjusted EKS resources definition
eksResources: {
    "aws_eks_cluster": {
        "example": {
            name:    exampleCluster.name, // Corrected to 'name' from 'cluster_name'
            version: exampleCluster.version,
            role_arn: exampleCluster.role_arn, // Added role_arn
            vpc_config: {
                subnet_ids: exampleCluster.vpcConfig.subnets
            },
            depends_on: ["aws_vpc.example"]
        }
    }
}

// Merge all resources into a single blueprint
cueform: {
    resource: vpcResources & eksResources
}

