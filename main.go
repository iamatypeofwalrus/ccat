package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
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
	appUsage     = "cloud cat\n\n Stream objects from S3 to STDOUT"
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
	// * any object that ends in "/" assumes you want to stream the whole folder to STDOUT
	// * add a bytes flag for max number of bytes to scan
	// * add a max number of objects flag. bytes or max which ever comes first
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "profile, p",
			Usage: "AWS credentials profile",
		},
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "show this help message",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "log verbosely to stderr",
		},
	}

	app.Action = do

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "there was an error trying to stream objects: %v\n", err)
		os.Exit(1)
	}
}

func do(c *cli.Context) error {
	if c.Bool("help") || len(c.Args()) == 0 {
		cli.ShowAppHelp(c)
		return nil
	}

	if c.Bool("verbose") {
		// Logging to stderr since we're streaming the object to STDOUT
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	profile := c.String("profile")

	err := validateCredentials(profile)
	if err != nil {
		return fmt.Errorf("could not find any valid credentials")
	}

	ctx := context.Background()
	return streamObjectsFromS3(ctx, profile, c.Args())
}

func streamObjectsFromS3(ctx context.Context, profile string, objects []string) error {
	for _, obj := range objects {
		bucket, key, err := parseS3ObjectString(obj)
		if err != nil {
			return fmt.Errorf("could not parse %v into bucket and key: %v", obj, err)
		}

		log.Println("bucket:", bucket)
		log.Println("key:", key)

		err = streamObjectFromS3(ctx, profile, bucket, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func streamObjectFromS3(ctx context.Context, profile string, bucket string, key string) error {
	// figure out which region the bucket is in
	log.Println("getting bucket region")
	sess := session.Must(
		session.NewSessionWithOptions(
			session.Options{
				Profile: profile,
			},
		),
	)

	// TODO: consider caching the region for a bucket in memory
	// TODO: the HTTP URLS have the region built in and we could pass the hint down here
	region, err := s3manager.GetBucketRegion(ctx, sess, bucket, "us-east-1")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			fmt.Fprintf(os.Stderr, "unable to find bucket %s's region not found\n", bucket)
		}
		return err
	}

	// TODO: consider caching sessions
	log.Println("creating session for", region)
	streamSess := session.Must(
		session.NewSession(
			&aws.Config{Region: aws.String(region)},
		),
	)

	svc := s3manager.NewDownloader(streamSess, func(d *s3manager.Downloader) {
		// Force the downloader to stream the object sequentially
		d.Concurrency = 1
	})
	log.Printf("streaming %s/%s\n", bucket, key)
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
func parseS3ObjectString(obj string) (string, string, error) {
	if strings.HasPrefix(obj, prefixS3) {
		return parseS3Key(obj)
	}
	return parseHTTPKey(obj)
}

func parseS3Key(obj string) (string, string, error) {
	str := strings.Replace(obj, prefixS3, "", 1)
	split := strings.SplitN(str, "/", 2)
	if len(split) != 2 {
		return "", "", fmt.Errorf("could not parse s3 key: %s", obj)
	}

	return split[0], split[1], nil
}

func parseHTTPKey(obj string) (string, string, error) {
	str := strings.Replace(obj, prefixHTTPS, "", 1)
	escaped, err := url.PathUnescape(str)
	if err != nil {
		return "", "", fmt.Errorf("could not unescape url: %s", obj)
	}

	split := strings.SplitN(escaped, "/", 3)
	if len(split) != 3 {
		return "", "", fmt.Errorf("could not parse http s3 key: %s", obj)
	}

	return split[1], split[2], nil
}

func validateCredentials(profile string) error {
	sess := session.Must(
		session.NewSessionWithOptions(
			session.Options{
				Profile: profile,
			},
		),
	)
	creds := sess.Config.Credentials
	_, err := creds.Get()
	return err
}
