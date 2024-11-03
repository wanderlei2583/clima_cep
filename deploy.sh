#!/bin/bash

source .env

echo "Building and pushing Docker image..."
gcloud builds submit --config cloudbuild.yaml \
  --substitutions=_WEATHER_API_KEY="$WEATHER_API_KEY"

echo "Waiting for deployment to complete..."
gcloud run services describe clima-cep \
  --platform managed \
  --region us-central1 \
  --format='value(status.url)'

echo "Deployment completed!"
