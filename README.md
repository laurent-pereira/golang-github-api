# Canvas for Backend Technical Test at Scalingo

## Instructions

* From this canvas, respond to the project which has been communicated to you by our team
* Feel free to change everything

## Setup and Execution

Create a personal Github Token
[https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
````
git clone git@github.com:laurent-pereira/golang-github-api.git
cd golang-github-api
cp sample.env .env
# Add your GitHub token to .env
docker-compose up
````

Application will be then running on port `5000`

## Test

````
$ curl localhost:5000/ping
{ "status": "pong" }
````

## Basic Usage

### /repos Endpoint
Returning data about the last 100 public GitHub repositories

````console
$ curl -XPOST http://localhost:5000/repos
````

### /stats Endpoint
Returning langagues statistics about the last 100 public GitHub repositories

````console
$ curl -XPOST http://localhost:5000/stats
````

## Advanced usages
4 parameters are available to limit the search result on endpoints:

- Owner: Filter repositories by owner account
- Language: Filter repositories by language used
- Repository: Filter the result on a single repository
- License: Filter results by License

JSON payload example
````json
{
  "Owner": "scalingo",
  "License": "mit",
  "Language": "Go",
  "Repository": "Scalingo/sclng-backend-test-v1"
}
````

Filters can be used together (AND logic) except for 'Owner' and 'Repository' filters (You can NOT used togethers)

The request payload of POST endpoints include a JSON object contains zero, one, or more key-value pairs and a MIME Type for JSON ('Content-Type: application/json')

### Example : Get last 100 public GitHub repositories of Scalingo 
````console
curl -XPOST http://localhost:5000/repos -H 'Content-Type: application/json' -d '{"Owner":"scalingo"}'

curl -XPOST http://localhost:5000/stats -H 'Content-Type: application/json' -d '{"Owner":"scalingo"}'
````

## Some examples

### Get last "Go" public repositories on MIT License

````console
 curl -XPOST http://localhost:5000/repos -H 'Content-Type: application/json' -d '{"License": "mit", "Language": "Go"}'
````

````json
{
  "total_count": 186216,
  "Items": [
    {
      "ID": 468424882,
      "Name": "twipinion",
      "full_name": "jonathancowling/twipinion",
      "git_url": "git://github.com/jonathancowling/twipinion.git",
      "html_url": "https://github.com/jonathancowling/twipinion",
      "languages_url": "https://api.github.com/repos/jonathancowling/twipinion/languages",
      "created_at": "2022-03-10T16:30:52Z",
      "pushed_at": "2022-04-06T10:26:25Z",
      "updated_at": "2022-03-24T18:18:04Z",
      "language": "Go",
      "Languages": {
        "Go": 14407,
        "Java": 8386,
        "Shell": 3991
      }
    },
    // Etc...
    {
      "ID": 239971699,
      "Name": "erp",
      "full_name": "re-star-ru/erp",
      "git_url": "git://github.com/re-star-ru/erp.git",
      "html_url": "https://github.com/re-star-ru/erp",
      "languages_url": "https://api.github.com/repos/re-star-ru/erp/languages",
      "created_at": "2020-02-12T09:21:19Z",
      "pushed_at": "2022-04-06T10:25:54Z",
      "updated_at": "2021-12-21T10:13:12Z",
      "language": "Go",
      "Languages": {
        "CSS": 620,
        "Dockerfile": 1720,
        "Go": 129764,
        "HCL": 1614,
        "HTML": 26377,
        "JavaScript": 31176,
        "Makefile": 430,
        "SCSS": 1616,
        "Vue": 92049
      }
    }
  ]
}
````

### Get last languages used with "Python"
````console
 curl -XPOST http://localhost:5000/stats -H 'Content-Type: application/json' -d '{"Language": "Python"}'
````

````json
{
  "Languages": {
    "Batchfile": 1030,
    "C": 654157,
    "C++": 144721,
    "CMake": 27191,
    "CSS": 13717,
    "Dockerfile": 7945,
    "Go": 45825,
    "HTML": 48071,
    "JavaScript": 244186,
    "Jupyter Notebook": 65657,
    "MATLAB": 1630,
    "Makefile": 3488,
    "Mako": 494,
    "Procfile": 42,
    "Python": 13487434,
    "R": 15849,
    "SCSS": 19898,
    "Shell": 77583,
    "Tcl": 1309426,
    "TeX": 2481,
    "TypeScript": 4796,
    "Vue": 55687
  }
}
````