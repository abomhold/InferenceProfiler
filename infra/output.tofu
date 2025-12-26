output "cluster" {
  value = {
    name              = module.ecs_cluster.cluster_name
    arn               = module.ecs_cluster.cluster_arn
    capacity_provider = module.ecs_cluster.capacity_provider_name
  }
}

output "iam" {
  value = {
    role_arn    = module.identity.role_arn
    profile_arn = module.identity.instance_profile_arn
  }
}

output "networking" {
  value = {
    vpc_id            = module.networking.vpc_id
    subnet_ids        = module.networking.subnet_ids
    security_group_id = module.networking.security_group_id
  }
}

output "local_public_ip" {
  description = "Detected local public IP with full access"
  value       = local.local_public_ip
}