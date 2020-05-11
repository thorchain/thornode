#providers
provider "aws" {
  shared_credentials_file   = "$HOME/.aws/credentials"
  profile                   = "thornode"
  region                    = var.aws_region
}

# data
# sets the AMI to use
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

#resources
resource "aws_vpc" "vpc" {
  cidr_block = "${var.cidr_vpc}"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = {
    Environment = "${terraform.workspace}"
    ManagedBy   = "Terraform"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = "${aws_vpc.vpc.id}"
  tags =  {
    Environment = "${terraform.workspace}"
    ManagedBy   = "Terraform"
  }
}

resource "aws_subnet" "subnet_public" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "${var.cidr_subnet}"
  map_public_ip_on_launch = "true"
  availability_zone = "${var.availability_zone}"
  tags =  {
    Environment = "${terraform.workspace}"
    ManagedBy   = "Terraform"
  }
}

resource "aws_route_table" "rtb_public" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
      cidr_block = "0.0.0.0/0"
      gateway_id = "${aws_internet_gateway.igw.id}"
  }

  tags =  {
    Environment = "${terraform.workspace}"
    ManagedBy   = "Terraform"
  }
}

resource "aws_route_table_association" "rta_subnet_public" {
  subnet_id      = "${aws_subnet.subnet_public.id}"
  route_table_id = "${aws_route_table.rtb_public.id}"
}

### Security

# Traffic to the ECS Cluster should only come from the ALB
resource "aws_security_group" "sg_thornode" {
  name        = "thornode"
  description = "allow inbound access for thornode"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress { # ssh
    protocol        = "tcp"
    from_port       = 22
    to_port         = 22
    cidr_blocks     = ["0.0.0.0/0"]
  }

  ingress { # midgard
    protocol        = "tcp"
    from_port       = 8080
    to_port         = 8080
    cidr_blocks     = ["0.0.0.0/0"]
  }

  ingress { # tendermint
    protocol        = "tcp"
    from_port       = 26656
    to_port         = 26656
    cidr_blocks     = ["0.0.0.0/0"]
  }

  ingress { # tendermint
    protocol        = "tcp"
    from_port       = 26657
    to_port         = 26657
    cidr_blocks     = ["0.0.0.0/0"]
  }

  ingress { # tss port
    protocol        = "tcp"
    from_port       = 5040
    to_port         = 5040
    cidr_blocks     = ["0.0.0.0/0"]
  }

  ingress { # tss info port
    protocol        = "tcp"
    from_port       = 6060
    to_port         = 6060
    cidr_blocks     = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "thornode-${terraform.workspace}"
    Environment = "${terraform.workspace}"
    ManagedBy   = "Terraform"
  }
}

resource "aws_security_group" "sg_binance" {
  name        = "binance-node"
  description = "allow inbound access for binance node"
  vpc_id      = "${aws_vpc.vpc.id}"

  ingress { # ssh
    protocol        = "tcp"
    from_port       = 22
    to_port         = 22
    cidr_blocks     = ["0.0.0.0/0"]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 27146
    to_port         = 27146
    security_groups = ["${aws_security_group.sg_thornode.id}"]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 27147
    to_port         = 27147
    security_groups = ["${aws_security_group.sg_thornode.id}"]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 26660
    to_port         = 26660
    security_groups = ["${aws_security_group.sg_thornode.id}"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "binance-${terraform.workspace}"
    Environment = "${terraform.workspace}"
    ManagedBy   = "Terraform"
  }
}

resource "aws_key_pair" "ec2key" {
  key_name = "publicKey"
  public_key = "${file(var.public_key_path)}"
}

resource "aws_instance" "thornode" {
  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "${var.instance_type}"
  subnet_id = "${aws_subnet.subnet_public.id}"
  vpc_security_group_ids = ["${aws_security_group.sg_thornode.id}"]
  key_name = "${aws_key_pair.ec2key.key_name}"

  tags = {
    Name        = "thornode-${terraform.workspace}"
    Environment = "${terraform.workspace}"
    ManagedBy   = "Terraform"
  }
}

resource "aws_eip" "thornode" {
  vpc = true

  instance                  = "${aws_instance.thornode.id}"
  associate_with_private_ip = "${aws_instance.thornode.private_ip}"
  depends_on                = [aws_internet_gateway.igw]
}

resource "null_resource" "thornode" {

  connection {
    user        = "ubuntu"
    private_key = "${file("${var.private_key_path}")}"
    agent       = true
    timeout     = "3m"
    host        = "${aws_eip.thornode.public_ip}"
  }
 
  provisioner "file" {
    source      = "./scripts/ec2-userdata.bash"
    destination = "/tmp/ec2-userdata.bash"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo bash /tmp/ec2-userdata.bash",
    ]
  }
}
