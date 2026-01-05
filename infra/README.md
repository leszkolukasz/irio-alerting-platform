# Terraform

Terraform is configured to store state on Google Bucket. First create bucket with:

```bash
gcloud storage buckets create gs://${BUCKET_NAME}  --location=europe-west2
```

then initialize each terraform directory with:

```bash
terraform init -backend-config="bucket=${BUCKET_NAME}"
```

### Note:

Service account may need this role to manage Firestore:

```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/datastore.owner"
```
