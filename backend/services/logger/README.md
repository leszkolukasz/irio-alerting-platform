# Logger

This is the Logger microservice that stores logs regarding metrics (ServiceUp/Down) and incidents in Firestrore database.

## Running

API requires a running database and configured topics and subsriptions on GCP. One can use terraform files to set these things up:

- topics and subsriptions are defined in `pubsub.tf`
- firestore is set up in `main.tf`

Note: `terraform destroy` seems to not destroy these resources. Destroy them manually to avoid costs.

To run the Logger service locally, use the following command:

```bash
go run .
```

To test if it works, messages can be created manually in Cloud Console or using gcloud CLI. Example:

```bash
gcloud pubsub topics publish service-up   --message='{"service_id": "payment-gateway-01"}'
```
