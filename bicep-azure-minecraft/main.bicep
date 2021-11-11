param location string = resourceGroup().location

module minecraftServer 'modules/minecraft.bicep' = {
  name: 'minecraft'
  params: {
    location: location
    adminUsername: 'minecraft'
    computerName: 'minecraft'
    sshPublicKey: loadTextContent('../ssh/workshop.pub')
    customData: loadFileAsBase64('../config/cloud-init.yaml')
  }
}

output minecraftPublicIP string = minecraftServer.outputs.minecraftPublicIP
output sshCommand string = minecraftServer.outputs.sshCommand
