output "cluster_name"    { value = module.eks.cluster_name }
output "rds_endpoint"    { value = module.rds.endpoint }
output "ecr_urls"        { value = module.ecr.repository_urls }
