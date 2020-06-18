package v1alpha1

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestData_ParseTile(t *testing.T) {

}

func TestData_ParseDeployment(t *testing.T) {
	x, _ := os.Getwd()
	deploymentSchema = "file://"+x+"/../.././schema/deployment-schema.json"
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
				assert.NoError(t,err)
			}

		})
	}
}
