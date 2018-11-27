#!/bin/bash

#param 1 image id
#param 2 tag

$(aws ecr get-login --no-include-email --region us-east-1)
docker tag $1 950539697784.dkr.ecr.us-east-1.amazonaws.com/operator:$2
docker push 950539697784.dkr.ecr.us-east-1.amazonaws.com/operator:$2

