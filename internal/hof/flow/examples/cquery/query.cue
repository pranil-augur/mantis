// CUE path expressions to evaluate
expressions: [
    // Basic service queries
    // "service",                    // Get all services
    "service.web",               // Get web service
    // "service.api",               // Get api service
    
    // Specific field queries
    // "service.web.port",          // Get web service port
    // "service.api.port",          // Get api service port
    
    // Pattern-based queries
    // "service[string].name",      // Get all service names
    // "service[string].port",      // Get all service ports
    // "service[string].replicas",
]

// Remove or modify filters since we don't have environment in our service.cue
filters: {}
