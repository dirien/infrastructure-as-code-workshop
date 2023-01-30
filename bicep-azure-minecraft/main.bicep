param location string = resourceGroup().location

param objectId string
param userId string

module minecraftVault 'modules/keyvault.bicep' = {
  name: 'minecraft-vault'
  params: {
    location: location
    serviceAccountObjectID: objectId
    currentUserObjectId: userId
    sshKey: loadTextContent('../ssh/workshop')
  }
}

module minecraftVnet 'modules/network.bicep' = {
  name: 'minecraft-vnet'
  params: {
    location: location
  }
}

module minecraftBastion 'modules/bastion.bicep' = {
  name: 'minecraft-bastion'
  params: {
    location: location
    vnetName: minecraftVnet.outputs.vnetName
  }
}

var minecraftServeNames = [
  'minecraft-server-1'
]

module minecraftServer 'modules/minecraft.bicep' = [for name in minecraftServeNames: {
  name: name
  params: {
    location: location
    vnetName: minecraftVnet.outputs.vnetName
    adminUsername: 'minecraft'
    computerName: name
    sshPublicKey: loadTextContent('../ssh/workshop.pub')
    customData: loadFileAsBase64('config/papermc-cloud-init.yaml')
  }
}]

output minecraftPublicIP array = [for (item, index) in minecraftServeNames: {
  name: item
  value: minecraftServer[index].outputs.minecraftPublicIP
}]
