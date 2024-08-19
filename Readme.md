### README: Efficient Actions for Retrieving Commit Data

This README provides instructions on how to efficiently perform the following actions using the API:

1. **Get the Top N Commit Authors by Commit Counts from the Database**
2. **Retrieve Commits of a Repository by Repository Name from the Database**

Each section includes example `curl` requests for interacting with the API.

---

### 1. Get the Top N Commit Authors by Commit Counts from the Database

#### Description
This action retrieves the top N commit authors, ranked by the number of commits they have made. It can be useful for identifying the most active contributors to a repository.

#### Endpoint
**`GET /authors/top?limit=N`**

- **Parameters**:
  - `limit`: The number of top authors you wish to retrieve (N).

#### Example `curl` Request
```bash
curl -X GET "http://localhost:8080/authors/top?limit=5" -H "accept: application/json"
```

#### Response Example
```json
[
  {
    "author_name": "Jane Doe",
    "commit_count": 120
  },
  {
    "author_name": "John Smith",
    "commit_count": 95
  }
]
```

---

### 2. Retrieve Commits of a Repository by Repository Name from the Database

#### Description
This action retrieves all the commits for a given repository, identified by its name. It provides an overview of the commit history for the specified repository.

### Endpoint

**`GET /repositories/{name}/commits`**

- **Path Parameters**:
  - `name`: The name of the repository whose commits you want to retrieve.

#### Example `curl` Request
```bash
curl -X GET "http://localhost:8080/repositories/example-repo/commits" -H "accept: application/json"
```

#### Response Example
```json
[
  {
    "commit_hash": "abc123",
    "message": "Initial commit",
    "date": "2024-08-01T12:34:56Z",
    "author_name": "Jane Doe"
  },
  {
    "commit_hash": "def456",
    "message": "Added new feature",
    "date": "2024-08-02T14:22:11Z",
    "author_name": "John Smith"
  }
]
```