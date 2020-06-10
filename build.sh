#!/usr/bin/env bash

operator-sdk build registry.prod.bbdops.com/common/drift:v0.0.4
docker push registry.prod.bbdops.com/common/drift:v0.0.4