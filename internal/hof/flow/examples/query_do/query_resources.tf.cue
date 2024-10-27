package main

// Optional: Define schemas for filters
#DropletFilters: {
    name?:    string
    region?:  string
    tag?:     string
    status?:  "active" | "off" | "archive"
    size?:    string
}

#VolumeFilters: {
    name?:    string
    region?:  string
    size?:    string
    tag?:     string
}

common: {
    digitalocean_token: "dopxyz"
}

query_digitalocean_resources: {
    @flow(query_digitalocean_resources)
    
    // Query Droplets with filters
    get_droplets: {
        @task(mantis.cloudprovider.DigitalOcean)
        config: {
            token: "\(common.digitalocean_token)"
            resourceType: "Droplet"
            filters: #DropletFilters & {
                name: "debian-s-4vcpu-8gb-nyc3-01"
                region: "nyc3"
                size: "s-2vcpu-8gb-160gb-intel"
                status: "active"
            }
        }
        exports: [{
            // jqpath: ".droplets[] | {name: .name, id: .id, status: .status, region: .region.slug, size: .size_slug}"
            jqpath: ".droplets[0]"
            var:  "do_droplets"
        }]
    }

    print_resources: {
        @task(os.Stdout)
        dep: [get_droplets]
        // text: string | *null @var(do_droplets.name)
        text: string | *null @var(do_droplets.name)
    }
}
