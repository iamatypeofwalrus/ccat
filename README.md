# ccat
Cloud Cat: Cat objects from S3 to STDOUT

## CLI Usage
```
NAME:
   ccat - Cloud cat

 A simple CLI that streams objects from S3 to STDOUT

USAGE:
   ccat s3://your-bucket/your-key https://s3-us-west-2.amazonaws.com/your-bucket/your-other-key

GLOBAL OPTIONS:
   --help, -h  show this help message
```
## Examples
### Print an object to STDOUT
```
ccat s3://your-bucket/your-key.json
```

### Print an object with an S3 URL to STDOUT
```
ccat https://s3-us-west-2.amazonaws.com/your-bucket/your-key.json
```

### Print multiple objects to STDOUT
```
ccat s3://your-bucket/your-key.json s3://your-bucket/your-other-key.json
```

### Print a gzipped object to STDOUT
```
ccat s3://your-bucket/your-key.json.gz | zcat
```