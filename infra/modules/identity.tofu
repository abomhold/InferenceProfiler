variable "module_name" { default = "" }
variable "role_name" { default = "" }
variable "profile_name" { default = "" }
variable "tags" { default = {} }

# Skip if not identity module
locals {
  is_identity = var.module_name == "identity"
}

resource "aws_iam_role" "ecs_instance" {
  count = local.is_identity ? 1 : 0
  name  = var.role_name
  tags  = var.tags

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs" {
  count      = local.is_identity ? 1 : 0
  role       = aws_iam_role.ecs_instance[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_role_policy_attachment" "ssm" {
  count      = local.is_identity ? 1 : 0
  role       = aws_iam_role.ecs_instance[0].name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_role_policy_attachment" "cloudwatch" {
  count      = local.is_identity ? 1 : 0
  role       = aws_iam_role.ecs_instance[0].name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
}

resource "aws_iam_instance_profile" "ecs_instance" {
  count = local.is_identity ? 1 : 0
  name  = var.profile_name
  role  = aws_iam_role.ecs_instance[0].name
}

output "role_arn" {
  value = local.is_identity ? aws_iam_role.ecs_instance[0].arn : ""
}

output "role_name" {
  value = local.is_identity ? aws_iam_role.ecs_instance[0].name : ""
}

output "instance_profile_arn" {
  value = local.is_identity ? aws_iam_instance_profile.ecs_instance[0].arn : ""
}