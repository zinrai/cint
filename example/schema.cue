package example

// Application configuration schema
#Config: {
    // Application name (required, lowercase with hyphens)
    name: string & =~"^[a-z][a-z0-9-]*$"
    
    // Version following semantic versioning (required)
    version: string & =~"^v?[0-9]+\\.[0-9]+\\.[0-9]+$"
    
    // Number of replicas (optional, 1-10)
    replicas?: int & >=1 & <=10
    
    // Deployment environment (required)
    environment: "development" | "staging" | "production"
    
    // Resource limits (optional)
    resources?: {
        // CPU limit (e.g., "100m", "2")
        cpu: string & =~"^[0-9]+m?$"
        // Memory limit (e.g., "256Mi", "1Gi")
        memory: string & =~"^[0-9]+[MG]i?$"
    }
    
    // Service ports (optional)
    ports?: [...{
        name?: string
        port!: int & >0 & <=65535
        protocol?: "TCP" | "UDP" | *"TCP"
    }]
    
    // Health check configuration (optional)
    healthCheck?: {
        enabled?: bool | *true
        path?: string | *"/health"
        interval?: string & =~"^[0-9]+s$" | *"30s"
        timeout?: string & =~"^[0-9]+s$" | *"5s"
    }
    
    // Labels for organizing (optional)
    labels?: {[string]: string}
}
