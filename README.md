[![Build Status](https://codebuild.us-west-2.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoiQ09OKzJybUl1b21jMnBFbDczVGtYQmlmalFkczl6MlphTmNlM1BFQS85OGRFMDdXOWNYT2hibGtyUjhOQXBBVnZJN3ZUdXo5RjdtdFVRdlY2UnRHMEZ3PSIsIml2UGFyYW1ldGVyU3BlYyI6ImRTd25EZ0lOWnNBQVVMVEciLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=master)](https://codebuild.us-west-2.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoiQ09OKzJybUl1b21jMnBFbDczVGtYQmlmalFkczl6MlphTmNlM1BFQS85OGRFMDdXOWNYT2hibGtyUjhOQXBBVnZJN3ZUdXo5RjdtdFVRdlY2UnRHMEZ3PSIsIml2UGFyYW1ldGVyU3BlYyI6ImRTd25EZ0lOWnNBQVVMVEciLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=master)

# cloud cat
Inspired by the [`cat` Unix command](https://en.wikipedia.org/wiki/Cat_(Unix)), cloud cat (`ccat`) can stream one or more objects from the Amazon Web Services (AWS) Simple Storage Service (S3) and print the results to STDOUT.

Because cloud cat streams the objects, not downloads, you can work with large objects in S3 as if they're local without having to worry about the amount of space you have on your local disk.

## CLI Usage
```
NAME:
   ccat - cloud cat

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
