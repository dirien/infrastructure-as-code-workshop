package main

import (
	"fmt"
	"github.com/pulumi/pulumi-openstack/sdk/v3/go/openstack/compute"
	"github.com/pulumi/pulumi-openstack/sdk/v3/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"io/ioutil"
)

const (
	StackName = "minecraft-pulumi"
	Pool      = "floating-net"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		conf := config.New(ctx, "")

		pubKeyFile, err := ioutil.ReadFile("../ssh/workshop.pub")
		if err != nil {
			return err
		}

		keypair, err := compute.NewKeypair(ctx, "keypair", &compute.KeypairArgs{
			Name:      pulumi.String(fmt.Sprintf("%s-kp", StackName)),
			PublicKey: pulumi.String(pubKeyFile),
		})
		if err != nil {
			return err
		}
		network, err := networking.NewNetwork(ctx, "minecraft_net", &networking.NetworkArgs{
			Name:         pulumi.String(fmt.Sprintf("%s-net", StackName)),
			AdminStateUp: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}
		subnet, err := networking.NewSubnet(ctx, "minecraft_snet", &networking.SubnetArgs{
			Name:      pulumi.String(fmt.Sprintf("%s-sub", StackName)),
			NetworkId: network.ID(),
			Cidr:      pulumi.String(conf.Get("subnet-cidr")),
			IpVersion: pulumi.Int(4),
			DnsNameservers: pulumi.StringArray{
				pulumi.String("8.8.8.8"),
				pulumi.String("8.8.4.4"),
			},
		})
		if err != nil {
			return err
		}

		pool := Pool
		floating, err := networking.LookupNetwork(ctx, &networking.LookupNetworkArgs{
			Name: &pool,
		}, nil)

		if err != nil {
			return err
		}
		router, err := networking.NewRouter(ctx, "minecraft_router", &networking.RouterArgs{
			Name:              pulumi.String(fmt.Sprintf("%s-router", StackName)),
			AdminStateUp:      pulumi.Bool(true),
			ExternalNetworkId: pulumi.String(floating.Id),
		})
		if err != nil {
			return err
		}
		_, err = networking.NewRouterInterface(ctx, "minecraft_ri", &networking.RouterInterfaceArgs{
			RouterId: router.ID(),
			SubnetId: subnet.ID(),
		})
		if err != nil {
			return err
		}
		secgroup, err := networking.NewSecGroup(ctx, "minecraft_sg", &networking.SecGroupArgs{
			Name:        pulumi.String(fmt.Sprintf("%s-sg", StackName)),
			Description: pulumi.String("Security group for the Terraform nodes instances"),
		})
		if err != nil {
			return err
		}
		_, err = networking.NewSecGroupRule(ctx, "minecraft_22_sgr", &networking.SecGroupRuleArgs{
			Direction:       pulumi.String("ingress"),
			Ethertype:       pulumi.String("IPv4"),
			Protocol:        pulumi.String("tcp"),
			PortRangeMin:    pulumi.Int(22),
			PortRangeMax:    pulumi.Int(22),
			RemoteIpPrefix:  pulumi.String("0.0.0.0/0"),
			SecurityGroupId: secgroup.ID(),
		})
		if err != nil {
			return err
		}
		_, err = networking.NewSecGroupRule(ctx, "minecraft_25565_sgr", &networking.SecGroupRuleArgs{
			Direction:       pulumi.String("ingress"),
			Ethertype:       pulumi.String("IPv4"),
			Protocol:        pulumi.String("tcp"),
			PortRangeMin:    pulumi.Int(25565),
			PortRangeMax:    pulumi.Int(25565),
			RemoteIpPrefix:  pulumi.String("0.0.0.0/0"),
			SecurityGroupId: secgroup.ID(),
		})
		if err != nil {
			return err
		}

		userData, err := ioutil.ReadFile("../config/cloud-init.yaml")
		if err != nil {
			return err
		}

		vm, err := compute.NewInstance(ctx, "minecraft_vm", &compute.InstanceArgs{
			FlavorName: pulumi.String(conf.Get("flavor")),
			KeyPair:    keypair.Name,
			SecurityGroups: pulumi.StringArray{
				pulumi.String("default"),
				secgroup.Name,
			},
			UserData: pulumi.String(userData),
			Networks: compute.InstanceNetworkArray{
				&compute.InstanceNetworkArgs{
					Name: network.Name,
				},
			},
			BlockDevices: compute.InstanceBlockDeviceArray{
				&compute.InstanceBlockDeviceArgs{
					Uuid:                pulumi.String(conf.Get("image-id")),
					SourceType:          pulumi.String("image"),
					BootIndex:           pulumi.Int(0),
					DestinationType:     pulumi.String("volume"),
					VolumeSize:          pulumi.Int(10),
					DeleteOnTermination: pulumi.Bool(true),
				},
			},
		})
		if err != nil {
			return err
		}
		fip, err := networking.NewFloatingIp(ctx, "minecraft_fip", &networking.FloatingIpArgs{
			Pool: pulumi.String(Pool),
		})
		if err != nil {
			return err
		}
		_, err = compute.NewFloatingIpAssociate(ctx, "minecraft_fipa", &compute.FloatingIpAssociateArgs{
			InstanceId: vm.ID(),
			FloatingIp: fip.Address,
		})
		if err != nil {
			return err
		}

		ctx.Export("minecraft-public", fip.Address)
		return nil
	})
}
