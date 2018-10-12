# Text-to-speech Lambda

A simple program which takes textual input and converts it to an audio file.
The audio is then uploaded to AWS S3. Along with the upload, a message containing the
audio file name is pushed to AWS SQS.

## Request
```json
{
    "text": "This is the text you want to convert to speech.",
    "voice": "Emma" // Must be a valid VoiceId for AWS Polly
}
```

## Response
```json
{
    "success": "true", // or false
    "file_name": "de14c41a-ce5f-11e8-8341-ee4ec834ee6a", // empty if `success` is false
    "file_type": "mp3" // empty if `success` is false
}
```