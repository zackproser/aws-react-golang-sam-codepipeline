# TODO

- [] Instrument correct region selection in SAM backend (for querying dynamoDB table)
- [] Codify dynamoDB table in SAM template
- [x] Extend UI to read Ripcount on pageload and on response
- [] Extend dynamoDB logic to write all rips to dynamoDB

# Create code pipeline

```
aws cloudformation deploy --template-file code-pipeline.yaml --stack-name pageripper-code-pipeline --profile zack-test-aws --region us-west-2 --parameter-overrides GithubOAuthToken=$GITHUB_OAUTH_TOKEN --capabilities CAPABILITY_NAMED_IAM
```

This app leverages the git push deployment model, by creating a code pipeline leveraging CodeBuild and CodeDeploy for the frontend and backend. The code pipeline itself is codified via a Cloudformation template, which must be deployed prior to development on the frontend or backend.

This app is comprised of a React.js frontend and an AWS lambda leveraging the Golang runtime as a backend.

The code pipeline builds the frontend React.js app and pushes its artifacts to an S3 bucket that serves the Single Page Application (SPA) as a static website.

The backend leverages the AWS Serverless Application Model (SAM). Changes to the backend code are therefore built and delivered via calls to Cloudformation itself.

# Run the UI locally

Export the environment variable that points the UI at the serverless endpoint:

`export REACT_APP_API_URL=<url output from most recent deployment>`

```
cd ui && npm run start
```

If there are issues resolving packages

```
# blow away the local node_modules folder and reinstall
rm -rf node_modules && npm i
```

# Run the lambda + API Gateway backend locally

We pass the `--profile` flag so that the local instance of SAM has sufficient credentials to interact with remote AWS DynamoDB tables, etc.

```
sam local start-api --profile zack-test-aws
```

# Deploying the lambda + API Gateway backend stack to AWS manually

Note that, because the code pipeline currently relies on a cross-stack reference to the lambda backend's URL endpoint variable, you must first deploy this backend stack manually before creating the code pipeline, otherwise the pipeline Cloudformation deployment will fail and rollback automatically when the cross-stack reference cannot be resolved.

1. Build the stack `sam build`
2. Deploy the stack `sam deploy --guided --profile zack-test-aws`
