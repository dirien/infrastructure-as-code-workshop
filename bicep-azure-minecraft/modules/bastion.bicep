@description('The Azure region into which the resources should be deployed.')
param location string = 'westeurope'

@description('The virtual network name to use for the resources.')
param vnetName string

resource vnet 'Microsoft.Network/virtualNetworks@2022-11-01' existing = {
  name: vnetName
}

resource publicIp 'Microsoft.Network/publicIPAddresses@2022-11-01' = {
  name: 'bastion-${uniqueString(resourceGroup().id)}-pip'
  location: location
  tags: {
    app: 'minecraft'
    resources: 'bastion-publicIp'
  }
  sku: {
    name: 'Standard'
  }
  properties: {
    publicIPAllocationMethod: 'Static'
    publicIPAddressVersion: 'IPv4'
  }
}

resource bastionHost 'Microsoft.Network/bastionHosts@2022-11-01' = {
  name: 'bastion-${uniqueString(resourceGroup().id)}-bh'
  location: location
  tags: {
    app: 'minecraft'
    resources: 'bastionHost'
  }
  properties: {
    ipConfigurations: [
      {
        name: 'IpConf'
        properties: {
          subnet: {
            id: vnet.properties.subnets[1].id
          }
          publicIPAddress: {
            id: publicIp.id
          }
        }
      }
    ]
  }
}
