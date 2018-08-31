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
## Issues
Found issue getting dynamic ports to work with AWS ALB Target Group:
```
Good day,

 My name is Junaid and I'll be working with you on this case today.

  I understand that your cloud formation stack with different health and traffic ports fails with the error message as :
"The task definition is configured to use a dynamic host port, but the target group with targetGroupArn arn:aws:elasticloadbalancing:us-west-1:405093580753:targetgroup/ECSTG/da48f96c7786ab3f has a health check port specified.
(Service: AmazonECS; Status Code: 400; Error Code: InvalidParameterException; Request ID: 39537fc7-aa4c-11e8-9700-6b831e64f11c)"

I was able to reproduce the issue from my side when utilizing a CloudFormation stack to create an ECS service with a target group which has a specific health check port and dynamic port mapping is used. I also got the same error message: "The task definition is configured to use a dynamic host port, but the target group with targetGroupArn xxxxxxxxxxx/MyTargetGroup has a health check port specified.

I dig deeper into this behaviour and found that custom health port is not supported by ECS yet, However, Allowing a custom health check options in ECS is an existing feature request which the service team is looking into and I have added your company name to the list of customers requesting this feature. So just to set the expectation here, this is something that may take months to be implemented. I do suggest that you keep an eye on the AWS release page [1].

There are a few workarounds I see which you may consider to fix the issue :

1: Use the same port for Traffic and Health checks, that is whichever port is registered to the TargetGroup will be checked. You can provide a custom path for health checks as well. Ref[2][3]

2: Use "healthCheck" property in the Task definition where you can specify curl commands like this  [ "CMD-SHELL", "curl -f http://localhost:8080/ || exit 1" ] Ref[4][5]

3:  Use a static port for the health check by exposing multiple ports on the container: Traffic port(8080) and health check port(8081).
Have dynamic port mapping for traffic port and static host port mapping for health check port(8081). There is one limitation to this approach: you will have only one task per instance. You can optimize resource utilization by using as small an instance as possible so that CPU and memory resource is not wasted.

--- from your task definition ---
      "portMappings": [
        {
          "containerPort": 8080,
          "hostPort": 0,
          "protocol": "tcp"
        },
        {
          "containerPort": 8081,
          "hostPort": 8081,
          "protocol": "tcp"
        }
      ],
--- end ---


Hope that the above information was helpful. Please let me know in case I have missed anything or should you have any other queries/concerns, or if I did not understand your concern.

Should you require additional assistance, please don't hesitate to contact me. I'm more than happy to assist.

 Have a good day ahead!

References:
[1] https://aws.amazon.com/new/
[2] https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-load-balancing.html
[3] https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html
[4] https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#container_definitions
[5]https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_HealthCheck.html#API_HealthCheck_Contents

Best regards,

Junaid B.
Amazon Web Services

Check out the AWS Support Knowledge Center, a knowledge base of articles and videos that answer customer questions about AWS services: https://aws.amazon.com/premiumsupport/knowledge-center/?icmpid=support_email_category

We value your feedback. Please rate my response using the link below.
===================================================

To contact us again about this case, please return to the AWS Support Center using the following URL:

https://console.aws.amazon.com/support/home#/case/?displayId=5319422871&language=en

(If you are connecting by federation, log in before following the link.)

*Please note: this e-mail was sent from an address that cannot accept incoming e-mail. Please use the link above if you need to contact us again about this same issue.

====================================================================
Learn to work with the AWS Cloud. Get started with free online videos and self-paced labs at
http://aws.amazon.com/training/
====================================================================

Amazon Web Services, Inc. is an affiliate of Amazon.com, Inc. Amazon.com is a registered trademark of Amazon.com, Inc. or its affiliates.

```
Need to specify
[CPU and Memory in a prescribed mix](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-cpu-memory-error.html)
otherwise
you get an error that is obscure!
```
I was able to reproduce this and look at the Service before it terminated.

service soheil-test-2-service-55QHR2FE9FKP was unable to place a task because no container instance met all of its requirements. The closest matching container-instance 40f53686-1332-4b7b-97b1-5b7de7586fd3 has insufficient memory available. For more information, see the Troubleshooting section.

Aug 28, 2018
03:01 PM -0700

Service eventually failed...

Service arn:aws:ecs:us-west-1:405093580753:service/soheil-test-service-YQJ5A9MCG8K1 did not stabilize.

Aug 28, 2018
01:48 PM -0700

The following Cloud Formation Template is hung it has been over 2 hours now, I will leave it here for you to look at, at this point all resources have been created except the Service.

https://us-west-1.console.aws.amazon.com/cloudformation/home?region=us-west-1#/stack/detail?stackId=arn:aws:cloudformation:us-west-1:405093580753:stack%2Fsoheil-test%2F58efe5b0-aaf0-11e8-8f79-50fae8e994c6
```
