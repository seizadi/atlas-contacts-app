# AWS ECS
There has been a lot of focus on Kubernetes, as seen in this ./deploy
project. I wanted to start examining the current state of AWS ECS.
It is the native AWS container service, although due to the popurarity
of Kubernetes they offer a managed solution for that now.

## Design
We will focus on creating a manifest based on AWS CloudFormation (CF)
template that we will use to launch our instance. The CF template will
have following major sections and you see the major resources that
are managed, the solution will auto-scale as there is demand:
```json
{
   "AWSTemplateFormatVersion":"2010-09-09",
   "Parameters":{  },
   "Mappings":{  },
   "Resources":{
      "ECSCluster":{  },
      "EcsSecurityGroup":{  },
      "EcsSecurityGroupHTTPinbound":{  },
      "EcsSecurityGroupSSHinbound":{  },
      "EcsSecurityGroupALBports":{  },
      "CloudwatchLogsGroup":{  },
      "taskdefinition":{  },
      "ECSALB":{  },
      "ALBListener":{  },
      "ECSALBListenerRule":{  },
      "ECSTG":{  },
      "ECSAutoScalingGroup":{  },
      "ContainerInstances":{  },
      "service":{  },
      "ECSServiceRole":{  },
      "ServiceScalingTarget":{  },
      "ServiceScalingPolicy":{  },
      "ALB500sAlarmScaleUp":{  },
      "EC2Role":{  },
      "AutoscalingRole":{  },
      "EC2InstanceProfile":{  }
   },
   "Outputs":{  }
}
```

The parameters and mappings are basic inputs to drive the template,
you could imagine that these could come from a CMDB and orchestrated
by a template engine.
```json
   "Parameters":{
      "KeyName":{
         "Type":"AWS::EC2::KeyPair::KeyName",
         "Description":"Name of an existing EC2 KeyPair to enable SSH access to the ECS instances."
      },
      "VpcId":{
         "Type":"AWS::EC2::VPC::Id",
         "Description":"Select a VPC that allows instances to access the Internet."
      },
      "SubnetId":{
         "Type":"List<AWS::EC2::Subnet::Id>",
         "Description":"Select at two subnets in your selected VPC."
      },
      "DesiredCapacity":{
         "Type":"Number",
         "Default":"1",
         "Description":"Number of instances to launch in your ECS cluster."
      },
      "MaxSize":{
         "Type":"Number",
         "Default":"1",
         "Description":"Maximum number of instances that can be launched in your ECS cluster."
      },
      "InstanceType":{
         "Description":"EC2 instance type",
         "Type":"String",
         "Default":"t2.micro",
         "AllowedValues":[
            "t2.micro",
            "t2.small",
            "t2.medium",
            "t2.large",
            "m3.medium",
            "m3.large",
            "m3.xlarge",
            "m3.2xlarge",
            "m4.large",
            "m4.xlarge",
            "m4.2xlarge",
            "m4.4xlarge",
            "m4.10xlarge",
            "c4.large",
            "c4.xlarge",
            "c4.2xlarge",
            "c4.4xlarge",
            "c4.8xlarge",
            "c3.large",
            "c3.xlarge",
            "c3.2xlarge",
            "c3.4xlarge",
            "c3.8xlarge",
            "r3.large",
            "r3.xlarge",
            "r3.2xlarge",
            "r3.4xlarge",
            "r3.8xlarge",
            "i2.xlarge",
            "i2.2xlarge",
            "i2.4xlarge",
            "i2.8xlarge"
         ],
         "ConstraintDescription":"Please choose a valid instance type."
      }
   },
   "Mappings":{
      "AWSRegionToAMI":{
         "us-east-1":{
            "AMIID":"ami-eca289fb"
         },
         "us-east-2":{
            "AMIID":"ami-446f3521"
         },
         "us-west-1":{
            "AMIID":"ami-9fadf8ff"
         },
         "us-west-2":{
            "AMIID":"ami-7abc111a"
         },
         "eu-west-1":{
            "AMIID":"ami-a1491ad2"
         },
         "eu-central-1":{
            "AMIID":"ami-54f5303b"
         },
         "ap-northeast-1":{
            "AMIID":"ami-9cd57ffd"
         },
         "ap-southeast-1":{
            "AMIID":"ami-a900a3ca"
         },
         "ap-southeast-2":{
            "AMIID":"ami-5781be34"
         }
      }
   }
```
There are some imperative mixed in with CF immutable specification:
```json
    "ContainerInstances":{
      "Type":"AWS::AutoScaling::LaunchConfiguration",

        "UserData":{
          "Fn::Base64":{
            "Fn::Join":[
              "",
              [
                "#!/bin/bash -xe\n",
                "echo ECS_CLUSTER=",
                {
                  "Ref":"ECSCluster"
                },
                " >> /etc/ecs/ecs.config\n",
                "yum install -y aws-cfn-bootstrap\n",
                "/opt/aws/bin/cfn-signal -e $? ",
                "         --stack ",
                {
                  "Ref":"AWS::StackName"
                },
                "         --resource ECSAutoScalingGroup ",
                "         --region ",
                {
                  "Ref":"AWS::Region"
                },
                "\n"
              ]
            ]
          }
        }
      }
    },
```
The Security Group (SG) is setup with EcsSecurityGroup highlevel
reference attached to the ContainerInstances and ECSALB. We open SSH
port 22 and HTTP port 80.
```json
      "EcsSecurityGroup":{  },
      "EcsSecurityGroupHTTPinbound":{  },
      "EcsSecurityGroupSSHinbound":{  },
      "EcsSecurityGroupALBports":{  },
```
TODO detail about internal ports to be investigated.
```json
```
The definition of the container image is in the
[taskdefinition](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html)
```json
      "taskdefinition":{  },
```
The critical element is the ContainerDefinitions:
```json
        "ContainerDefinitions":[
          {
            "Name":"contacts-server",
            "Cpu":"10",
            "Essential":"true",
            "Image":"https://hub.docker.com/r/soheileizadi/contacts-server:latest",
            "Memory":"300",
            "LogConfiguration":{
              "LogDriver":"awslogs",
              "Options":{
                "awslogs-group":{
                  "Ref":"CloudwatchLogsGroup"
                },
                "awslogs-region":{
                  "Ref":"AWS::Region"
                },
                "awslogs-stream-prefix":"ecs-demo-app"
              }
            },
            "PortMappings":[
              {
                "ContainerPort":8080,
                "ContainerPort":8081,
                "ContainerPort":9090
              }
            ]
          }
```

