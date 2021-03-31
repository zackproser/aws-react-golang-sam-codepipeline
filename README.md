# Create code pipeline

```
aws cloudformation deploy --template-file code-pipeline.yaml --stack-name pageripper-code-pipeline --profile zack-test-aws --region us-west-2 --parameter-overrides GithubOAuthToken=$GITHUB_OAUTH_TOKEN --capabilities CAPABILITY_NAMED_IAM
```

This app leverages the `git push deployment` model, by creating a code pipeline leveraging CodeBuild and CodeDeploy for the frontend and backend. The code pipeline itself is codified via a Cloudformation template, which must be deployed prior to development on the frontend or backend.

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


# Deploying

1. Build the stack `sam build`
2. Deploy the stack `sam deploy --guided --profile zack-test-aws`
