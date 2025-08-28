# OAC Client CLI

`oac-client` is a simple CLI tool for interacting with **Oracle Analytics Cloud (OAC)** REST APIs. It supports both **Client Credentials** and **Resource Owner Password** OAuth2 flows to obtain access tokens and execute API calls.

---

## Features

- Obtain OAuth2 access tokens from IDCS (Client Credentials or Password grant).  
- Cache tokens on disk (`~/.cache/opsbox/oac_token.json`) to avoid repeated requests.  
- Make REST API calls to OAC with automatic token injection.  
- Retry requests once if a token expires (401 response).  
- Pretty-print JSON responses for readability.  

---

## Installation

Make sure you have Go installed (Go 1.20+ recommended).  

```bash
git clone <repo-url>
cd oac-client
make build
```

## Environment Variables

Set the following environment variables before running the CLI:
```bash
IDCS_TOKEN_URL	        Your IDCS token endpoint URL
IDCS_OAC_CLIENT_ID	    OAuth2 client ID
IDCS_OAC_CLIENT_SECRET	OAuth2 client secret
IDCS_OAC_SCOPE	        OAuth2 scope for the token
IDCS_GRANT_TYPE	        client_credentials/resource_owner
OAC_INSTANCE	        Base URL of your OAC instance

# Resource_owner grant only
OAC_USERNAME	          User login for OAC
OAC_PASSWORD	          User password for OAC 
```

## Make a REST API Call
```bash
./oac-client rest GET /analytics/some-endpoint
./oac-client rest POST /analytics/some-endpoint payload.json


method – HTTP method: GET, POST, PUT, DELETE
path – API path relative to OAC_INSTANCE
payload.json – Optional JSON body file for POST/PUT requests

Responses are automatically pretty-printed.
```