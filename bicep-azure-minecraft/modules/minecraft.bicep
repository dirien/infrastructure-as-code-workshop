@description('The Azure region into which the resources should be deployed.')
param location string = 'westeurope'

@description('The vm size to use for the virtual machine.')
param vmSize string = 'Standard_D2_v2'

@description('The admin username for the virtual machine.')
param adminUsername string

@description('The computer name for the virtual machine.')
param computerName string
@description('The ssh public key to use for the virtual machine.')
param sshPublicKey string
@description('the custom data to use for the virtual machine.')
param customData string

@description('The virtual network name to use for the resources.')
param vnetName string

resource nsg 'Microsoft.Network/networkSecurityGroups@2023-05-01' = {
  name: '${computerName}-${uniqueString(resourceGroup().id)}-nsg'
  location: location
  tags: {
    app: 'minecraft'
    name: computerName
    resources: 'nsg'
  }
  properties: {
    securityRules: [
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
      {
        name: 'minecraft-rcon'
        properties: {
          priority: 1002
          protocol: 'Tcp'
          access: 'Allow'
          direction: 'Inbound'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '25575'
        }
      }
      {
        name: 'minecraft-prom'
        properties: {
          priority: 1003
          protocol: 'Tcp'
          access: 'Allow'
          direction: 'Inbound'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '9090'
        }
      }
    ]
  }
}

resource publicIP 'Microsoft.Network/publicIPAddresses@2023-05-01' = {
  name: '${computerName}-${uniqueString(resourceGroup().id)}-pip'
  location: location
  tags: {
    app: 'minecraft'
    name: computerName
    resources: 'publicIP'
  }
  properties: {
    publicIPAllocationMethod: 'Static'
    publicIPAddressVersion: 'IPv4'
  }
  sku: {
    name: 'Basic'
  }
}

output minecraftPublicIP string = publicIP.properties.ipAddress

resource vnet 'Microsoft.Network/virtualNetworks@2023-05-01' existing = {
  name: vnetName
}

resource nic 'Microsoft.Network/networkInterfaces@2023-05-01' = {
  name: '${computerName}-${uniqueString(resourceGroup().id)}-nic'
  location: location
  tags: {
    app: 'minecraft'
    name: computerName
    resources: 'nic'
  }
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

resource vm 'Microsoft.Compute/virtualMachines@2023-07-01' = {
  name: '${computerName}-${uniqueString(resourceGroup().id)}-vm'
  location: location
  tags: {
    app: 'minecraft'
    name: computerName
    vmSize: vmSize
    resources: 'virtualMachine'
  }
  properties: {
    hardwareProfile: {
      vmSize: vmSize
    }
    storageProfile: {
      osDisk: {
        createOption: 'FromImage'
        name: '${computerName}-${uniqueString(resourceGroup().id)}-disk'
        diskSizeGB: 30
      }
      imageReference: {
        publisher: 'Canonical'
        offer: '0001-com-ubuntu-server-focal'
        sku: '20_04-lts'
        version: 'latest'
      }
    }
    priority: 'Spot'
    evictionPolicy: 'Deallocate'
    billingProfile: {
      maxPrice: -1
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
        patchSettings: {
          patchMode: 'AutomaticByPlatform'
        }
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
