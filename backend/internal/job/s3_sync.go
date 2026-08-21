package job

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SyncToS3 is a background worker that runs after a job completes.
// It groups records by source_url, streams them to S3, and deletes the ephemeral DB rows.
//
//nolint:staticcheck
func (s *PipelineService) SyncToS3(jobID string) {
	if s.s3Client == nil {
		s.log.Warn("S3 client is not configured, skipping S3 Sync", "job_id", jobID)
		return
	}

	ctx := context.Background()
	s.log.Info("Starting S3 Sync for completed job", "job_id", jobID)

	sources, err := s.repo.GetDistinctSources(ctx, jobID)
	if err != nil {
		s.log.Error("S3 Sync failed to get sources", "job_id", jobID, "error", err)
		return
	}

	uploader := manager.NewUploader(s.s3Client.Client)

	for i, sourceURL := range sources {
		// Determine file extension and format based on source URL roughly
		format := "json"
		if strings.Contains(strings.ToLower(sourceURL), ".csv") {
			format = "csv"
		}

		key := fmt.Sprintf("job_%s_source_%d.%s", jobID, i+1, format)

		err := s.uploadSourceToS3(ctx, uploader, jobID, sourceURL, key, format)
		if err != nil {
			s.log.Error("Failed to upload source to S3", "job_id", jobID, "source", sourceURL, "error", err)
			continue
		}
		s.log.Info("Successfully uploaded source to S3", "job_id", jobID, "key", key)
	}

	// Delete from Ephemeral buffer
	if err := s.repo.DeleteExportedRecords(ctx, jobID); err != nil {
		s.log.Error("Failed to delete ephemeral records after S3 sync", "job_id", jobID, "error", err)
	} else {
		s.log.Info("Successfully cleaned up ephemeral records", "job_id", jobID)
	}
}

//nolint:staticcheck
func (s *PipelineService) uploadSourceToS3(ctx context.Context, uploader *manager.Uploader, jobID, sourceURL, key, format string) error {
	rows, err := s.repo.GetExportedRecordsBySource(ctx, jobID, sourceURL)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Use io.Pipe to stream DB rows directly into S3 upload without buffering in memory
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		if format == "csv" {
			csvWriter := csv.NewWriter(pw)
			defer csvWriter.Flush()
			headersWritten := false
			var headers []string

			for rows.Next() {
				var dataBytes []byte
				if err := rows.Scan(&dataBytes); err != nil {
					continue
				}
				var record map[string]any
				if err := json.Unmarshal(dataBytes, &record); err != nil {
					continue
				}

				if !headersWritten {
					for k := range record {
						headers = append(headers, k)
					}
					_ = csvWriter.Write(headers)
					headersWritten = true
				}

				var rowVals []string
				for _, h := range headers {
					val := record[h]
					if val == nil {
						rowVals = append(rowVals, "")
					} else {
						rowVals = append(rowVals, fmt.Sprintf("%v", val))
					}
				}
				_ = csvWriter.Write(rowVals)
			}
		} else {
			// JSON format
			_, _ = pw.Write([]byte("[\n"))
			first := true
			for rows.Next() {
				var dataBytes []byte
				if err := rows.Scan(&dataBytes); err != nil {
					continue
				}
				if !first {
					_, _ = pw.Write([]byte(",\n"))
				}
				_, _ = pw.Write(dataBytes)
				first = false
			}
			_, _ = pw.Write([]byte("\n]"))
		}
	}()

	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.s3Client.Bucket),
		Key:    aws.String(key),
		Body:   pr,
	})

	return err
}

func (s *PipelineService) GetExportURLs(ctx context.Context, jobID string, format string) ([]string, error) {
	if s.s3Client == nil {
		return nil, fmt.Errorf("S3 client not configured")
	}

	prefix := fmt.Sprintf("job_%s_", jobID)

	paginator := s3.NewListObjectsV2Paginator(s.s3Client.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.s3Client.Bucket),
		Prefix: aws.String(prefix),
	})

	var urls []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			if strings.HasSuffix(key, "."+format) {
				req, err := s.s3Client.Presign.PresignGetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(s.s3Client.Bucket),
					Key:    aws.String(key),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to presign URL for %s: %w", key, err)
				}
				// Fix the host for the host machine when running in Docker
				urlStr := strings.Replace(req.URL, "http://localstack:4566", "http://localhost:4566", 1)
				urls = append(urls, urlStr)
			}
		}
	}

	return urls, nil
}
