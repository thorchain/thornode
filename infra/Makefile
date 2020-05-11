ifndef ENVIRONMENT
    ENVIRONMENT=testnet
endif

create:
	terraform workspace new ${ENVIRONMENT}

apply:
	terraform workspace select ${ENVIRONMENT} && terraform apply -auto-approve

plan:
	terraform workspace select ${ENVIRONMENT} && terraform plan 

destroy:
	terraform workspace select ${ENVIRONMENT} && terraform destroy 
