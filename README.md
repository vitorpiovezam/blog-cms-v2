# blog-cms-v2
## Serverless v3 - AWS Lambda (Go)

Blog API deployed with the [Serverless framework](https://www.serverless.com/). Handlers are written in Go and run on Lambda with the `provided.al2023` runtime.

For detailed instructions, please refer to the [documentation](https://www.serverless.com/framework/docs/providers/aws/).

## Installation/deployment instructions

Follow the steps below to run locally or deploy to AWS.

> **Requirements**: Go `(1.23+)`, Python 3 (for packaging), and the [Serverless CLI](https://www.serverless.com/framework/docs/getting-started) installed globally. AWS CLI credentials configured for deploy.

### Setup

- Run `make setup` to fetch Go dependencies

### Local development

- Run `make serve` to start the API on `http://localhost:3000`
- `make offline` is an alias for `make serve`

Endpoints:

- `GET http://localhost:3000/dev/posts`
- `GET http://localhost:3000/dev/post/{slug}`

### Deploy

- Run `make deploy` to build, package, and deploy to AWS

### Other commands

- `make build` — compile the Linux binary
- `make package` — build + create the deployment zip
- `make clean` — remove build artifacts
