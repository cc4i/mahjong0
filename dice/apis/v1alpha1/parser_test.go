package v1alpha1

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestData_ParseTile(t *testing.T) {
	x, _ := os.Getwd()
	tileSchema = "file://" + x + "/../.././schema/tile-schema.json"
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"Empty content", "", "kind does not match: \"Tile\": kind: kind does not match: \"Tile\""},
		{"Crazy content", "121234~!@#%@#%!$#!@#x", "error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type v1alpha1.Tile"},
		{"Normal content", `# API version
apiVersion: mahjong.io/v1alpha1
# Kind of entity
kind: Tile
# Metadata
metadata:
    # Name of entity
    name: Eks0
    # Category of entity
    category: ContainerProvider
    # Vendor
    vendorService: EKS
    # Version of entity
    version: 0.0.5
# Specification
spec:
  # Dependencies represent dependency with other Tile
  dependencies:
      # As a reference name 
    - name: network
      # Tile name
      tileReference: Network0
      # Tile version
      tileVersion: 0.0.1
  # Inputs are input parameters when lauching 
  inputs:
    # String
    - name: cidr
      inputType: String
      require: true
      override: 
        name: network
        field: cidr
    # CDKObject
    - name: vpc
      inputType: CDKObject
      description: 'Refer to VPC object on Tile - Network0'
      dependencies:
          # Reference name in Dependencies
        - name: network
          # Filed name in refered Tile
          field: baseVpc
      # Input is mandatory or not, true is mandatory and false is optional
      require: false
    # CDKObject[]
    - name: vpcSubnets
      inputType: CDKObject[]
      description: ''
      dependencies:
        - name: network
          field: publicSubnet1
        - name: network
          field: publicSubnet2
        - name: network
          field: privateSubnet1
        - name: network
          field: privateSubnet2
      require: false
    # String
    - name: clusterName
      inputType: String
      description: ''
      defaultValue: default-eks-cluster
      require: true
    # Number/ default: 2
    - name: capacity
      inputType: Number
      description: ''
      defaultValue: 2
      require: false
    # String/ default: 'c5.large'
    - name: capacityInstance
      inputType: String
      description: ''
      defaultValue: 'c5.large'
      require: false
    # String/ default: '1.15'
    - name: clusterVersion
      inputType: String
      description: ''
      defaultValue: '1.16'
      require: false
  # Ouptputs represnt output value after launched, for 'ContainerApplication' might need leverage specific command to retrive output.
  outputs:
    # String
    - name: clusterName
      outputType: String
      description: AWS::EKS::Cluster.Name
    # String
    - name: clusterArn
      outputType: String
      description: AWS::EKS::Cluster.ARN          
    # String
    - name: clusterEndpoint
      outputType: String
      description: AWS::EKS::Cluster.Endpoint
    # String
    - name: masterRoleARN
      outputType: String
      description: AWS::IAM::Role.ARN
    # String
    - name: capacityInstance
      outputType: String
      description: AWS::EKS::Cluster.capacityInstance
    # String/ default: '1.15'
    - name: capacity
      outputType: String
      description: AWS::EKS::Cluster.capacity

  # Notes are description list for addtional information.
  notes:
    - "Tag public subnets with 'kubernetes.io/role/elb=1'"
    - "Tag priavte subnets with 'kubernetes.io/role/internal-elb=1'"

`, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(testing *testing.T) {
			d := Data([]byte(test.input))
			_, err := d.ParseTile(context.TODO())
			if test.output != "" {
				assert.Error(testing, err)
			} else {
				assert.NoError(testing, err)
			}

		})
	}
}

func TestData_ParseDeployment(t *testing.T) {
	x, _ := os.Getwd()
	deploymentSchema = "file://" + x + "/../.././schema/deployment-schema.json"
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"Empty content", "", "deployment specification didn't include tiles"},
		{"Crazy content", "121234~!@#%@#%!$#!@#x", "deployment specification was invalid"},
		{"Normal content", `apiVersion: mahjong.io/v1alpha1
kind: Deployment 
metadata:
  name: eks-simple
spec:
  template:
    tiles:
      tileEks0005:
        tileReference: Eks0
        tileVersion: 0.0.5
        inputs:
          - name: clusterName
            inputValue: mahjong-cluster0
          - name: capacity
            inputValue: 3
          - name: capacityInstance
            inputValue: m5.large
          - name: version
            inputValue: 1.16
  summary:
    description: ""
    outputs: []
    notes: []`, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(testing *testing.T) {
			d := Data([]byte(test.input))
			_, err := d.ParseDeployment(context.TODO())
			if test.output != "" {
				assert.EqualError(testing, err, test.output)
			} else {
				assert.NoError(testing, err)
			}

		})
	}
}

func TestData_ValidateTile(t *testing.T) {
	x, _ := os.Getwd()
	tileSchema = "file://" + x + "/../.././schema/tile-schema.json"
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"Validate normal Tile", `# API version
apiVersion: mahjong.io/v1alpha1
# Kind of entity
kind: Tile
# Metadata
metadata:
    # Name of entity
    name: Eks0
    # Category of entity
    category: ContainerProvider
    # Vendor
    vendorService: EKS
    # Version of entity
    version: 0.0.5
# Specification
spec:
  # Dependencies represent dependency with other Tile
  dependencies:
      # As a reference name 
    - name: network
      # Tile name
      tileReference: Network0
      # Tile version
      tileVersion: 0.0.1
  # Inputs are input parameters when lauching 
  inputs:
    # String
    - name: cidr
      inputType: String
      require: true
      override: 
        name: network
        field: cidr
    # CDKObject
    - name: vpc
      inputType: CDKObject
      description: 'Refer to VPC object on Tile - Network0'
      dependencies:
          # Reference name in Dependencies
        - name: network
          # Filed name in refered Tile
          field: baseVpc
      # Input is mandatory or not, true is mandatory and false is optional
      require: false
    # CDKObject[]
    - name: vpcSubnets
      inputType: CDKObject[]
      description: ''
      dependencies:
        - name: network
          field: publicSubnet1
        - name: network
          field: publicSubnet2
        - name: network
          field: privateSubnet1
        - name: network
          field: privateSubnet2
      require: false
    # String
    - name: clusterName
      inputType: String
      description: ''
      defaultValue: default-eks-cluster
      require: true
    # Number/ default: 2
    - name: capacity
      inputType: Number
      description: ''
      defaultValue: 2
      require: false
    # String/ default: 'c5.large'
    - name: capacityInstance
      inputType: String
      description: ''
      defaultValue: 'c5.large'
      require: false
    # String/ default: '1.15'
    - name: clusterVersion
      inputType: String
      description: ''
      defaultValue: '1.16'
      require: false
  # Ouptputs represnt output value after launched, for 'ContainerApplication' might need leverage specific command to retrive output.
  outputs:
    # String
    - name: clusterName
      outputType: String
      description: AWS::EKS::Cluster.Name
    # String
    - name: clusterArn
      outputType: String
      description: AWS::EKS::Cluster.ARN          
    # String
    - name: clusterEndpoint
      outputType: String
      description: AWS::EKS::Cluster.Endpoint
    # String
    - name: masterRoleARN
      outputType: String
      description: AWS::IAM::Role.ARN
    # String
    - name: capacityInstance
      outputType: String
      description: AWS::EKS::Cluster.capacityInstance
    # String/ default: '1.15'
    - name: capacity
      outputType: String
      description: AWS::EKS::Cluster.capacity

  # Notes are description list for addtional information.
  notes:
    - "Tag public subnets with 'kubernetes.io/role/elb=1'"
    - "Tag priavte subnets with 'kubernetes.io/role/internal-elb=1'"

`, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(testing *testing.T) {
			ctx := context.TODO()
			d := Data([]byte(test.input))
			deploy, _ := d.ParseTile(ctx)
			err := d.ValidateTile(ctx, deploy)
			assert.NoError(testing, err)
		})
	}
}

func TestData_ValidateDeployment(t *testing.T) {
	x, _ := os.Getwd()
	deploymentSchema = "file://" + x + "/../.././schema/deployment-schema.json"
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"Normal deployment", `apiVersion: mahjong.io/v1alpha1
kind: Deployment 
metadata:
  name: eks-simple
spec:
  template:
    tiles:
      tileEks0005:
        tileReference: Eks0
        tileVersion: 0.0.5
        inputs:
          - name: clusterName
            inputValue: mahjong-cluster0
          - name: capacity
            inputValue: 3
          - name: capacityInstance
            inputValue: m5.large
          - name: version
            inputValue: 1.16
  summary:
    description: ""
    outputs: []
    notes: []`, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(testing *testing.T) {
			ctx := context.TODO()
			d := Data([]byte(test.input))
			deploy, _ := d.ParseDeployment(ctx)
			err := d.ValidateDeployment(ctx, deploy)
			assert.NoError(testing, err)
		})
	}
}

func TestLicense_LicenseString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"apache2.0", Apache2.LicenseString(), "Apache2.0"},
		{"mit", MIT.LicenseString(), "MIT"},
	}
	for _, test := range tests {
		t.Run(test.name, func(testing *testing.T) {
			assert.Equal(testing, test.output, test.input)
		})
	}
}

func TestData_CheckParameter(t *testing.T) {
	x, _ := os.Getwd()
	deploymentSchema = "file://" + x + "/../.././schema/deployment-schema.json"
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"Empty content", "", "deployment specification didn't include tiles"},
		{"Crazy content", "121234~!@#%@#%!$#!@#x", "deployment specification was invalid"},
		{"Normal content", `apiVersion: mahjong.io/v1alpha1
kind: Deployment 
metadata:
  name: eks-simple
spec:
  template:
    tiles:
      tileEks0005:
        tileReference: Eks0
        tileVersion: 0.0.5
        inputs:
          - name: clusterName
            inputValue: mahjong-cluster0
          - name: capacity
            inputValue: 3
          - name: capacityInstance
            inputValue: <<parameter>> - m5.large
          - name: version
            inputValue: 1.16
  summary:
    description: ""
    outputs: []
    notes: []`, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(testing *testing.T) {
			d := Data([]byte(test.input))
			deployment, err := d.ParseDeployment(context.TODO())
			if test.output != "" {
				assert.EqualError(testing, err, test.output)
			} else {
				err = d.CheckParameter(context.TODO(), deployment)
				assert.EqualError(testing, err, "input parameters: [capacityInstance] should be replaced by values")
			}

		})
	}
}
