### README: Efficient Actions for Retrieving Commit Data

This README provides instructions on how to efficiently perform the following actions using the API:

1. **Installation**
2. **Add a repo to the DB**
3. **Usage**


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
Add a GITHUB_TOKEN variable to the env if you possess a github token

### 2. Add a repo to the DB

#### Description

This action gets repo data from github saves to the database and starts indexing the commits.

#### Endpoint
**`POST /repositories`**


#### Example `curl` Request

```bash
curl --request POST \
  --url http://localhost:8080/repositories \
  --header 'Content-Type: application/json' \
  --header 'User-Agent: insomnia/9.3.3' \
  --data '{"name": "swaggo/swag"}'
```

#### Response Example
1. "Repository successfully indexed, its commits are being fetched..." (successfull)
2. {"error":"unexpected response status: 403"} (rate limited)
---

### 3. Usage

#### Get the Top N Commit Authors by Commit Counts from the Database ####

#### Description

This action retrieves the top N commit authors, ranked by the number of commits they have made. It can be useful for identifying the most active contributors to a repository.

#### Endpoint

**`GET /authors/top?n=N&repo=R`**

- **Parameters**:
  - `n`: The number of top authors you wish to retrieve.
  - `repo`: The repository name

#### Example `curl` Request

```bash
curl -X GET "http://localhost:8080/authors/chromium/chromium/top?n=5" -H "accept: application/json"
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
curl -X GET "http://localhost:8080/commits/chromium/chromium" -H "accept: application/json"
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
