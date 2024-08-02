@description('The Azure region into which the resources should be deployed.')
param location string = 'westeurope'
@description('The bastion subnet ip prefix.')
param bastionSubnetIpPrefix string = '10.1.1.0/27'

@description('The address prefixes of the network interfaces.')
var addressPrefix = '10.0.0.0/8'
@description('The address prefixes of the subnets to create.')
var subnetAddressPrefix = '10.0.0.0/16'

resource vnet 'Microsoft.Network/virtualNetworks@2024-01-01' = {
  name: 'vnet-${uniqueString(resourceGroup().id)}'
  location: location
  tags: {
    app: 'minecraft'
    resources: 'vnet'
  }
  properties: {
    addressSpace: {
      addressPrefixes: [
        addressPrefix
      ]
    }
    subnets: [
      {
        name: 'subnet-${uniqueString(resourceGroup().id)}'
        properties: {
          addressPrefix: subnetAddressPrefix
        }
      }
      {
        name: 'AzureBastionSubnet'
        properties: {
          addressPrefix: bastionSubnetIpPrefix
        }
      }
    ]
  }
}

output vnetName string = vnet.name
