// VPC module: creates a VPC with public + private subnets, IGW, NAT gateway, and route tables.

data "aws_availability_zones" "this" {
  state = "available"
}

resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(
    var.tags,
    {
      Name = "cloudpulse-vpc"
    }
  )
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = merge(
    var.tags,
    {
      Name = "cloudpulse-igw"
    }
  )
}

// public subnets (for ALB, NAT gateway, etc.)
resource "aws_subnet" "public" {
  for_each = {
    for idx, cidr in var.public_subnet_cidrs :
    idx => {
      cidr = cidr
      az   = data.aws_availability_zones.this.names[idx]
    }
  }

  vpc_id                  = aws_vpc.this.id
  cidr_block              = each.value.cidr
  availability_zone       = each.value.az
  map_public_ip_on_launch = true

  tags = merge(
    var.tags,
    {
      Name = "cloudpulse-public-${each.key}"
      Tier = "public"
    }
  )
}

// private subnets (for ECS tasks, DBs, etc.)
resource "aws_subnet" "private" {
  for_each = {
    for idx, cidr in var.private_subnet_cidrs :
    idx => {
      cidr = cidr
      az   = data.aws_availability_zones.this.names[idx]
    }
  }

  vpc_id            = aws_vpc.this.id
  cidr_block        = each.value.cidr
  availability_zone = each.value.az

  tags = merge(
    var.tags,
    {
      Name = "cloudpulse-private-${each.key}"
      Tier = "private"
    }
  )
}

// allocate EIP for the NAT gateway in the first public subnet
resource "aws_eip" "nat" {
  tags = merge(
    var.tags,
    {
      Name = "cloudpulse-nat-eip"
    }
  )
}

// NAT gateway lives in the first public subnet
locals {
  first_public_subnet_id = element(values(aws_subnet.public)[*].id, 0)
}

resource "aws_nat_gateway" "this" {
  allocation_id = aws_eip.nat.id
  subnet_id     = local.first_public_subnet_id

  tags = merge(
    var.tags,
    {
      Name = "cloudpulse-nat-gateway"
    }
  )

  depends_on = [aws_internet_gateway.this]
}

// public route table (0.0.0.0/0 → IGW)
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

  tags = merge(
    var.tags,
    {
      Name = "cloudpulse-public-rt"
    }
  )
}

// associate all public subnets with the public route table
resource "aws_route_table_association" "public" {
  for_each       = aws_subnet.public
  subnet_id      = each.value.id
  route_table_id = aws_route_table.public.id
}

// private route table (0.0.0.0/0 → NAT)
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.this.id
  }

  tags = merge(
    var.tags,
    {
      Name = "cloudpulse-private-rt"
    }
  )
}

// associate all private subnets with the private route table
resource "aws_route_table_association" "private" {
  for_each       = aws_subnet.private
  subnet_id      = each.value.id
  route_table_id = aws_route_table.private.id
}
