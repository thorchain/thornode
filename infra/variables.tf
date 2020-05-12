variable "aws_region" {
  description = "The AWS region to create things in."
  default     = "eu-west-1"
}
variable "ebs_size" {
  description = "Size of ebs volumes"
  default = 50
}
variable "cidr_vpc" {
  description = "CIDR block for the VPC"
  default = "10.1.0.0/16"
}
variable "cidr_subnet" {
  description = "CIDR block for the subnet"
  default = "10.1.0.0/24"
}
variable "availability_zone" {
  description = "availability zone to create subnet"
  default = "eu-west-1a"
}
variable "public_key_path" {
  description = "Public key path"
  default = "~/.ssh/id_rsa.pub"
}
variable "private_key_path" {
  description = "Private key path"
  default = "~/.ssh/id_rsa"
}
variable "instance_type" {
  description = "type for aws EC2 instance"
  default = "t2.micro"
}
variable "ssh_username" {
  description = "SSH username to connect to host"
  default = "ubuntu"
}
