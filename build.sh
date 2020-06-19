#!/usr/bin/env bash

operator-sdk build registry.prod.bbdops.com/common/drift:v0.0.7
docker push registry.prod.bbdops.com/common/drift:v0.0.7