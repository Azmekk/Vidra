#!/bin/bash

bunx @openapitools/openapi-generator-cli generate -i ../backend/gen/docs/swagger/swagger.json -g typescript-axios -o ./src/api --global-property models,apis,supportingFiles --additional-properties=withSeparateModelsAndApi=true,apiPackage=api,modelPackage=models