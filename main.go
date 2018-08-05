package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/urfave/cli"
)

const (
	appName    = "ccat"
	appUsage   = "Cloud cat\n\n A simple CLI that streams objects from S3 to STDOUT"
	appVersion = "0.1.0"
)

var (
	out = stdout{}
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = appUsage
	app.Version = appVersion
	app.HideHelp = true
	app.HideVersion = true

	// TODO:
	// * Handle a s3:// path
	// * handle "*" prefix and stream multiple objects sequentially
	// * add a bytes flag for max number of bytes to scan
	// * if S3 supports an offset flag add a CLI option
	// * add a max number of objects flag. bytes or max which ever comes first
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "bucket, b",
			Usage: "Bucket `NAME`",
		},
		cli.StringFlag{
			Name:  "key, k",
			Usage: "Key `NAME`",
		},
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

	bucket := c.String("bucket")
	if bucket == "" {
		return fmt.Errorf("--bucket, -b flag is required")
	}

	key := c.String("key")
	if bucket == "" {
		return fmt.Errorf("--key, -k flag is required")
	}

	return streamFromS3(context.Background(), bucket, key)
}

func streamFromS3(ctx context.Context, bucket string, key string) error {
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
	read, err := svc.DownloadWithContext(ctx, out, input)
	fmt.Println("read this many bytes:", read)

	return err
}
