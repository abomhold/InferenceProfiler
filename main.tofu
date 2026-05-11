variable "REGION" { type = string }
variable "PROJECT" { type = string }
variable "SERVER_INSTANCE_TYPE" { type = string }
variable "CLIENT_INSTANCE_TYPE" { type = string }
variable "AVAILABILITY_ZONE" { type = number }
variable "PROFILE" { type = string }

terraform {
  required_version = ">= 1.6"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region  = var.REGION
  profile = var.PROFILE
}

locals {
  server_is_arm = can(regex("^[a-z][0-9]g", var.SERVER_INSTANCE_TYPE))
  client_is_arm = can(regex("^[a-z][0-9]g", var.CLIENT_INSTANCE_TYPE))
  server_ami_id = local.server_is_arm ? data.aws_ami.ubuntu_arm.id : data.aws_ami.ubuntu_x86.id
  client_ami_id = local.client_is_arm ? data.aws_ami.ubuntu_arm.id : data.aws_ami.ubuntu_x86.id
}

# --- SSH ---

resource "tls_private_key" "ssh" { algorithm = "ED25519" }
resource "aws_key_pair" "ssh" {
  key_name   = "${var.PROJECT}-key"
  public_key = tls_private_key.ssh.public_key_openssh
}
resource "local_file" "ssh_key" {
  content         = tls_private_key.ssh.private_key_openssh
  filename        = "${path.root}/../${var.PROJECT}.pem"
  file_permission = "0600"
}

# --- AMIs ---

data "aws_ami" "ubuntu_x86" {
  most_recent = true
  owners = ["099720109477"]
  filter {
    name = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }
  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }
  filter {
    name = "architecture"
    values = ["x86_64"]
  }
}

data "aws_ami" "ubuntu_arm" {
  most_recent = true
  owners = ["099720109477"]
  filter {
    name = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-arm64-server-*"]
  }
  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }
  filter {
    name = "architecture"
    values = ["arm64"]
  }
}

# --- Networking ---

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
  filter {
    name = "default-for-az"
    values = ["true"]
  }
}

data "aws_subnet" "selected" {
  id = data.aws_subnets.default.ids[var.AVAILABILITY_ZONE]
}

data "http" "my_ip" {
  url = "https://checkip.amazonaws.com"
}

resource "aws_security_group" "nodes" {
  name   = "${var.PROJECT}-nodes"
  vpc_id = data.aws_vpc.default.id
  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    cidr_blocks = ["${trimspace(data.http.my_ip.response_body)}/32"]
  }
  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }
  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# --- IAM ---

resource "aws_iam_role" "node" {
  name = "${var.PROJECT}-node"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = { Service = "ec2.amazonaws.com" }
      }
    ]
  })
}

resource "aws_iam_instance_profile" "node" {
  name = "${var.PROJECT}-node"
  role = aws_iam_role.node.name
}

# --- Server ---

resource "aws_instance" "server" {
  ami                  = local.server_ami_id
  instance_type        = var.SERVER_INSTANCE_TYPE
  subnet_id            = data.aws_subnet.selected.id
  vpc_security_group_ids = [aws_security_group.nodes.id]
  key_name             = aws_key_pair.ssh.key_name
  iam_instance_profile = aws_iam_instance_profile.node.name

  instance_market_options {
    market_type = "spot"
    spot_options {
      spot_instance_type             = "one-time"
      instance_interruption_behavior = "terminate"
    }
  }

  user_data = file("${path.root}/configs/cloud-init/server.yaml")

  root_block_device {
    volume_size = 150
    volume_type = "gp3"
    iops        = 3000
    throughput  = 125
    encrypted   = true
  }
}

# --- Client ---

resource "aws_instance" "client" {
  ami                  = local.client_ami_id
  instance_type        = var.CLIENT_INSTANCE_TYPE
  subnet_id            = data.aws_subnet.selected.id
  vpc_security_group_ids = [aws_security_group.nodes.id]
  key_name             = aws_key_pair.ssh.key_name
  iam_instance_profile = aws_iam_instance_profile.node.name

  instance_market_options {
    market_type = "spot"
    spot_options {
      spot_instance_type             = "persistent"
      instance_interruption_behavior = "stop"
    }
  }

  user_data = file("${path.root}/configs/cloud-init/client.yaml")

  root_block_device {
    volume_size = 100
    volume_type = "gp3"
    encrypted   = true
  }
}

resource "local_file" "cloud_env" {
  filename = "${path.root}/configs/cloud.env"
  content  = "\nVLLM_HOST=${aws_instance.server.private_ip}\nINFPRO_HOST=${aws_instance.server.private_ip}\n"
}

output "az" { value = data.aws_subnet.selected.availability_zone }
output "ssh_key" { value = local_file.ssh_key.filename }
output "server_ip" { value = aws_instance.server.public_ip }
output "client_ip" { value = aws_instance.client.public_ip }
output "all_ips" { value = join(",", [aws_instance.server.public_ip, aws_instance.client.public_ip]) }
