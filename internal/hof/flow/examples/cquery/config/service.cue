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

#HealthCheck: {
    path: string
    interval: string
    timeout: string
}
