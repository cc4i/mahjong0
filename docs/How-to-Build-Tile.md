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

## Tile with Anything

Building process is going to be almost same and slightly diferent at Step 3. Only need to create folders and add Tile specification without using CDK or mctl, look at Tiles example for further detail.