# Example local usage

After creating topic and subscription you can publish message using this command:

```bash
gcloud pubsub topics publish notify-oncaller --message='{"incident_id": "INC-555", "service_id": 101, "oncaller": "dwieexgfrcrjdrbftf@xfavaj.com", "timestamp": "2024-01-03T16:00:00Z"}' --project=$PROJECT_ID
```

Locally it will send email catched by Mailtrap. On deployment it will send real mail message.
