var addressPrefix = '10.0.0.0/8'
var subnetAddressPrefix = '10.0.0.0/16'

param location string = 'westeurope'

param vmSize string = 'Standard_D2_v2'
param adminUsername string
param computerName string
param sshPublicKey string
param customData string

resource nsg 'Microsoft.Network/networkSecurityGroups@2020-06-01' = {
  name: 'nsg-${uniqueString(resourceGroup().id)}'
  location: location
  properties: {
    securityRules: [
      {
        name: 'SSH'
        properties: {
          priority: 1000
          protocol: 'Tcp'
          access: 'Allow'
          direction: 'Inbound'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '22'
        }
      }
      {
        name: 'minecraft'
        properties: {
          priority: 1001
          protocol: 'Tcp'
          access: 'Allow'
          direction: 'Inbound'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '25565'
        }
      }
    ]
  }
}

resource publicIP 'Microsoft.Network/publicIPAddresses@2021-03-01' = {
  name: 'pup-${uniqueString(resourceGroup().id)}'
  location: location
  properties: {
    publicIPAllocationMethod: 'Static'
    publicIPAddressVersion: 'IPv4'
  }
  sku: {
    name: 'Basic'
  }
}

output minecraftPublicIP string = publicIP.properties.ipAddress

resource vnet 'Microsoft.Network/virtualNetworks@2021-03-01' = {
  name: 'vnet-${uniqueString(resourceGroup().id)}'
  location: location
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
    ]
  }
}

resource nic 'Microsoft.Network/networkInterfaces@2021-03-01' = {
  name: 'nic-${uniqueString(resourceGroup().id)}'
  location: location
  properties: {
    ipConfigurations: [
      {
        name: 'ipconfig1'
        properties: {
          subnet: {
            id: vnet.properties.subnets[0].id
          }
          privateIPAllocationMethod: 'Dynamic'
          publicIPAddress: {
            id: publicIP.id
          }
        }
      }
    ]
    networkSecurityGroup: {
      id: nsg.id
    }
  }
}

resource vm 'Microsoft.Compute/virtualMachines@2021-07-01' = {
  name: 'minecraft-${uniqueString(resourceGroup().id)}'
  location: location
  properties: {
    hardwareProfile: {
      vmSize: vmSize
    }
    storageProfile: {
      osDisk: {
        createOption: 'FromImage'
        name: 'osdisk-${uniqueString(resourceGroup().id)}'
        diskSizeGB: 30
      }
      imageReference: {
        publisher: 'Canonical'
        offer: '0001-com-ubuntu-server-focal'
        sku: '20_04-lts'
        version: 'latest'
      }
    }
    networkProfile: {
      networkInterfaces: [
        {
          id: nic.id
          properties: {
            primary: true
          }
        }
      ]
    }
    osProfile: {
      computerName: computerName
      adminUsername: adminUsername
      customData: customData
      linuxConfiguration: {
        ssh: {
          publicKeys: [
            {
              path: '/home/${adminUsername}/.ssh/authorized_keys'
              keyData: sshPublicKey
            }
          ]
        }
      }
    }
  }
}

output sshCommand string = 'ssh -i <sshkey> ${adminUsername}@${publicIP.properties.ipAddress}'
