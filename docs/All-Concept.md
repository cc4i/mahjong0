# Motivation

## Bridge gap between builders and customers.
Builders are good at  technology with deep experience, but customers are good at business and seeking best practice to accelerating their business. They were in different focus in terms of business and technology, and need to have a way to glue them together and coordinate smoothly. 
## Share technical & industry best practice across communities.
Builders have deep knowledge of various of solutions, but need a way to scale impact in communities easily and effectively. Customers are constantly looking for best practice across communities to transformer their business, but need a way to allow them experience easily and cost efficiently. 
## Simplify building Cloud Native application on AWS.
Building Cloud Native application is fairly hard, building Cloud Native application with best practice is even hard. Too much to learn and too much to operate, so it became a impossible mission.
## Accelerating adoption to AWS platform.
Customer could quickly spin up solutions on AWS and accelerating their adoption to AWS platform.


# Goal 
- Builders and customers can easily sharing technical & industry best practices through abstract building blocks.
- Builders can easily test those building blocks without real provision.  
- Customers have out-of-box experience without undifferentiated heavy-lifting.
- With built-in abstract layer to accelerate solution building process.
- Builders and customers can collaborate with the most efficient way.  
- Building a community driven project to scale impact.

# Proposal

- Building a platform for both customers and builders to fulfil our goals, which we called Mahjong.
- Mahjong has built-in mechanism to abstract best practice away from traditional solutions , so builders can quickly build new abstract block or full solution based on other building blocks or their own.
- Customer can easily consumer full solution or combine any possible options to build a favorite one.


# High-level Architecture

<img src='./high-level-arch.png'>

Mahjong is a platform to bridge the gap betewwn builders and customers through abstract solution and building blocks. It also could be a tool to simplify the experience and accelerate the cloud adoption on AWS.


# Core concepts

## Tile

A building block, defined by YAML, represents a cloud component or a combination of multiple cloud components or resources. Tile is categorized by Network, ContainerProvider, Storage, Database, Application, ContainerApplication, Analysis, ML. Application and ContainerApplication are represented through commands and files, and rest of categories are represented through Construct::CDK

## Deployment

A unit of deployment,  defined by YAML, and all resources defined within the scope of Tiles.

## Hu

A high level collection of deployment units,  defined by YAML, represents a full solution and includes multiple Tiles with specific definition.

## Dice

A control plane and core orchestration component inside Mahjong, which could be deployed as distributed services when reliability and scalability are highly concerned.
- Work as backend service and responsible for requests from Mahjong CLI/Mahjong Console.
- Dynamically validate and compose Tile, Deployment or Hu into high level provision definition.
- Deploy into AWS Cloud through CDK/CloudFormation and wrapped scripts. 
- Manage the provision process and monitoring relative status and generate final report.  

## Mahjong CLI

A terminal tool to interact with Mahjong, called mctl.

## Mahjong Console

A web UI to interact with Mahjong.

## Solution Marketplace 

A front store, for builders to publish Tile and Hu, for customers to search solutions and kick off provision on AWS straightway.

