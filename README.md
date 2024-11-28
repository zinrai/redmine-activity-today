# redmine-activity-today

A lightweight web application that displays today's Redmine activities in a consolidated view. This tool helps teams track their progress by aggregating issues from one or multiple Redmine instances.

## Features

- Supports multiple Redmine instances
- Groups issues by assignee
- Configurable through YAML

## Installation

```bash
$ go build
```

## Configuration

Create a `config.yaml` file in the project root:

```yaml
redmine_urls:
  - url: "http://your-redmine-instance/"
    api_key: "your-api-key"
    query_id: 12
    limit: 100
# Add more Redmine instances as needed
# - url: "http://another-redmine-instance/"
#   api_key: "another-api-key"
#   query_id: 7
#   limit: 100
```

Configuration parameters:

- `url`: Your Redmine instance URL
- `api_key`: Your Redmine API key
- `query_id`: The ID of your saved query in Redmine
- `limit`: Maximum number of issues to fetch

## Usage

1. Start the server:

```bash
go run main.go
```

2. Open your web browser and navigate to:

```
http://localhost:8000
```

The application will display a table of today's Redmine activities, grouped by assignee.

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
