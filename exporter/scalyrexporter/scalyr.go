// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scalyrexporter

import (
	"context"
	"net/http"
	"encoding/json"
	"fmt"
        "bytes"
        "strconv"

	"go.uber.org/zap"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
        _ "google.golang.org/protobuf/encoding/protojson"
	"go.opentelemetry.io/collector/translator/trace/zipkin"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
)

func Serialize(spans []*zipkinmodel.SpanModel) ([]byte, error) {
	events := make([]scalyrEvent, len(spans))
	for i, span := range spans {
		events[i].TS=strconv.FormatInt(span.Timestamp.UnixNano(),10)
		events[i].Attrs=span
	}
	return json.Marshal(events)
}

type scalyrEvent struct {
  Thread string `json:"thread"`
  TS string     `json:"ts"`
  Sev int       `json:"sev"`
  Attrs *zipkinmodel.SpanModel  `json:"attrs"`

}


type scalyrExporter struct {
	client     *http.Client
        logger     *zap.Logger
	serializer (func([]*zipkinmodel.SpanModel) ([]byte, error))
	url        string
	apikey        string
}

// newTraceExporter creates an zipkin trace exporter.
func newTraceExporter(config *Config, params component.ExporterCreateParams) (component.TraceExporter, error) {
	se, err := createScalyrExporter(config, params)
	if err != nil {
		return nil, err
	}
	sexp, err := exporterhelper.NewTraceExporter(config, se.pushTraceData)
	if err != nil {
		return nil, err
	}

	return sexp, nil
}

func createScalyrExporter(cfg *Config, params component.ExporterCreateParams) (*scalyrExporter, error) {
	client, err := cfg.HTTPClientSettings.ToClient()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/addEvents", cfg.Endpoint)
	params.Logger.Info("url", zap.Any("url", url))
	ze := &scalyrExporter{
		client:             client,
		url: url,
		apikey: cfg.APIKey,
		logger:		params.Logger,
                serializer:     Serialize,
	}

	return ze, nil
}
func (se *scalyrExporter) pushTraceData(ctx context.Context, td pdata.Traces) (int, error) {
      tbatch, err := zipkin.InternalTracesToZipkinSpans(td)
        if err != nil {
                return td.SpanCount(), consumererror.Permanent(fmt.Errorf("failed to push trace data via Zipkin exporter: %w", err))
        }

        events, err := se.serializer(tbatch)
        if err != nil {
                return td.SpanCount(), consumererror.Permanent(fmt.Errorf("failed to push trace data via Zipkin exporter: %w", err))
        }
	request:=`{ 
  "token": "%s",
  "session": "%s",
  "events": %s
}`
	body := fmt.Sprintf(request, se.apikey, "meh", events)
	se.logger.Info("Request", zap.Any("body", body))
        req, err := http.NewRequestWithContext(ctx, "POST", se.url, bytes.NewReader([]byte(body)))
        if err != nil {
                return td.SpanCount(), fmt.Errorf("failed to push trace data via Zipkin exporter: %w", err)
        }
        req.Header.Set("Content-Type", "text/plain")

        resp, err := se.client.Do(req)
        if err != nil {
                return td.SpanCount(), fmt.Errorf("failed to push trace data via Zipkin exporter: %w", err)
        }
        _ = resp.Body.Close()
        if resp.StatusCode < 200 || resp.StatusCode > 299 {
                return td.SpanCount(), fmt.Errorf("failed the request with status code %d", resp.StatusCode)
        }
        return 0, nil
}
