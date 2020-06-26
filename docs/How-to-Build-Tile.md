# How to Build a Tile

There's two way to build Tile as fundemental parts of Mahjong. One way is built on CDK, which is a combination of  [CDK Construct](https://docs.aws.amazon.com/cdk/latest/guide/constructs.html) and Tile specification. Other way is build with various artifacts and Tile specification, let you manipulate vast majority resources.


## Prerequisite
- Install [Docker Desktop](https://docs.docker.com/desktop/#download-and-install)
- Install [AWS CDK](https://docs.aws.amazon.com/cdk/latest/guide/getting_started.html#getting_started_install) 
- Specify [AWS Configuration and Credential](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) setting

## Tile with CDK

1. Create a local folder as a Tile Repo, so that you can develop and test your Tile instantly. Dice will try to load Tile from public Tile repo if can't load Tile from this local repo. 

```bash
# Make foder as your local Tiles repo, see following example folders.
#
# + local-tiles-repo
#    |
#    +- <tile name (lower case)>
#        |
#        +- <version of tile (eg: 0.1.0)>
#            |
#            +- <all content here>

mkdir ~/local-tiles-repo

```



2. Running Dice in DEV mode and then attach local repo and AWS configuration and credentials setting as volume, so Dice could load your Tile and confiuration as needed.

```bash
docker run -it -v ~/local-tiles-repo:/workspace/tiles-repo \
    -v ~/.aws:/root/.aws \
    -e M_MODE=dev \
    -p 9090:9090 \
    herochinese/dice

```

3. Initial sample Tile by mctl and modify futher from there, or using CDK generate a construct and add specification manually.

```bash

# Generate sample Tile by mctl
mctl init tile -n sample-tile
# Or mctl init tile -n <tile name>


# Create Tile manually/
# Step 1: Initial Tile with CDK
export TILE_NAME=<tile name>
export TILE_FOLDER=`echo $TILE_NAME|tr '[:upper:]' '[:lower:]'`
export TILE_VERSION=<tile version, eg: 0.1.0>
mkdir $TILE_FOLDER
cd $TILE_FOLDER
cdk init lib
# Step 2: Adding Tile specification YAML as per schema
# Step 3: Moving files to local repo
mkdir -p ~/local-tiles-repo/$TILE_FOLDER/$TILE_VERSION
cp -R * ~/local-tiles-repo/$TILE_FOLDER/$TILE_VERSION

```
> Here's the [Tile Schema](../templates/tile-schema.json) as mentioned before. 

4. Kick off deployment to try out  very fisrt Tile once it's ready. 

```bash
# Create a deployment specification to refer the Tile.
# Deploy with mctl
mctl deploy -f try-my-tile.yaml

```
> Here's the [Deployment schema](../templates/deployment-schema.json) to fulfil your trial. 


## Useful Tips

0. Key factors for a good Tile

- Think about handle repeated deployment
- Think about multiple Tiles could be deploy at sametime
- Manage all kinds of potential errors

1. How to get rid of back slash
```bash

```
2. Don't use 'sed' to replace any orginal files think about the Tile could be referred multiple time.
```bash
# Don't do 'sed -i -e ...' instead of 'sed -e -e ...'

```
3. Using $cdk() to refer CDK object

4. Using $(*.inputs.*) to refer input values

5. Using $(*.outputs.*) to refer output values

6. Using Global ENV section to simply your Tile specification.

7. Place 'DependentOnVendorService' under 'Metadata' will enforce to inject kube.config, so far only support EKS. You must supply 'clusterName' and 'masterRoleARN' if there's no dependent Tile is EKS.

8. All Tile should be built by CDK or a combination of commands so far, it might change in the future. The ideal language suppose to be Typescript, but you can use other language but you have to convert to TypeScript by [jsii](https://github.com/aws/jsii) before using it. 

## Tile with Anything

Building process is going to be almost same and slightly diferent at Step 3. Only need to create folders and add Tile specification without using CDK or mctl, look at Tiles example for further detail.