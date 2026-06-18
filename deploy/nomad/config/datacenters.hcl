datacenter "functionfly" {
  # Primary datacenter for FunctionFly workloads
}

datacenter "us-west-2" {
  # US West region for AI Gateway
}

datacenter "eu-west-1" {
  # EU region for AI Gateway
}

node_pool "functionfly" {
  description = "Default node pool for FunctionFly services"
}

node_pool "gpu" {
  description = "GPU-enabled nodes for AI workloads"
}

node_pool "workers" {
  description = "General worker nodes"
}