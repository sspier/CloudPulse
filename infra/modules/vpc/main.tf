# fetch available azs so subnets can be spread across them
data "aws_availability_zones" "this" {
  state = "available"
}

# base vpc for cloudpulse
# dns hostnames + support enabled so ecs tasks and load balancers work correctly
resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(
    var.tags,
    { Name = "cloudpulse-vpc" }
  )
}

# internet gateway used by all public subnets
resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = merge(
    var.tags,
    { Name = "cloudpulse-igw" }
  )
}

# public subnets across azs
# used for the alb, nat gateway, and anything that needs direct internet routing
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
  map_public_ip_on_launch = true # give instances public IPs automatically

  tags = merge(
    var.tags,
    {
      Name = "cloudpulse-public-${each.key}"
      Tier = "public"
    }
  )
}

# private subnets across azs
# used for ecs tasks, lambdas with vpc config, and databases
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

# elastic ip used by the nat gateway
# nat gateway provides outbound internet access for private subnets
resource "aws_eip" "nat" {
  tags = merge(
    var.tags,
    { Name = "cloudpulse-nat-eip" }
  )
}

# pick the first public subnet to host the nat gateway
locals {
  first_public_subnet_id = element(values(aws_subnet.public)[*].id, 0)
}

# nat gateway for outbound internet from private subnets
# required for tasks pulling images, lambdas doing outbound calls, etc
resource "aws_nat_gateway" "this" {
  allocation_id = aws_eip.nat.id
  subnet_id     = local.first_public_subnet_id

  tags = merge(
    var.tags,
    { Name = "cloudpulse-nat-gateway" }
  )

  depends_on = [aws_internet_gateway.this]
}

# route table for public subnets
# sends all traffic to the internet gateway
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

  tags = merge(
    var.tags,
    { Name = "cloudpulse-public-rt" }
  )
}

# associate each public subnet with the public route table
resource "aws_route_table_association" "public" {
  for_each       = aws_subnet.public
  subnet_id      = each.value.id
  route_table_id = aws_route_table.public.id
}

# route table for private subnets
# private subnets go out via the nat gateway instead of directly to the internet
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.this.id
  }

  tags = merge(
    var.tags,
    { Name = "cloudpulse-private-rt" }
  )
}

# associate each private subnet with the private route table
resource "aws_route_table_association" "private" {
  for_each       = aws_subnet.private
  subnet_id      = each.value.id
  route_table_id = aws_route_table.private.id
}
