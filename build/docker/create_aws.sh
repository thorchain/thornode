#!/bin/bash

CIDR_BLOCK=192.168.0.0/16
SUBNET=192.168.0.0/24
REGION=us-east-1
export SECURITY_GROUP=docker-machine-thornode

# create vpc 
VPC=$(aws ec2 create-vpc \
    --cidr-block ${CIDR_BLOCK} \
    --region=${REGION})
VPC_ID=$(echo  ${VPC} | jq '.Vpc.VpcId')
VPC_ID=$(sed -e 's/^"//' -e 's/"$//' <<<${VPC_ID})

sleep 5

# create subnet 
aws ec2 create-subnet \
    --vpc-id ${VPC_ID} \
    --cidr-block ${SUBNET} \
    --region=${REGION} \
    --availability-zone ${REGION}a

# make subnet public 
IGW=$(aws ec2 create-internet-gateway \
    --region=${REGION})
IGW_ID=$(echo ${IGW} | jq '.InternetGateway.InternetGatewayId')
IGW_ID=$(sed -e 's/^"//' -e 's/"$//' <<<${IGW_ID})

aws ec2 attach-internet-gateway \
    --vpc-id ${VPC_ID} \
    --internet-gateway-id ${IGW_ID} \
    --region=${REGION}

# create route
ROUTE_TABLE=$(aws ec2 create-route-table \
    --vpc-id ${VPC_ID} \
    --region=${REGION})
ROUTE_TABLE_ID=$(echo ${ROUTE_TABLE} |jq '.RouteTable.RouteTableId')
ROUTE_TABLE_ID=$(sed -e 's/^"//' -e 's/"$//' <<<${ROUTE_TABLE_ID})

aws ec2 create-route \
    --route-table-id ${ROUTE_TABLE_ID} \
    --destination-cidr-block 0.0.0.0/0 \
    --gateway-id ${IGW_ID} \
    --region=${REGION}

# verify route
aws ec2 describe-route-tables \
    --route-table-id ${ROUTE_TABLE_ID} \
    --region=${REGION}

# associate subnet with route table
SUBNET=$(aws ec2 describe-subnets \
    --filters "Name=vpc-id,Values=${VPC_ID}" \
    --query 'Subnets[*].{ID:SubnetId,CIDR:CidrBlock}' \
    --region=${REGION})
SUBNET_ID=$(echo $SUBNET |jq '.[].ID')
SUBNET_ID=$(sed -e 's/^"//' -e 's/"$//' <<<${SUBNET_ID})

aws ec2 associate-route-table  \
    --subnet-id ${SUBNET_ID} \
    --route-table-id ${ROUTE_TABLE_ID} \
    --region=${REGION}

aws ec2 modify-subnet-attribute \
    --subnet-id ${SUBNET_ID} \
    --map-public-ip-on-launch \
    --region=${REGION}

# create security group
SG=$(aws ec2 create-security-group \
    --group-name ${SECURITY_GROUP} \
    --description "Security group for docker-machine" \
    --vpc-id ${VPC_ID} \
    --region=${REGION})
SG_ID=$(echo  ${SG} | jq '.GroupId')
SG_ID=$(sed -e 's/^"//' -e 's/"$//' <<<${SG_ID})

PORTS=(8080 22 2376 1317)

for port in "${PORTS[@]}"
do
    aws ec2 authorize-security-group-ingress \
    --group-id ${SG_ID} \
    --region=${REGION} \
    --protocol tcp \
    --port ${port} \
    --cidr 0.0.0.0/0
done


export AWS_INSTANCE_TYPE=c5.xlarge
export AWS_VPC_ID=${VPC_ID}
export AWS_REGION=${REGION}
export AWS_PROFILE=${AWS_PROFILE} && bash -x ./aws_docker_server.sh
