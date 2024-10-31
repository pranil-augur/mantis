// Selection expressions
select: [
    // "service",                    // Select all services
    "service.web",               // Select web service
    // "service[string].name",      // Select all service names
]

// Predicate conditions (WHERE clause)
where: {
    name: "web-frontend"    // WHERE name = "web-frontend"
    // "name": "^web-*"
    // "web.name": "^web-frontend"        // WHERE web.name MATCHES '^web-.*'
}
