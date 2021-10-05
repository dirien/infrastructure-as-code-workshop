package main

import (
	"encoding/base64"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/compute"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/network"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"io/ioutil"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		c := config.New(ctx, "")
		name := c.Require("name")
		// Create an Azure Resource Group
		resourceGroup, err := resources.NewResourceGroup(ctx, "minecraft", &resources.ResourceGroupArgs{})
		if err != nil {
			return err
		}

		virtualNetwork, err := network.NewVirtualNetwork(ctx, "minecraft-vnic", &network.VirtualNetworkArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			AddressSpace: &network.AddressSpaceArgs{
				AddressPrefixes: pulumi.StringArray{
					pulumi.String("10.0.0.0/8"),
				},
			},
		})
		if err != nil {
			return err
		}
		subnet, err := network.NewSubnet(ctx, "minecraft-subnet", &network.SubnetArgs{
			Name:               pulumi.String("minecraft-subnet"),
			ResourceGroupName:  resourceGroup.Name,
			VirtualNetworkName: virtualNetwork.Name,
			AddressPrefix:      pulumi.String("10.0.0.0/16"),
		})
		if err != nil {
			return err
		}

		ipAddress, err := network.NewPublicIPAddress(ctx, "minecraft-pubip", &network.PublicIPAddressArgs{
			Location:                 resourceGroup.Location,
			ResourceGroupName:        resourceGroup.Name,
			PublicIPAddressVersion:   network.IPVersionIPv4,
			PublicIPAllocationMethod: network.IPAllocationMethodStatic,
		})
		if err != nil {
			return err
		}

		networkInterface, err := network.NewNetworkInterface(ctx, "minecraft-nic", &network.NetworkInterfaceArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			IpConfigurations: network.NetworkInterfaceIPConfigurationArray{
				&network.NetworkInterfaceIPConfigurationArgs{
					Name:                      pulumi.String("internal"),
					PrivateIPAllocationMethod: network.IPAllocationMethodDynamic,
					Subnet: network.SubnetTypeArgs{
						Id: subnet.ID(),
					},
					PublicIPAddress: network.PublicIPAddressTypeArgs{
						Id: ipAddress.ID(),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		userData, err := ioutil.ReadFile("../config/cloud-init.yaml")
		if err != nil {
			return err
		}
		pubKeyFile, err := ioutil.ReadFile("../ssh/workshop.pub")
		if err != nil {
			return err
		}

		_, err = compute.NewVirtualMachine(ctx, "minecraft-server", &compute.VirtualMachineArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			HardwareProfile: compute.HardwareProfileArgs{
				VmSize: compute.VirtualMachineSizeTypes("Standard_D2_v2"),
			},
			StorageProfile: compute.StorageProfileArgs{
				ImageReference: compute.ImageReferenceArgs{
					Publisher: pulumi.String("Canonical"),
					Offer:     pulumi.String("0001-com-ubuntu-server-focal"),
					Sku:       pulumi.String("20_04-lts"),
					Version:   pulumi.String("latest"),
				},
			},
			OsProfile: compute.OSProfileArgs{
				ComputerName:  pulumi.String(name),
				AdminUsername: pulumi.String("minecraft"),
				CustomData:    pulumi.String(base64.StdEncoding.EncodeToString([]byte(userData))),
				LinuxConfiguration: compute.LinuxConfigurationArgs{
					Ssh: compute.SshConfigurationArgs{
						PublicKeys: compute.SshPublicKeyTypeArray{
							compute.SshPublicKeyTypeArgs{
								Path:    pulumi.String("/home/minecraft/.ssh/authorized_keys"),
								KeyData: pulumi.String(pubKeyFile),
							},
						},
					},
				},
			},
			NetworkProfile: compute.NetworkProfileArgs{
				NetworkInterfaces: compute.NetworkInterfaceReferenceArray{
					compute.NetworkInterfaceReferenceArgs{
						Id:      networkInterface.ID(),
						Primary: pulumi.BoolPtr(true),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("ip", ipAddress.IpAddress)

		return nil
	})

}
