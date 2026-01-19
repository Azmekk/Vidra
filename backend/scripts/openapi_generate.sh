#!/bin/bash

bunx @openapitools/openapi-generator-cli generate \
  -i ./docs/swagger.json \
  -g typescript-axios \
  -o ./src/api \
  --global-property models,apis \
  --additional-properties=withSeparateModelsAndApi=true,apiPackage=api,modelPackage=models