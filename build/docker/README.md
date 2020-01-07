Thornode Docker
===============

This directory contains helper commands and docker image to run a complete
thornode.

When using "run" commands it runs the service with the data available in
`~/.thornode`. When using the "reset" commands, its the same thing but deletes
all the data before starting the service so we start with a fresh instance.

#### Environments
 * MockNet - is a testnet but using a mock binance server instead of the
   binance testnet
 * TestNet - is a testnet
 * MainNet - is a mainnet

### Standalone Node
To run a single isolated node...
```bash
make run-mocknet-standalone
```

### Genesis Ceremony
To run a 4 node setup conducting a genesis ceremony...

```bash
make run-mocknet-genesis
```

### Run Validator
To run a single node to join an already existing blockchain...

```bash
PEER=<SEED IP ADDRESS> make run-validator
```

Thornode Docker Cloud 
=====================

This directory contains helper commands to run a complete Thornode on any cloud of your choice.

It does this by orchestrating a Linux Server, installs docker and then starts Thornode

Firstly, please use the links below to install docker and docker-compose

https://docs.docker.com/install/

https://docs.docker.com/compose/install/

The supported clouds are
1) AWS
2) Virtualbox 

Virtualbox is the default cloud. And it builds testnet environment on top of it.

To create a server on AWS cloud and start Thornode, you will need to:
1) create and activate an AWS account 
2) create AWS access keys 
3) install AWS CLI
4) configure AWS credentials 
5) create AWS VPC with public subnet and configure the routes 

Please see the useful links below to guide you on how to setup AWS pre-requisites

Run the command below once everything has been setup 

```bash
export THORNODE_ENV=testnet
export AWS_VPC_ID=vpc-***
export AWS_INSTANCE_TYPE=c5.2xlarge
export AWS_REGION=us-east-1
bash docker_server.sh
```

***THORNODE_ENV***: can either be testnet or mocknet but defaults to testnet

***AWS_VPC_ID***:   should be the VPC_ID of the VPC you should just created 

***AWS_REGION***: the region you have created your VPC

***AWS_INSTANCE_TYPE***: See link for more information on instance types 
https://aws.amazon.com/ec2/instance-types/


 

### AWS Useful links

1) Create AWS Account 
https://aws.amazon.com/premiumsupport/knowledge-center/create-and-activate-aws-account/

2) create AWS access keys 
https://aws.amazon.com/premiumsupport/knowledge-center/create-access-key/

3) install AWS CLI
https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html

4) configure AWS credentials
https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/setup-credentials.html

5) create AWS VPC with public subnet 
https://docs.aws.amazon.com/batch/latest/userguide/create-public-private-vpc.html


### VirtualBox

If you do not specify ***AWS_VPC_ID***, ***AWS_REGION*** and ***AWS_INSTANCE_TYPE***, then your server will be provisioned using Virtualbox


```bash
export THORNODE_ENV=testnet
bash docker_server.sh
```


### VirtualBox Useful links 
1) How to install Virtual Box 

https://www.virtualbox.org/wiki/Downloads
