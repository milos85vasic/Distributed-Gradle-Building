data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  azs = slice(data.aws_availability_zones.available.names, 0, 3)
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = var.name
  cidr = var.cidr

  azs             = local.azs
  private_subnets = var.private_subnets
  public_subnets  = var.public_subnets

  enable_nat_gateway   = var.enable_nat_gateway
  single_nat_gateway   = var.single_nat_gateway
  enable_dns_hostnames = var.enable_dns_hostnames
  enable_dns_support   = var.enable_dns_support

  # VPC Flow Logs
  enable_flow_log                      = true
  create_flow_log_cloudwatch_log_group = true
  create_flow_log_cloudwatch_iam_role  = true
  flow_log_max_aggregation_interval    = 60

  # Default security group - restrict access
  manage_default_security_group = true
  default_security_group_ingress = []
  default_security_group_egress  = []

  # Additional tags
  tags = merge(var.tags, {
    "kubernetes.io/cluster/${var.name}" = "shared"
  })

  public_subnet_tags = merge(var.public_subnet_tags, {
    "kubernetes.io/cluster/${var.name}" = "shared"
    "kubernetes.io/role/elb"            = "1"
  })

  private_subnet_tags = merge(var.private_subnet_tags, {
    "kubernetes.io/cluster/${var.name}" = "shared"
    "kubernetes.io/role/internal-elb"   = "1"
  })
}

# VPC Endpoints for AWS services (cost optimization and security)
resource "aws_vpc_endpoint" "s3" {
  vpc_id       = module.vpc.vpc_id
  service_name = "com.amazonaws.${var.region}.s3"

  tags = {
    Name = "${var.name}-s3-endpoint"
  }
}

resource "aws_vpc_endpoint" "dynamodb" {
  vpc_id       = module.vpc.vpc_id
  service_name = "com.amazonaws.${var.region}.dynamodb"
  vpc_endpoint_type = "Gateway"

  tags = {
    Name = "${var.name}-dynamodb-endpoint"
  }
}

resource "aws_vpc_endpoint" "ecr_dkr" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${var.region}.ecr.dkr"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  security_group_ids  = [aws_security_group.vpc_endpoint.id]
  private_dns_enabled = true

  tags = {
    Name = "${var.name}-ecr-dkr-endpoint"
  }
}

resource "aws_vpc_endpoint" "ecr_api" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${var.region}.ecr.api"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  security_group_ids  = [aws_security_group.vpc_endpoint.id]
  private_dns_enabled = true

  tags = {
    Name = "${var.name}-ecr-api-endpoint"
  }
}

resource "aws_vpc_endpoint" "logs" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${var.region}.logs"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  security_group_ids  = [aws_security_group.vpc_endpoint.id]
  private_dns_enabled = true

  tags = {
    Name = "${var.name}-logs-endpoint"
  }
}

# Security group for VPC endpoints
resource "aws_security_group" "vpc_endpoint" {
  name_prefix = "${var.name}-vpc-endpoint-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [var.cidr]
  }

  tags = {
    Name = "${var.name}-vpc-endpoint"
  }
}

# Route table associations for VPC endpoints
resource "aws_vpc_endpoint_route_table_association" "s3" {
  route_table_id  = module.vpc.private_route_table_ids[0]
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
}

resource "aws_vpc_endpoint_route_table_association" "dynamodb" {
  route_table_id  = module.vpc.private_route_table_ids[0]
  vpc_endpoint_id = aws_vpc_endpoint.dynamodb.id
}