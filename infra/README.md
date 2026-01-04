# Terraform

Terraform is configured to store state on Google Bucket. First create bucket with:

```bash
gcloud storage buckets create gs://${BUCKET_NAME}  --location=europe-west2
```

then initialize each terraform directory with:

```bash
terraform init -backend-config="bucket=${BUCKET_NAME}"
```
