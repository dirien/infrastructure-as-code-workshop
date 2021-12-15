 az group create --name minecraft-rg --location westeurope

 az configure --defaults group=minecraft-rg

 az deployment group create --template-file main.bicep


az group show --name minecraft-rg

az ad sp create-for-rbac \
--name MinecraftGithubDeployer \
--role Contributor \
--scopes /subscriptions/d62eb9ef-fe7c-45f5-8bd6-7b2b727869c4/resourceGroups/minecraft-rg \
--sdk-auth

save output in GH secret

az deployment group what-if --template-file main.bicep --mode Complete --parameters objectId=1cbd6d17-3682-4175-9ba2-adb7582c3507 --parameters userId=9ddd81b2-ebc6-498b-95bb-8ec7f57c7558az


az deployment group show  -n main --query properties.outputs.minecraftPublicIP