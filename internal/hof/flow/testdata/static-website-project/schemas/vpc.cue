package schemas

#ModuleVPC: { // actually resourceVPC
    source:  string
    version: string
    name:    string
    cidr:    string
    azs:     [...string]
    public_subnets:  [...string]
    private_subnets: [...string]
    enable_nat_gateway:   bool
    single_nat_gateway:   bool
    enable_dns_hostnames: bool
    tags:                 [string]:string
}

//
// module -> 
// resource -> 
//