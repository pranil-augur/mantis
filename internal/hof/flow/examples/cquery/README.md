# Mantis Query Language

A declarative query language for CUE configurations that follows SQL-like semantics with path-based expressions.

## Query Structure 

```cue
query: {
    // FROM clause - specifies the data source path
    from: string
    // SELECT clause - what to retrieve
    select: [...string]
    // WHERE clause - predicates for filtering
    where: [string]: string
}
```

## SELECT Expressions

### Basic Path Selection
```cue
from: "service"
select: [
    "*",                  // Select entire service struct
    "web",               // Select specific service
    "web.port"          // Select specific field
]
```

### Pattern-Based Selection
```cue
from: "service[string]"    // Select all service entries
select: [
    "name",               // Select all service names
    "env",                // Select all service environments
    "port"                // Select all ports
]
```

Pattern Types:
- `[string]`: Match string values
- `[int]`: Match integer values
- `[float]`: Match float values
- `[number]`: Match any numeric value
- `[bool]`: Match boolean values
- `[struct]`: Match struct values
- `[list]`: Match list values
- `[_]` or `[any]`: Match any type

## WHERE Predicates

### Path-Based Predicates
```cue
where: {
    "name": "web-frontend"           // Exact match
    "web.name": "^web-.*"            // Regex match with path
    "service.web.port": "8080"       // Deep path match
}
```

### Regular Expressions
Uses RE2 syntax for pattern matching:
```cue
where: {
    "name": "^web-.*"                // Starts with "web-"
    "env.LOG_LEVEL": "^(info|debug)$" // Enum values
    "env.API_KEY": "^[A-Za-z0-9]+$"   // Pattern matching
}
```

## Examples

### Basic Query
```cue
from: "service[string]"
select: [
    "name"
]
where: {
    name: "web-*"   // Match services starting with "web-"
}
```

### Deep Path Query
```cue
from: "service[string]"
select: [
    "env"
]
where: {
    "env.LOG_LEVEL": "^(info|debug)$"   // Match services with specific log levels
}
```

### Multiple Field Selection
```cue
from: "service[string]"
select: [
    "name",
    "replicas",
    "env"
]
where: {
    "replicas": "3"   // Match services with 3 replicas
}
```

### Wildcard Selection
```cue
from: "service[string]"
select: [
    "*"    // Select all fields
]
where: {
    "name": "api-*"   // Match services starting with "api-"
}
```

## Path Resolution Rules
1. Absolute paths start from root
2. Relative paths are based on selected context
3. Pattern matches apply to immediate children
4. WHERE predicates use full paths from query context

## Type System
- Preserves CUE's type system
- Pattern matching respects CUE types
- Regular expressions for string values
- Exact matching for other types

## Limitations
1. Single-level pattern matching only
2. String-based predicate values
3. No complex boolean operations in WHERE
4. No aggregations or grouping

## Example Configuration

### Service Definition (service.cue)
```cue
package services

service: {
    web: {
        name: "web-frontend"
        port: 8080
        replicas: 3
        env: {
            POSTGRES_HOST: "db.example.com"
            API_KEY: string & =~"^[A-Za-z0-9]{32}$"
        }
    }
    
    api: {
        name: "backend-api"
        port: 3000
        replicas: 2
        env: {
            REDIS_URL: "redis://cache:6379"
            LOG_LEVEL: "info" | "debug" | "error"
        }
    }
}
```

### Query Definition (query.cue)
```cue
// Selection expressions
select: [
    "service.web",               // Select web service
]

// Predicate conditions (WHERE clause)
where: {
    name: "web-frontend"    // WHERE name = "web-frontend"
}
```

## Operational Query Examples

The CUE Query Language can help answer critical operational questions about your infrastructure. Here are some examples:

### Container Image Security
```cue
from: "service[string]"
select: [
    "*"
]
where: {
    "image": "^vulnerable-.*"
    // OR
    "image.tag": "^.*-cve-.*"
}
```

### Resource Consumption
```cue
from: "service[string]"
select: [
    "*"
]
where: {
    "resources.requests.cpu": "^[1-9][0-9]*$"     // CPU > 1
    "resources.requests.memory": "^[1-9]Gi$"      // Memory > 1Gi
}
```

### Resource Configuration
Find potential resource misconfigurations:
```cue
from: "service[string]"
select: [
    "resources"
]
where: {
    "resources.requests.cpu": "^[0-9]*m$"         // CPU in millicores
    "resources.limits.cpu": "^[1-9][0-9]*$"       // High CPU limit
}
```

### High Availability
Check replica placement:
```cue
from: "service[string]"
select: [
    "topology"
]
where: {
    "node": ".*-zone-a"                          // Check zone placement
    "replicas": "^[1-9][0-9]*$"                  // Multiple replicas
}
```

### Image Tracking
Track specific images across clusters:
```cue
from: "service[string]"
select: [
    "image"
]
where: {
    "image.repository": "^nginx.*"
    "image.tag": "1.19.*"
}
```

### Resource Management
Track high resource consumers by namespace:
```cue
from: "service[string]"
select: [
    "resources",
    "namespace"
]
where: {
    "namespace": "production"
    "resources.requests.memory": "^[1-9][0-9]*Gi$"
}
```

### Application Ownership
Find resources by owner:
```cue
from: "service[string]"
select: [
    "metadata"
]
where: {
    "metadata.labels.app": "frontend"
    "metadata.labels.team": "platform"
}
```

### Change History
Track configuration changes:
```cue
from: "service[string]"
select: [
    "history"
]
where: {
    "history.timestamp": "^2024-03.*"
    "history.type": "ConfigChange"
}
```

### Example Extended Configuration
To support these operational queries, your CUE configuration should include operational metadata:

```cue
package services

service: {
    web: {
        name: "web-frontend"
        image: {
            repository: "nginx"
            tag: "1.19.0"
            vulnerabilities: []
        }
        resources: {
            requests: {
                cpu: "500m"
                memory: "512Mi"
            }
            limits: {
                cpu: "1000m"
                memory: "1Gi"
            }
        }
        topology: {
            zone: "us-east-1a"
            node: "node-1"
            replicas: 3
        }
        metadata: {
            labels: {
                app: "frontend"
                team: "platform"
            }
        }
        history: [{
            timestamp: "2024-03-14T12:00:00Z"
            type: "ConfigChange"
            change: "Updated resource limits"
        }]
    }
}
```

### Common Operational Questions Answered
- Are any vulnerable container images running across clusters?
- Which deployments or workloads are consuming the most CPU and memory?
- Are resource requests configured too low, potentially causing CPU throttling?
- Are replicas of the same service deployed on the same node and availability zone?
- Where are specific container images or tags running across clusters?
- Which clusters and namespaces host deployments with high resource consumption?
- Which applications manage the resources running a specific container image?
- What is the history of changes and events affecting specific workloads?

### Key Benefits for Operations
1. Path-based navigation of complex configurations
2. Regular expression support for flexible matching
3. Type-safe querying
4. Hierarchical data representation
5. Cross-cluster configuration analysis

## Change Impact Analysis

The CUE Query Language can help assess the blast radius of configuration changes across cloud and Kubernetes resources.

### Dependency Tracking
Find all resources depending on a specific component:
```cue
from: "resource[string]"
select: [
    "*"
]
where: {
    "depends_on": ".*"
}
```

### Cross-Resource References
Track resource references across configurations:
```cue
from: "resource[string]"
select: [
    "references"
]
where: {
    "type": "SecurityGroup"
    "id": "sg-.*"
}
```

### Service Mesh Impact
Analyze service mesh dependencies:
```cue
from: "service[string]"
select: [
    "ingress",
    "egress"
]
where: {
    "target.service": "payment-api"  // Find all services communicating with payment-api
}
```

### Infrastructure Dependencies
```cue
from: "resource[string]"
select: [
    "*"
]
where: {
    "provider": "aws"
    "type": "subnet|vpc|security_group"
    "used_by": ".*frontend.*"
}
```

### Example Configuration with Dependencies
```cue
resource: {
    "frontend-app": {
        type: "kubernetes_deployment"
        name: "frontend"
        depends_on: ["redis-cache", "postgres-db"]
        references: [{
            type: "SecurityGroup"
            id: "sg-123456"
        }]
        ingress: [{
            from: "api-gateway"
            port: 8080
        }]
        egress: [{
            to: "payment-api"
            port: 3000
        }]
        infrastructure: {
            vpc: "vpc-123456"
            subnet: "subnet-123456"
            security_groups: ["sg-123456"]
        }
    }
}
```

### Common Change Impact Questions
1. **Direct Dependencies**
```cue
// What directly depends on this component?
from: "resource[string]"
select: [
    "*"
]
where: {
    "depends_on": ".*redis-cache.*"
}
```

2. **Network Impact**
```cue
// What services might be affected by network changes?
from: "resource[string]"
select: [
    "*"
]
where: {
    "infrastructure.vpc": "vpc-123456"
    "type": "kubernetes_service|aws_lb"
}
```

3. **Security Impact**
```cue
// What resources share security groups?
from: "resource[string]"
select: [
    "*"
]
where: {
    "infrastructure.security_groups": ".*sg-123456.*"
}
```

4. **Service Chain Impact**
```cue
// Trace service chain dependencies
from: "service[string]"
select: [
    "(ingress|egress)"
]
where: {
    "target.service": "payment-api"
    "type": "http|grpc"
}
```

### Change Risk Assessment
```cue
// Find high-risk changes
from: "resource[string]"
select: [
    "*"
]
where: {
    "criticality": "high"
    "dependencies_count": "^[5-9][0-9]*$"  // More than 50 dependencies
}
```

### Benefits for Change Management
1. **Dependency Visualization**
   - Track direct and indirect dependencies
   - Identify critical paths
   - Map resource relationships

2. **Risk Assessment**
   - Identify high-impact changes
   - Evaluate dependency chains
   - Assess security implications

3. **Change Planning**
   - Plan maintenance windows
   - Coordinate cross-team changes
   - Validate change safety

4. **Compliance Verification**
   - Track security group changes
   - Monitor network modifications
   - Validate configuration compliance

### Best Practices
1. Always include dependency metadata in resource definitions
2. Use consistent naming patterns for resources
3. Tag resources with criticality and ownership
4. Document service dependencies explicitly
5. Include infrastructure references in service definitions
```

This demonstrates how the query language can help:
- Trace resource dependencies
- Assess change impact
- Plan maintenance
- Validate changes
- Ensure compliance

Would you like me to elaborate on any specific aspect of change impact analysis?