output "eks_node_role_arn" {
  value = aws_iam_role.eks_node.arn
}

output "lbc_role_arn" {
  value = aws_iam_role.lbc.arn
}
