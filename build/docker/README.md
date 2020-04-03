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

At this point in time only AWS cloud is supported

Virtualbox is the default cloud. And it builds testnet environment on top of it.

To create a server on AWS cloud and start Thornode, you will need to:
1) create and activate an AWS account 
2) create AWS access keys 
3) install AWS CLI
4) configure AWS credentials 

Please see the useful links below to guide you on how to setup AWS pre-requisites

### AWS Useful links

1) Create AWS Account 
https://aws.amazon.com/premiumsupport/knowledge-center/create-and-activate-aws-account/

2) create AWS access keys 
https://aws.amazon.com/premiumsupport/knowledge-center/create-access-key/

3) install AWS CLI
https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html

4) configure AWS credentials
https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/setup-credentials.html


Please don't forget to export AWS_PROFILE on your shell.


Run the command below once your AWS account and profile has been configured

```bash
export THORNODE_ENV=testnet
bash create_aws.sh
```

This will create AWS environment including VPC, subnet and routing.

After the environment is setup, it will build thornode


If you already have your AWS VPC, public subnet and routes setup you can just run the command below 


```bash
export THORNODE_ENV=testnet
export AWS_VPC_ID=vpc***
export AWS_REGION=us-east-1
export AWS_INSTANCE_TYPE=c5.xlarge
bash aws_docker_server.sh
```
 

***THORNODE_ENV***: can either be testnet or mocknet but defaults to testnet

***AWS_VPC_ID***:   should be the VPC_ID of the VPC you should just created 

***AWS_REGION***: the region you have created your VPC

***AWS_INSTANCE_TYPE***: See link for more information on instance types 
https://aws.amazon.com/ec2/instance-types/




