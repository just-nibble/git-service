### README: Efficient Actions for Retrieving Commit Data

This README provides instructions on how to efficiently perform the following actions using the API:

1. **Installation**
2. **Add a repo to the DB**
3. **Get the Top N Commit Authors by Commit Counts from the Database**
4. **Retrieve Commits of a Repository by Repository Name from the Database**

Each section includes example `curl` requests for interacting with the API.

###  Warning: Change the values in .env to more production suitable values!
---

### 1. Installation

#### Prerequisite

1. **A linux based operating system**
2. **Docker and docker compose**
3. **Git**

#### Description

1. *Clone the repo*.

```bash
git clone https://github.com/just-nibble/git-service
```

2. cd into directory

```bash
cd git-service
```

3. *Run the code*.

```bash
make
```

The above will create a .env file, run tests and start all containers and seed the database with commits from chromium

### 2. Add a repo to the DB

#### Description

This action gets repo data from github saves to the database and starts indexing the commits.

#### Endpoint
**`POST /repositories`**

- **Parameters**:
  - `limit`: The number of top authors you wish to retrieve (N).

#### Example `curl` Request

```bash
curl --request POST \
  --url http://localhost:8080/repositories \
  --header 'Content-Type: application/json' \
  --data '{
  "owner": "zostera",
  "repo": "django-bootstrap4",
  "since": "2020-01-02"
}'
```

#### Response Example

```json
{
  "id": 2,
  "owner_name": "swaggo",
  "repo_name": "swag",
  "url": "https://github.com/swaggo/swag",
  "forks_count": 1176,
  "stargazers_count": 10339,
  "open_issues_count": 339,
  "watchers_count": 10339,
  "created_at": "2024-08-19T08:59:39.98424535Z",
  "updated_at": "2024-08-19T08:59:39.98424535Z",
  "commits": null,
  "since": "2020-01-02T00:00:00Z"
}
```

---

### 3. Get the Top N Commit Authors by Commit Counts from the Database

#### Description

This action retrieves the top N commit authors, ranked by the number of commits they have made. It can be useful for identifying the most active contributors to a repository.

#### Endpoint

**`GET /authors/top?n=N&repo=R`**

- **Parameters**:
  - `limit`: The number of top authors you wish to retrieve (N).

#### Example `curl` Request

```bash
curl -X GET "http://localhost:8080/authors/top?n=5&repo=chromium" -H "accept: application/json"
```

#### Response Example

```json
[
  {
    "id": 1,
    "name": "Jane Doe",
    "email": "jane@doe.com",
    "commit_count": 120
  },
  {
    "id": 2,
    "name": "John Smith",
    "email": "json@smith.com",
    "commit_count": 95
  }
]
```

---

### 4. Retrieve Commits of a Repository by Repository Name from the Database

#### Description

This action retrieves all the commits for a given repository, identified by its name. It provides an overview of the commit history for the specified repository.

### Endpoint

**`GET /commits/?repo=R`**

- **Path Parameters**:
  - `repo`: The name of the repository whose commits you want to retrieve.

#### Example `curl` Request

```bash
curl -X GET "http://localhost:8080/commits/?repo=chromium" -H "accept: application/json"
```

#### Response Example

```json
[
  {
    "id": 1,
    "hash": "abc123",
    "message": "Initial commit",
    "date": "2024-08-01T12:34:56Z",
    "author": {
      "id": 1,
      "name": "Jane Doe",
      "email": "jane@doe.com",
    },
  },
  {
    "id": 1,
    "commit_hash": "def456",
    "message": "Added new feature",
    "date": "2024-08-02T14:22:11Z",
    "author": {
      "id": 2,
      "name": "John Smith",
      "email": "json@smith.com",
    }
  }
]
```
