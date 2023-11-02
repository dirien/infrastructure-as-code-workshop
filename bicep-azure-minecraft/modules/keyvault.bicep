@description('The Azure region into which the resources should be deployed.')
param location string = 'westeurope'

@description('The sku of the Azure resources to create.')
param sku string = 'Standard'

@description('The Object Id of the service principal to use the access policy.')
param serviceAccountObjectID string

@secure()
@description('The private ssh key content')
param sshKey string

@description('Object Id of the real user, could be done via the UI too')
param currentUserObjectId string

@description('The access policies for the Key Vault.')
param accessPolicies array = [
  {
    tenantId: subscription().tenantId
    objectId: serviceAccountObjectID
    permissions: {
      keys: [
        'Get'
        'List'
        'Update'
        'Create'
        'Import'
        'Delete'
        'Recover'
        'Backup'
        'Restore'
      ]
      secrets: [
        'Get'
        'List'
        'Set'
        'Delete'
        'Recover'
        'Backup'
        'Restore'
      ]
      certificates: [
        'Get'
        'List'
        'Update'
        'Create'
        'Import'
        'Delete'
        'Recover'
        'Backup'
        'Restore'
        'ManageContacts'
        'ManageIssuers'
        'GetIssuers'
        'ListIssuers'
        'SetIssuers'
        'DeleteIssuers'
      ]
    }
  }
  {
    tenantId: subscription().tenantId
    objectId: currentUserObjectId
    permissions: {
      keys: [
        'Get'
        'List'
        'Update'
        'Create'
        'Import'
        'Delete'
        'Recover'
        'Backup'
        'Restore'
      ]
      secrets: [
        'Get'
        'List'
        'Set'
        'Delete'
        'Recover'
        'Backup'
        'Restore'
      ]
      certificates: [
        'Get'
        'List'
        'Update'
        'Create'
        'Import'
        'Delete'
        'Recover'
        'Backup'
        'Restore'
        'ManageContacts'
        'ManageIssuers'
        'GetIssuers'
        'ListIssuers'
        'SetIssuers'
        'DeleteIssuers'
      ]
    }
  }
]

resource keyvault 'Microsoft.KeyVault/vaults@2023-07-01' = {
  name: 'keyvaul-${uniqueString(resourceGroup().id)}'
  location: location
  tags: {
    app: 'minecraft'
    resources: 'keyvault'
  }
  properties: {
    enabledForDeployment: true
    enabledForTemplateDeployment: true
    enabledForDiskEncryption: false
    tenantId: subscription().tenantId
    accessPolicies: accessPolicies
    sku: {
      name: sku
      family: 'A'
    }
    networkAcls: {
      defaultAction: 'Allow'
      bypass: 'AzureServices'
    }
  }
}

resource secret 'Microsoft.KeyVault/vaults/secrets@2023-07-01' = {
  name: '${keyvault.name}/ssh'
  tags: {
    app: 'minecraft'
    resources: 'secret'
  }
  properties: {
    value: sshKey
  }
}
