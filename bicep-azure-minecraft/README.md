 az group create --name minecraft-rg --location westeurope

 az configure --defaults group=minecraft-rg

 az deployment group create --template-file main.bicep