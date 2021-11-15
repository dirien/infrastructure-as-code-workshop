 az group create --name minecraft-rg --location westeurope

 az configure --defaults group=minecraft-rg

 az deployment group create --template-file main.bicep


az ad sp create-for-rbac \
--name ToyWebsiteWorkflow \
--role Contributor \
--scopes RESOURCE_GROUP_ID \
--sdk-auth