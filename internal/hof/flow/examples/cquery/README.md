# Configuration Query Language

## Feature Status: Prototype

A declarative query language for CUE configurations that follows SQL-like semantics with path-based expressions.

## Query Structure 

```cue
{
    // FROM clause - specifies the data source path
    from: string
    // SELECT clause - what fields to retrieve
    select: [...string]
    // WHERE clause - predicates for filtering
    where: {
        [string]: _    // Key-value pairs for filtering
    }
}
```

## SELECT Expressions

### Basic Path Selection
```cue
from: "service"
select: [
    "_file",              // Select source file (system field)
    "name",              // Select service name
    "dependencies"       // Select dependencies
]
```

### Pattern-Based Selection
```cue
from: "service[string]"    // Select all service entries
select: [
    "_file",              // Include source file
    "name",              // Select all service names
    "dependencies"       // Select all dependencies
]
where: {
    dependencies: ["cache"]  // Filter by dependency
}
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


### Configuration Drift Detection

Configuration drift can occur when the actual configuration of resources deviates from the desired state. Using the Configuration Query Language, you can detect and address drift efficiently:

#### Identify Drift Across Environments
Run targeted queries on configurations in both staging and production to compare critical settings:

```cue
from: "service[string]"
select: [
    "name",
    "replicas",
    "env"
]
where: {
    "replicas": "3"   // Expected replica count
}
```
Run this query on both configurations to check for differences in replicas, environment variables, or other key settings.

#### Drift Detection Automation
Set up automated queries within CI/CD workflows to detect configuration drift:

```cue
from: "resource[string]"
select: [
    "name",
    "settings"
]
where: {
    "settings.deployment.region": "us-east-1"  // Match expected region
}
```
This query validates that resources remain in the intended region, detecting any unintentional drifts.

### Resource Import/Export

The Configuration Query Language facilitates the import and export of resource objects, allowing you to capture resource definitions from one environment and reapply them in another.

#### Exporting Resource Configurations
Use targeted queries to export specific resource configurations:

```cue
from: "resource[string]"
select: [
    "name",
    "type",
    "settings",
    "dependencies"
]
```
The query output can be saved in JSON or YAML, creating a versioned snapshot of the resource for:
- Reuse
- Documentation
- Backup

#### Importing Resource Configurations
1. Define a schema in CUE for importing configurations, ensuring compatibility with Mantis standards
2. Apply the exported configuration data to new or existing environments using `mantis apply`
3. Use for:
   - Cloning environments
   - Sharing configurations
   - Recovering from backups

#### Version Control and Synchronization
- Store exported configurations in Git to maintain a versioned history
- Enable rollback, cloning, or restoration of configurations
- Integrate with Mantis CI/CD pipelines to apply updated configurations
- Ensure environments stay synchronized with the exported state

These import/export capabilities enhance flexibility by allowing resources to be easily:
- Managed
- Replicated
- Shared across different environments

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

## Command Line Usage

### Querying Configurations
The `mantis query` command supports two modes of operation: natural language queries and CUE query configurations.

```bash
# Using natural language queries
mantis query \
  -S <system-prompt-path> \
  -C <code-dir> \
  -q "your natural language query" \
  -i <index-path>

# Using CUE query configuration
mantis query \
  -S <system-prompt-path> \
  -C <code-dir> \
  -c <query-config-path>
```

Options:
- `--system-prompt, -S`: Path to system prompt file (required)
- `--code-dir, -C`: Directory containing CUE configurations (required)
- `--query, -q`: Natural language query string
- `--index, -i`: Path to the query index file (required with -q)
- `--query-config, -c`: Path to CUE query configuration file

Examples:
```bash
# Natural language query
mantis query \
  -S ./prompts/cquery.txt \
  -C ./config \
  -q "Show me all services with more than 3 replicas" \
  -i ~/.mantis/index/mantis-index.json

# Using query config file
mantis query \
  -S ./prompts/cquery.txt \
  -C ./config \
  -c ./queries/replicas.cue
```

### Query Configuration File
When using the `-c` option, create a CUE file with your query:

```cue
// replicas.cue
{
    from: "service[string]"
    select: [
        "name",
        "replicas"
    ]
    where: {
        "replicas": "^[4-9]|[1-9][0-9]+$"  // Match 4 or more replicas
    }
}
```

### Natural Language Queries
When using `-q`, the query will be converted to CUE format automatically. Examples:

```bash
# Count replicas
mantis query -q "What is the total number of replicas across services"

# Find specific services
mantis query -q "Show me all frontend services"

# Resource queries
mantis query -q "Find services with high CPU limits"
```

The natural language query mode requires:
1. A system prompt file (-S) that guides the AI
2. An index file (-i) containing sample queries
3. The code directory to search (-C)

### Building Query Index
The `mantis index` command builds a search index for CUE files to optimize query performance and generate sample queries.

```bash
mantis index --code-dir <path-to-cue-files> [options]
```

Options:
- `--code-dir, -C`: Directory containing CUE configurations (required)
- `--system-prompt, -S`: Path to system prompt file for AI-powered query generation
- `--index-dir, -i`: Directory for storing the index (defaults to ~/.mantis/index)

Example:
```bash
# Basic indexing
mantis index --code-dir ./configs

# Custom index location
mantis index --code-dir ./configs --index-dir /path/to/index

# With custom system prompt
mantis index --code-dir ./configs --system-prompt ./prompts/index.txt
```

The index command:
1. Scans the specified directory for CUE configurations
2. Generates sample queries based on the configurations
3. Stores the index in the specified directory (default: ~/.mantis/index/mantis-index.json)
4. Uses AI to suggest relevant queries based on your configurations

### Index Structure
The generated index contains:
- Sample queries for common use cases
- Pre-computed paths for faster querying
- Query templates for common operational scenarios

This index is used by the query command to provide:
- Query suggestions
- Faster query execution
- Common operational queries