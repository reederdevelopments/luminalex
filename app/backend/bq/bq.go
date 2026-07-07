package bq

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type RatingEvent struct {
	UserID      string    `bigquery:"USER_ID"`
	ReportID    string    `bigquery:"REPORT_ID"`
	ReportTitle string    `bigquery:"TITLE"`
	CountryCode string    `bigquery:"COUNTRY"`
	Rating      string    `bigquery:"RATING"`
	Feedback    string    `bigquery:"FEEDBACK"`
	Timestamp   time.Time `bigquery:"TIMESTAMP"`
}

type TopReport struct {
	ReportTitle string `bigquery:"REPORT"`
	ReportID    string `json:"-"`
	ReportCC    string `json:"-"`
}

type Service struct {
	client  *bigquery.Client
	l       *log.Logger
	dataset string
	table   string
}

func NewService(ctx context.Context, l *log.Logger, projectID, dataset, table string) (*Service, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %w", err)
	}

	return &Service{
		client:  client,
		l:       l,
		dataset: dataset,
		table:   table,
	}, nil
}

func (s *Service) GetUserTopReports(ctx context.Context, userEmail string) ([]TopReport, error) {
	queryStr := "SELECT REPORT FROM `df-frontend.UJUZI.TOP_5` WHERE USER = @email"
	q := s.client.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "email", Value: userEmail},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("query.Read: %w", err)
	}

	var reports []TopReport
	for {
		var r TopReport
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterator.Next: %w", err)
		}
		reports = append(reports, r)
	}
	return reports, nil
}

func (s *Service) Close() {
	s.client.Close()
}
