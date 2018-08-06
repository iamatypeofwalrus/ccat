package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/urfave/cli"
)

const (
	appName      = "ccat"
	appUsage     = "Cloud cat\n\n A simple CLI that streams objects from S3 to STDOUT"
	appVersion   = "0.1.0"
	appUsageText = "ccat s3://your-bucket/your-key https://s3-us-west-2.amazonaws.com/your-bucket/your-other-key"

	prefixS3    = "s3://"
	prefixHTTPS = "https://"
)

var (
	// Stdout is a drop in replacement for os.Stdout when we have to use the io.WriterAt interface
	Stdout = stdout{}
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = appUsage
	app.UsageText = appUsageText
	app.Version = appVersion
	app.HideHelp = true
	app.HideVersion = true

	// TODO:
	// * add bytes range
	// * handle "*" prefix and stream multiple objects sequentially
	// * add a bytes flag for max number of bytes to scan
	// * add a max number of objects flag. bytes or max which ever comes first
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "show this help message",
		},
	}

	app.Action = do

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func do(c *cli.Context) error {
	if c.Bool("help") {
		cli.ShowAppHelp(c)
		return nil
	}

	ctx := context.Background()
	streamObjectsFromS3(ctx, c.Args())

	return nil
}

func streamObjectsFromS3(ctx context.Context, objects []string) error {
	for _, obj := range objects {
		bucket, key := parseS3ObjectString(obj)
		if bucket == "" || key == "" {
			return fmt.Errorf("could not parse %v into bucket and key", obj)
		}
		err := streamObjectFromS3(ctx, bucket, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func streamObjectFromS3(ctx context.Context, bucket string, key string) error {
	// figure out which region the bucket is in
	sess := session.Must(session.NewSession())
	region, err := s3manager.GetBucketRegion(ctx, sess, bucket, "us-east-1")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			fmt.Fprintf(os.Stderr, "unable to find bucket %s's region not found\n", bucket)
		}
		return err
	}

	streamSess := session.Must(
		session.NewSession(
			&aws.Config{Region: aws.String(region)},
		),
	)

	svc := s3manager.NewDownloader(streamSess, func(d *s3manager.Downloader) {
		// Force the downloader to stream the object sequentially
		d.Concurrency = 1
	})
	return stream(bucket, key, svc)
}

func stream(bucket string, key string, svc *s3manager.Downloader) error {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	ctx := context.Background()
	_, err := svc.DownloadWithContext(ctx, Stdout, input)

	return err
}

// parseS3ObjectString can parse definitions of the form s3://bucket/key
// and https://s3-us-west-2.amazonaws.com/bucket/key
// into bucket and key.
func parseS3ObjectString(obj string) (string, string) {
	if strings.HasPrefix(obj, prefixS3) {
		return parseS3Key(obj)
	} else if strings.HasPrefix(obj, prefixHTTPS) {
		return parseHTTPKey(obj)
	}

	return "", ""
}

func parseS3Key(obj string) (string, string) {
	str := strings.Replace(obj, prefixS3, "", 1)
	split := strings.SplitN(str, "/", 2)
	if len(split) != 2 {
		return "", ""
	}

	return split[0], split[1]
}

func parseHTTPKey(obj string) (string, string) {
	str := strings.Replace(obj, prefixHTTPS, "", 1)
	split := strings.SplitN(str, "/", 3)
	if len(split) != 3 {
		return "", ""
	}
	return split[1], split[2]
}
